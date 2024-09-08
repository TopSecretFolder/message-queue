package redisimpl

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	clnt *redis.Client
	ctx  context.Context
	ttl  time.Duration
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

	err = c.clnt.Expire(c.ctx, key, c.ttl).Err()
	if err != nil {
		return nil, err
	}

	return []byte(result[1]), nil
}

// Enqueue implements queue.QueueProvider.
func (c Client) Enqueue(key string, data []byte) error {
	err := c.clnt.RPush(context.Background(), key, data).Err()
	if err != nil {
		return err
	}

	err = c.clnt.Expire(c.ctx, key, c.ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func NewClient(ctx context.Context, ip string, ttl time.Duration) Client {
	return Client{
		ctx:  ctx,
		clnt: redis.NewClient(&redis.Options{Addr: ip}),
		ttl:  ttl,
	}
}

func IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}
