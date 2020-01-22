package streamer

import (
	"context"
	"encoding/json"
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
}

// NewRedisStreamer returns an instance of redisStreamer.
func NewRedisStreamer(client *redis.Client) (*redisStreamer, error) {
	file, err := os.Open("/scripts/ackAndAdd.lua")
	if err != nil {
		return nil, err
	}
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
func (s *redisStreamer) ReadGroup(ctx context.Context, args *StreamArgs, msgChan chan Message, exitChan chan struct{}, wg *sync.WaitGroup) error {
	// create consumer group if not done yet
	for _, stream := range args.Streams {
		_, err := s.rdb.XGroupCreateMkStream(stream, args.Group, "$").Result()
		if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
			return err
		}
	}

	go func() {
		lastID := "0-0"
		checkHistory := true

		for {
			if !checkHistory {
				lastID = ">"
			}

			streams := make([]string, len(args.Streams))
			for _, stream := range args.Streams {
				streams = append(streams, stream, lastID)
			}
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
			items := s.rdb.XReadGroup(xargs)
			if items == nil {
				// Timeout, check if it's time to exit
				if s.shouldExit(ctx, exitChan) {
					log.Debugf("Time to exit, stop reading streams [%s]", args.Streams)
					return
				}
				continue
			}

			// check if we are up to date
			if len(items.Val()) == 0 || len(items.Val()[0].Messages) == 0 {
				if checkHistory {
					log.Debug("Done reading stream history.")
				}
				checkHistory = false
				continue
			}

			respStreams := items.Val()
			for _, stream := range respStreams {
				msgs := len(stream.Messages)
				wg.Add(msgs)

				plural := ""
				if msgs > 1 {
					plural = "s"
				}
				log.Debugf("Consumer [%s] recived %d message%s", args.Consumer, msgs, plural)

				for _, rawMsg := range stream.Messages {
					log.Debugf("Consumer [%s] reading message [%s]", args.Consumer, rawMsg.ID)

					lastID = rawMsg.ID

					// parse message
					strMsg, ok := rawMsg.Values["message"].(string)
					if !ok {
						// error parsing stream message
						continue
					}

					var msg Message

					err := json.Unmarshal([]byte(strMsg), &msg)
					if err != nil {
						// malformed message, skip it.
						continue
					}
					msg.ID = rawMsg.ID
					msg.Stream = stream.Stream

					msgChan <- msg
				}
			}

			// avoid back-pressure
			wg.Wait()
		}
	}()
	return nil
}

func (s *redisStreamer) shouldExit(ctx context.Context, exitCh chan struct{}) bool {
	// TODO: is exitChan really necessary?
	select {
	case _, ok := <-exitCh:
		return !ok
	case <-ctx.Done():
		return true
	default:
	}
	return false
}
