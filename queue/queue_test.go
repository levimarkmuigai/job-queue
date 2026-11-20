package queue

import (
	"context"
	"errors"
	"testing"
)

type fakeStore struct {
	gotPayload []byte
	gotType    string
	id         int64
	err        error
	calls      int
}

func (f *fakeStore) Insert(ctx context.Context, jobType string, payload []byte) (int64, error) {
	f.calls++
	f.gotPayload = payload
	f.gotType = jobType
	return f.id, f.err
}

func TestEnqueue(t *testing.T) {
	store := &fakeStore{id: 254}
	q := NewQueue(store)

	id, err := q.Enqueue(t.Context(), "send_email", map[string]any{"to": "ke.com"})
	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if id != 254 {
		t.Fatalf("id = %d, want 254", id)
	}
	if store.gotType != "send_email" {
		t.Errorf("stored type = %q, want %q", store.gotType, "send_email")
	}
	if got, want := string(store.gotPayload), `{"to":"ke.com"}`; got != want {
		t.Errorf("store payload = %s, want %s", got, want)
	}
}

var errBoom = errors.New("boom")

func TestEnqueue_StoreError(t *testing.T) {
	store := &fakeStore{id: 254, err: errBoom}
	q := NewQueue(store)

	id, err := q.Enqueue(t.Context(), "send_email", map[string]any{"a": 1})

	if !errors.Is(err, errBoom) {
		t.Errorf("error = %v, want it to wrap %v", err, errBoom)
	}
	if id != 254 {
		t.Errorf("id = %d, want 254 on error", id)
	}
}

func TestEnqueue_CancelledContext(t *testing.T) {
	store := &fakeStore{id: 254}
	q := NewQueue(store)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := q.Enqueue(ctx, "send_email", map[string]any{"a": 1})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
	if store.calls != 0 {
		t.Errorf("store called %d times, want 0", store.calls)
	}
}

func TestEnqueue_UnencodablePayload(t *testing.T) {
	store := &fakeStore{id: 254}
	q := NewQueue(store)

	_, err := q.Enqueue(t.Context(), "send_email", map[string]any{"c": make(chan int)})

	if err == nil {
		t.Fatal("Enqueue() error = nil, want error")
	}
	if store.calls != 0 {
		t.Errorf("store called %d times, want 0", store.calls)
	}
}

func TestEnqueue_NilPayload(t *testing.T) {
	store := &fakeStore{id: 254}
	q := NewQueue(store)

	_, err := q.Enqueue(t.Context(), "send_email", nil)

	if !errors.Is(err, ErrNilPayload) {
		t.Errorf("error = %v, want ErrNilPayload", err)
	}

	if store.calls != 0 {
		t.Errorf("store called %d times, want 0", store.calls)
	}
}
