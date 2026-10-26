-- +goose Up

-- ENUM'ы (идемпотентно — на случай partial state после падения)
CREATE TYPE gender AS ENUM ('male', 'female', 'unspecified');
CREATE TYPE user_role AS ENUM ('user', 'moder', 'unspecified', 'admin');
CREATE TYPE user_status AS ENUM ('active', 'banned', 'deleted', 'pending', 'unspecified');

CREATE TABLE IF NOT EXISTS users (
                                     user_id       uuid PRIMARY KEY,
                                     username      text          NOT NULL,
                                     bio           text          NOT NULL DEFAULT '',
                                     roles         user_role[]   NOT NULL DEFAULT ARRAY['user']::user_role[],
                                     user_status   user_status   NOT NULL DEFAULT 'unspecified',
                                     gender        gender        NOT NULL DEFAULT 'unspecified',

                                     created_at    timestamptz   NOT NULL DEFAULT now(),
    updated_at    timestamptz   NOT NULL DEFAULT now(),
    deleted_at    timestamptz   NULL,
    banned_until  timestamptz   NULL
    );

-- логин уникален без учёта регистра (репозиторий сравнивает lower(username))
CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_idx ON users (lower(username));

-- +goose Down

DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS gender;