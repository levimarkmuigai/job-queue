-- +goose Up
UPDATE jobs SET attempts = 0 WHERE attempts IS NULL;

ALTER TABLE jobs
	ALTER COLUMN attempts SET NOT NULL,
	ALTER COLUMN scheduled_at SET DEFAULT floor(extract(epoch FROM now()))::BIGINT,
	ADD COLUMN updated_at BIGINT NOT NULL DEFAULT floor(extract(epoch FROM now()))::BIGINT;

CREATE INDEX jobs_claim_idx ON jobs (scheduled_at, id) WHERE status = 'pending';

-- +goose Down
DROP INDEX jobs_claim_idx

ALTER TABLE jobs
	DROP COLUMN updated_at,
	ALTER COLUMN scheduled_at  DROP DEFAULT,
	ALTER COLUMN attempts DROP NOT NULL;

