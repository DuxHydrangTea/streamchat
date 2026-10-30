-- +goose Up
CREATE TYPE upload_status AS ENUM ('pending', 'completed', 'aborted');
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    storage_key TEXT NOT NULL UNIQUE,
    original_filename TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    status upload_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);
CREATE TABLE message_attachments (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    attachment_id UUID NOT NULL UNIQUE REFERENCES attachments(id) ON DELETE CASCADE,
    PRIMARY KEY(message_id,attachment_id)
);
CREATE INDEX attachments_uploader_status_idx ON attachments(uploaded_by,status,created_at);

-- +goose Down
DROP TABLE IF EXISTS message_attachments;
DROP TABLE IF EXISTS attachments;
DROP TYPE IF EXISTS upload_status;
