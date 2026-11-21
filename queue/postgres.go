package queue

import (
	"context"
	"database/sql"
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
		`INSERT INTO jobs (jobType, payload) VALUES ($1, $2::jsonb) RETURNING id`,
		jobType, string(payload)).Scan(&id)
	return id, err
}

func (s *PostgresStore) Claim(ctx context.Context) (Job, error) { return Job{}, nil }

func (s *PostgresStore) Complete(ctx context.Context, id int64) error { return nil }

func (s *PostgresStore) Fail(ctx context.Context, id int64, msg string) error { return nil }
