-- +goose Up
ALTER TABLE "User"
    ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE "User"
    DROP COLUMN IF EXISTS must_change_password;
