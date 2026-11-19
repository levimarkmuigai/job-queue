-- +goose Up
CREATE TABLE jobs (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	type TEXT NOT NULL,
	payload JSONB NOT NULL DEFAULT '{}',
	status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
	attempts INT DEFAULT 0,
	created_at BIGINT NOT NULL DEFAULT floor(extract(epoch FROM now()))::BIGINT,
	scheduled_at BIGINT NOT NULL,
	last_error TEXT
);

-- +goose Down
DROP TABLE jobs;
