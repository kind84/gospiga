package streamer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
)

type redisStreamer struct {
	rdb *redis.Client
}

type StreamArgs struct {
	Stream   string
	Group    string
	Consumer string
}

func NewRedisStreamer(client *redis.Client) *redisStreamer {
	return &redisStreamer{client}
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
	_, err = s.rdb.XAdd(xargs).Result()
	return err
}

func (s *redisStreamer) ReadGroup(ctx context.Context, args *StreamArgs, msgChan chan interface{}, exitChan chan struct{}) {
	go func() {
		// create consumer group if not done yet
		s.rdb.XGroupCreateMkStream(args.Stream, args.Group, "$").Result()

		lastID := "0-0"
		checkHistory := true

		for {
			if !checkHistory {
				lastID = ">"
			}

			xargs := &redis.XReadGroupArgs{
				Group:    args.Group,
				Consumer: args.Consumer,
				// List of streams and ids.
				Streams: []string{args.Stream, lastID},
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
					// log.Debugf("Stop reading stream [%s]", args.Stream)
					return
				}
				continue
			}

			// check if we are up to date
			if len(items.Val()) == 0 || len(items.Val()[0].Messages) == 0 {
				if checkHistory {
					// log.Debugf("Done reading stream history.")
				}
				checkHistory = false
				continue
			}

			jobStream := items.Val()[0]
			// msgs := len(jobStream.Messages)
			// plural := ""
			// if msgs > 1 {
			// 	plural = "s"
			// }
			// log.Debugf("Consumer [%s] recived %d message%s", args.Consumer, msgs, plural)

			for _, rawMsg := range jobStream.Messages {
				// log.Debugf("Consumer [%s] reading message [%s]", args.Consumer, rawMsg.ID)

				lastID = rawMsg.ID

				// parse message
				strMsg, ok := rawMsg.Values["job"].(string)
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

				msgChan <- msg
			}
		}
	}()
}

func (s *redisStreamer) shouldExit(ctx context.Context, exitCh chan struct{}) bool {
	select {
	case _, ok := <-exitCh:
		return !ok
	case <-ctx.Done():
		return true
	default:
	}
	return false
}
