-- +goose Up
ALTER TABLE chirps
    ADD COLUMN "hashed_password" TEXT NOT NULL
    DEFAULT 'unset';

-- +goose Down
ALTER TABLE chirps
    DROP COLUMN "hashed_password";