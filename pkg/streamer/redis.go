package streamer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"gospiga/pkg/log"
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
}

// NewRedisStreamer returns an instance of redisStreamer.
func NewRedisStreamer(client *redis.Client) (*redisStreamer, error) {
	file, err := os.Open("/scripts/lua/ackAndAdd.lua")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	script, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	ackAndAddLua, err = client.ScriptLoad(context.Background(), string(script)).Result()
	if err != nil {
		return nil, err
	}
	return &redisStreamer{client}, nil
}

func (s *redisStreamer) Ack(ctx context.Context, stream, group string, ids ...string) error {
	_, err := s.rdb.XAck(ctx, stream, group, ids...).Result()
	return err
}

func (s *redisStreamer) Add(ctx context.Context, stream string, msg *Message) error {
	jmsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	xargs := &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{"message": string(jmsg)},
	}
	_, err = s.rdb.XAdd(ctx, xargs).Result()
	return err
}

// AckAndAdd atomically acknowledges a given message ID from a stream and
// sends the given message to another stream.
func (s *redisStreamer) AckAndAdd(ctx context.Context, fromStream, toStream, group, id string, msg *Message) error {
	jmsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// run pre-loaded script
	_, err = s.rdb.EvalSha(
		ctx,
		ackAndAddLua,
		[]string{fromStream, toStream}, // KEYS
		[]string{group, id, "message", string(jmsg)}, // ARGV
	).Result()

	return err
}

// ReadGroup reads messages on the given stream and sends them over a channel.
func (s *redisStreamer) ReadGroup(ctx context.Context, wg *sync.WaitGroup, args *StreamArgs) error {
	// create consumer group if not done yet
	for _, stream := range args.Streams {
		_, err := s.rdb.XGroupCreateMkStream(ctx, stream, args.Group, "0-0").Result()
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

			// pre-emptive check
			select {
			case <-ctx.Done():
				return
			default:
			}

			res, err := s.rdb.XReadGroup(ctx, xargs).Result()
			if err != nil {
				if err != redis.Nil {
					log.Errorf("error reading streams %s: %s", args.Streams, err)
				}
				// timeout reached
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
					log.Debugf("Found %d pending messages on stream %q, resuming..", msgs, stream.Stream)
				}

				wg.Add(msgs)

				log.Debugf("Consumer %q recived %d message(s)", args.Consumer, msgs)

				for _, rawMsg := range stream.Messages {
					lastIDs[stream.Stream] = rawMsg.ID

					msg, err := parseMessage(rawMsg, stream.Stream)
					if err != nil {
						log.Errorf(err.Error())
						// clear malformed message
						s.Ack(ctx, stream.Stream, args.Group, rawMsg.ID)
						continue
					}

					select {
					case <-ctx.Done():
						return
					case args.Messages <- *msg:
					}
				}

				// avoid back-pressure
				wg.Wait()
			}

			if !gotMessage && checkHistory {
				log.Debugf("Done reading history on streams: %v", args.Streams)
				checkHistory = false
			}
		}
	}()
	return nil
}

func (s *redisStreamer) shouldExit(exitCh <-chan struct{}) bool {
	select {
	case <-exitCh:
		return true
	default:
		return false
	}
}

func parseMessage(rawMsg redis.XMessage, stream string) (*Message, error) {
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
	msg.Stream = stream

	return &msg, nil
}
