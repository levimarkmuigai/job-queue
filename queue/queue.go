package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrNilPayload = errors.New("queue: payload is nil")

type Store interface {
	Insert(ctx context.Context, jobType string, payload []byte) (int64, error)
}

type Queue struct {
	store Store
}

func NewQueue(s Store) *Queue {
	return &Queue{store: s}
}

func (q *Queue) Enqueue(ctx context.Context, jobType string, payload map[string]any) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if payload == nil {
		return 0, ErrNilPayload
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}
	id, err := q.store.Insert(ctx, jobType, b)
	if err != nil {
		return id, fmt.Errorf("insert job: %w", err)
	}

	return id, nil
}
