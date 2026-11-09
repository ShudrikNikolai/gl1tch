-- +goose NO TRANSACTION
-- +goose Up
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'moderator';

-- +goose Down
-- Postgres не умеет удалять значения enum; откат не требуется.
