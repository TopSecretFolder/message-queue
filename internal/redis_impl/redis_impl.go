package redisimpl

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	clnt *redis.Client
}

// IsEmptyErr implements queue.QueueProvider.
func (c Client) IsEmptyErr(err error) bool {
	return errors.Is(err, redis.Nil)
}

// Dequeue implements queue.QueueProvider.
func (c Client) Dequeue(key string) ([]byte, error) {
	result, err := c.clnt.BLPop(context.Background(), time.Second, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(result[1]), nil
}

// Enqueue implements queue.QueueProvider.
func (c Client) Enqueue(key string, data []byte) error {
	return c.clnt.RPush(context.Background(), key, data).Err()
}

func NewClient(ip string) Client {
	return Client{
		clnt: redis.NewClient(&redis.Options{Addr: ip}),
	}
}

func IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}
