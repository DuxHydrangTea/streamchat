-- +goose Up
ALTER TABLE media_jobs ADD COLUMN storage_key TEXT;
ALTER TABLE media_jobs ADD COLUMN mime_type TEXT;
ALTER TABLE media_jobs ADD COLUMN size_bytes BIGINT;
ALTER TABLE media_jobs ADD COLUMN title TEXT;
ALTER TABLE media_jobs ADD COLUMN duration_seconds DOUBLE PRECISION;
ALTER TABLE media_jobs ADD COLUMN width INTEGER;
ALTER TABLE media_jobs ADD COLUMN height INTEGER;

-- +goose Down
ALTER TABLE media_jobs DROP COLUMN IF EXISTS height;
ALTER TABLE media_jobs DROP COLUMN IF EXISTS width;
ALTER TABLE media_jobs DROP COLUMN IF EXISTS duration_seconds;
ALTER TABLE media_jobs DROP COLUMN IF EXISTS title;
ALTER TABLE media_jobs DROP COLUMN IF EXISTS size_bytes;
ALTER TABLE media_jobs DROP COLUMN IF EXISTS mime_type;
ALTER TABLE media_jobs DROP COLUMN IF EXISTS storage_key;
