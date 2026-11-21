package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

var _ Store = (*PostgresStore)(nil)

type Job struct {
	ID              int64
	JobType         string
	Payload         map[string]any
	Status          string
	Attempts        int
	CreatedAtUnix   int64
	ScheduledAtUnix int64
	UpdatedAtUnix   int64
	LastErr         string
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Insert(ctx context.Context, jobType string, payload []byte) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO jobs (type, payload) VALUES ($1, $2::jsonb) RETURNING id`,
		jobType, string(payload)).Scan(&id)
	return id, err
}

const claimSQL = `
UPDATE jobs
SET status = 'running',
	attempts = attempts + 1,
	updated_at = floor(extract(epoch FROM now()))::bigint
WHERE id  = (
	SELECT id FROM jobs
	WHERE status = 'pending'
	ORDER BY id
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
RETURNING id, type, payload, status, attempts,
	created_at, scheduled_at, updated_at, COALESCE(last_error, '')`

func (s *PostgresStore) Claim(ctx context.Context) (Job, error) {
	var (
		j   Job
		raw []byte
	)

	err := s.db.QueryRowContext(ctx, claimSQL).Scan(
		&j.ID, &j.JobType, &raw, &j.Status, &j.Attempts,
		&j.CreatedAtUnix, &j.ScheduledAtUnix, &j.UpdatedAtUnix, &j.LastErr,
	)
	if err != nil {
		return Job{}, fmt.Errorf("claim job: %w", err)
	}
	if err := json.Unmarshal(raw, &j.Payload); err != nil {
		return Job{}, fmt.Errorf("decode payload of job %d: %w", j.ID, err)
	}

	return j, nil
}

func (s *PostgresStore) Complete(ctx context.Context, id int64) error { return nil }

func (s *PostgresStore) Fail(ctx context.Context, id int64, msg string) error { return nil }
