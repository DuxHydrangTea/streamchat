-- +goose Up
CREATE TYPE media_job_status AS ENUM ('queued','fetching_metadata','downloading','processing','uploading','completed','failed','cancelled');
CREATE TABLE media_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    requested_by UUID NOT NULL REFERENCES users(id),
    source_url TEXT NOT NULL,
    provider TEXT NOT NULL,
    status media_job_status NOT NULL DEFAULT 'queued',
    attempts INTEGER NOT NULL DEFAULT 0,
    error_code TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);
CREATE INDEX media_jobs_status_created_idx ON media_jobs(status,created_at);
CREATE UNIQUE INDEX media_jobs_active_url_idx ON media_jobs(requested_by,source_url) WHERE status IN ('queued','fetching_metadata','downloading','processing','uploading');

-- +goose Down
DROP INDEX IF EXISTS media_jobs_active_url_idx;
DROP INDEX IF EXISTS media_jobs_status_created_idx;
DROP TABLE IF EXISTS media_jobs;
DROP TYPE IF EXISTS media_job_status;
