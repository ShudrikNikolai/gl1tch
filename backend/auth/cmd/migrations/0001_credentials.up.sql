CREATE TABLE IF NOT EXISTS credentials (
                                           user_id       uuid PRIMARY KEY,
                                           username      text        NOT NULL,
                                           password_hash text        NOT NULL,
                                           created_at    timestamptz NOT NULL DEFAULT now()
    );

-- логин уникален без учёта регистра (репозиторий сравнивает lower(username))
CREATE UNIQUE INDEX IF NOT EXISTS credentials_username_lower_idx ON credentials (lower(username));
