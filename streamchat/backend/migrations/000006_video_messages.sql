-- +goose Up
ALTER TYPE message_type ADD VALUE IF NOT EXISTS 'video';

-- +goose Down
-- PostgreSQL cannot remove an enum value without recreating dependent objects.
