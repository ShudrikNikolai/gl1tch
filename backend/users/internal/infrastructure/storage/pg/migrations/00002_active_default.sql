-- +goose Up
-- Новый пользователь сразу активен; бан и удаление выставляются отдельно.
ALTER TABLE users ALTER COLUMN user_status SET DEFAULT 'active';

-- +goose Down
ALTER TABLE users ALTER COLUMN user_status SET DEFAULT 'unspecified';
