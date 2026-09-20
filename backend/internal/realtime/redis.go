package realtime

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisClient(addr, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
}

func NewRedisPublisher(client *redis.Client) *RedisPublisher {
	return &RedisPublisher{client: client}
}

func (p *RedisPublisher) Publish(ctx context.Context, event Event) error {
	if event.SentAt.IsZero() {
		event.SentAt = time.Now().UTC()
	}
	payload, err := Encode(event)
	if err != nil {
		return fmt.Errorf("encode realtime event: %w", err)
	}
	if err := p.client.Publish(ctx, Channel(event.AuctionID), payload).Err(); err != nil {
		return fmt.Errorf("publish realtime event: %w", err)
	}
	return nil
}
