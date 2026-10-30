-- +goose Up
CREATE TYPE conversation_type AS ENUM ('direct', 'group');
CREATE TYPE member_role AS ENUM ('member', 'admin');
CREATE TYPE message_type AS ENUM ('text', 'link', 'system');

CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type conversation_type NOT NULL,
    name TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_message_at TIMESTAMPTZ
);
CREATE TABLE conversation_members (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role member_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_read_message_id UUID,
    last_read_at TIMESTAMPTZ,
    PRIMARY KEY (conversation_id, user_id)
);
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    client_message_id UUID NOT NULL,
    type message_type NOT NULL DEFAULT 'text',
    text_content TEXT NOT NULL,
    reply_to_id UUID REFERENCES messages(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    edited_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    UNIQUE (sender_id, client_message_id)
);
CREATE INDEX messages_conversation_cursor_idx ON messages (conversation_id, created_at DESC, id DESC);
CREATE INDEX conversation_members_user_idx ON conversation_members (user_id, conversation_id);
CREATE INDEX conversations_last_message_idx ON conversations (last_message_at DESC NULLS LAST, id DESC);

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_members;
DROP TABLE IF EXISTS conversations;
DROP TYPE IF EXISTS message_type;
DROP TYPE IF EXISTS member_role;
DROP TYPE IF EXISTS conversation_type;
