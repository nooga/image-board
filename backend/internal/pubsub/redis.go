package pubsub

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const (
	ChannelFeed   = "imageboard:feed"
	ChannelPrefix = "imageboard:topic:"
)

type PubSub interface {
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string) (<-chan []byte, func())
	TopicChannel(topicID string) string
	Close() error
}

type RedisPubSub struct {
	client *redis.Client
}

func NewRedisPubSub(addr string) (*RedisPubSub, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisPubSub{client: client}, nil
}

func (r *RedisPubSub) Close() error {
	return r.client.Close()
}

func (r *RedisPubSub) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, data).Err()
}

func (r *RedisPubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, func()) {
	sub := r.client.Subscribe(ctx, channel)
	msgChan := make(chan []byte, 100)

	go func() {
		defer close(msgChan)
		ch := sub.Channel()
		for msg := range ch {
			select {
			case msgChan <- []byte(msg.Payload):
			case <-ctx.Done():
				return
			}
		}
	}()

	cleanup := func() {
		sub.Close()
	}

	return msgChan, cleanup
}

func (r *RedisPubSub) TopicChannel(topicID string) string {
	return ChannelPrefix + topicID
}

