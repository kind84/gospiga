package streamer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v7"
	log "github.com/sirupsen/logrus"
)

var ackAndAddLua = ""

type redisStreamer struct {
	rdb *redis.Client
}

// StreamArgs required to deal with streams.
type StreamArgs struct {
	Streams  []string
	Group    string
	Consumer string
	Messages chan Message
	Exit     chan struct{}
	WG       *sync.WaitGroup
}

// NewRedisStreamer returns an instance of redisStreamer.
func NewRedisStreamer(client *redis.Client) (*redisStreamer, error) {
	file, err := os.Open("/scripts/ackAndAdd.lua")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	script, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	ackAndAddLua, err = client.ScriptLoad(string(script)).Result()
	if err != nil {
		return nil, err
	}
	return &redisStreamer{client}, nil
}

func (s *redisStreamer) Ack(stream, group string, ids ...string) error {
	_, err := s.rdb.XAck(stream, group, ids...).Result()
	return err
}

func (s *redisStreamer) Add(stream string, msg *Message) error {
	jmsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	xargs := &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{"message": string(jmsg)},
	}
	_, err = s.rdb.XAdd(xargs).Result()
	return err
}

// AckAndAdd atomically acknowledges a given message ID from a stream and
// sends the given message to another stream.
func (s *redisStreamer) AckAndAdd(fromStream, toStream, group, id string, msg *Message) error {
	jmsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// run pre-loaded script
	_, err = s.rdb.EvalSha(
		ackAndAddLua,
		[]string{fromStream, toStream},               // KEYS
		[]string{group, id, "message", string(jmsg)}, // ARGV
	).Result()

	return err
}

// ReadGroup reads messages on the given stream and sends them over a channel.
func (s *redisStreamer) ReadGroup(args *StreamArgs) error {
	// create consumer group if not done yet
	for _, stream := range args.Streams {
		_, err := s.rdb.XGroupCreateMkStream(stream, args.Group, "0-0").Result()
		if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
			return err
		}
	}

	go func() {
		checkHistory := true

		lastIDs := make(map[string]string, len(args.Streams))
		for _, stream := range args.Streams {
			lastIDs[stream] = "0-0"
		}
		for {
			if !checkHistory {
				for _, stream := range args.Streams {
					lastIDs[stream] = ">"
				}
			}

			streams := make([]string, 0, len(args.Streams)*2)
			ids := make([]string, 0, len(args.Streams))
			for _, id := range lastIDs {
				ids = append(ids, id)
			}
			streams = append(streams, args.Streams...)
			streams = append(streams, ids...)

			xargs := &redis.XReadGroupArgs{
				Group:    args.Group,
				Consumer: args.Consumer,
				// List of streams and ids.
				Streams: streams,
				// Max no. of elements per stream fo each call.
				Count: 10,
				Block: time.Millisecond * 2000,
				// NoAck   bool
			}

			// TODO: use WithContext ?
			res, err := s.rdb.XReadGroup(xargs).Result()
			if err != nil {
				if err != redis.Nil {
					log.Errorf("error reading streams %s: %s", args.Streams, err)
				}
				// Timeout, check if it's time to exit
				if s.shouldExit(args.Exit) {
					log.Debugf("stop reading streams %s", args.Streams)
					return
				}
				continue
			}

			// check if we are up to date
			if len(res) == 0 {
				if checkHistory {
					log.Debugf("done reading history on streams: %v", args.Streams)
					checkHistory = false
				}
				continue
			}

			gotMessage := false
			for _, stream := range res {
				msgs := len(stream.Messages)
				if msgs == 0 {
					continue
				}
				gotMessage = true
				if checkHistory {
					log.Debugf("found pending messages on stream %q, resuming..", stream.Stream)
				}

				args.WG.Add(msgs)

				plural := ""
				if msgs > 1 {
					plural = "s"
				}
				log.Debugf("Consumer %q recived %d message%s", args.Consumer, msgs, plural)

				for _, rawMsg := range stream.Messages {
					log.Debugf("Consumer %q reading message %q", args.Consumer, rawMsg.ID)

					lastIDs[stream.Stream] = rawMsg.ID

					msg, err := parseMessage(rawMsg)
					if err != nil {
						log.Errorf(err.Error())
						// clear malformed message
						s.Ack(stream.Stream, args.Group, rawMsg.ID)
						continue
					}

					args.Messages <- *msg
				}

				// avoid back-pressure
				args.WG.Wait()
			}

			if !gotMessage && checkHistory {
				log.Debugf("Done reading history on streams: %v", args.Streams)
				checkHistory = false
			}
		}
	}()
	return nil
}

func (s *redisStreamer) shouldExit(exitCh chan struct{}) bool {
	// TODO: is exitChan really necessary?
	select {
	case _, ok := <-exitCh:
		return !ok
	default:
	}
	return false
}

func parseMessage(rawMsg redis.XMessage) (*Message, error) {
	strMsg, ok := rawMsg.Values["message"].(string)
	if !ok {
		return nil, fmt.Errorf("cannot parse stream message %q", rawMsg.ID)
	}

	var msg Message
	err := json.Unmarshal([]byte(strMsg), &msg)
	if err != nil {
		return nil, fmt.Errorf("malformed stream message %q, cannot unmarshal to mq.Message", rawMsg.ID)
	}
	msg.ID = rawMsg.ID

	return &msg, nil
}
