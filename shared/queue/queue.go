package queue

import (
	"encoding/json"
	"fmt"
)

type QueueProvider interface {
	Enqueue(key string, data []byte) error
	Dequeue(key string) ([]byte, error)
	IsEmptyErr(err error) bool
}

func Instance[T any](key Key, provider QueueProvider) Queue[T] {
	return Queue[T]{
		key:      key,
		provider: provider,
	}
}

func GetKey(id string) Key {
	return Key(fmt.Sprintf("%s:%s", "queue", id))
}

type Key string

type Queue[T any] struct {
	key      Key
	provider QueueProvider
}

func (q Queue[T]) Enqueue(obj T) error {
	serialized, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("could not serialize data before queueing: %w", err)
	}
	err = q.provider.Enqueue(string(q.key), serialized)
	if err != nil {
		return fmt.Errorf("could not enqueue data: %w", err)
	}
	return nil
}

func (q Queue[T]) Dequeue() (T, error) {
	var t T

	b, err := q.provider.Dequeue(string(q.key))
	if err != nil {
		return t, fmt.Errorf("could not dequeue data: %w", err)
	}

	err = json.Unmarshal(b, &t)
	if err != nil {
		return t, fmt.Errorf("could not unmarshal queue item: %w", err)
	}

	return t, nil
}

func (q Queue[T]) IsEmptyErr(err error) bool {
	return q.provider.IsEmptyErr(err)
}
