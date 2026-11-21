package queue

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newTestStore(t *testing.T) *PostgresStore {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.ExecContext(t.Context(), `TRUNCATE jobs RESTART IDENTITY`)
	if err != nil {
		t.Fatalf("truncate jobs: %v", err)
	}
	return NewPostgresStore(db)
}

func TestPostgresStore_Insert(t *testing.T) {
	s := newTestStore(t)
	before := time.Now().Unix()

	id, err := s.Insert(t.Context(), "send_email", []byte(`{"to": "ke.com"}`))
	if err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	var (
		jobType, status, to string
		attempts            int
		scheduledAt         int64
	)
	err = s.db.QueryRowContext(t.Context(), `SELECT type, status, attempts, scheduled_at, payload->>'to'
	FROM jobs WHERE id = $1`, id).Scan(&jobType, &status, &attempts, &scheduledAt, &to)
	if err != nil {
		t.Fatalf("read job %d: %v", id, err)
	}

	if jobType != "send_email" {
		t.Errorf("type = %q, want %q", jobType, "send_email")
	}
	if attempts != 0 {
		t.Errorf("attempts = %d, want 0", attempts)
	}
	if to != "ke.com" {
		t.Errorf(`payload["to"] = %q, want %q`, to, "ke.com")
	}
	if scheduledAt < before {
		t.Errorf("scheduled_at = %d, want >= %d", scheduledAt, before)
	}
}

func TestPostgresStore_Claim(t *testing.T) {
	s := newTestStore(t)
	first, err := s.Insert(t.Context(), "send_email", []byte(`{"to":"k@e.com"}`))
	if err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	job, err := s.Claim(t.Context())
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}

	if job.ID != first {
		t.Errorf("job.ID = %d, want %d", job.ID, first)
	}
	if job.JobType != "send_email" {
		t.Errorf("job.JobType = %q, want %q", job.JobType, "send_email")
	}
	if got := job.Payload["to"]; got != "k@e.com" {
		t.Errorf(`job.Payload["to"] = %v, want %q`, got, "k@e.com")
	}
	if job.Status != "running" {
		t.Errorf("job.Status = %q, want %q", job.Status, "running")
	}
	if job.Attempts != 1 {
		t.Errorf("job.Attempts = %d, want 1", job.Attempts)
	}
}
