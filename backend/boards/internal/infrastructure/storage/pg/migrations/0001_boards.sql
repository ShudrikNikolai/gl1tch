-- +goose Up

-- Доски: верхний уровень (разделы форума).
-- Перечисления хранятся текстом с CHECK, а не числами из proto: читаемо в psql, порядок значений в proto не важен.
CREATE TABLE IF NOT EXISTS boards (
                                      board_id               uuid        PRIMARY KEY,                  -- UUIDv7, генерирует сервис
                                      slug                   text        NOT NULL CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    title                  text        NOT NULL CHECK (title <> ''),
    description            text        NOT NULL DEFAULT '',
    visibility             text        NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'moderator')),
    sort_order             int         NOT NULL DEFAULT 0,

    -- Счётчики и «последний пост» для превью. Обновляются сервисом, на высокой нагрузке станут горячей строкой.
    thread_count           int         NOT NULL DEFAULT 0 CHECK (thread_count >= 0),
    post_count             int         NOT NULL DEFAULT 0 CHECK (post_count >= 0),
    last_post_author_id    uuid,
    last_post_author_name  text,
    last_post_at           timestamptz,

    created_by             uuid        NOT NULL,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now()
    );

CREATE UNIQUE INDEX IF NOT EXISTS boards_slug_lower_idx ON boards (lower(slug));
CREATE INDEX IF NOT EXISTS boards_sort_idx ON boards (sort_order, board_id);

-- Темы.
CREATE TABLE IF NOT EXISTS threads (
                                       thread_id              uuid        PRIMARY KEY,                  -- UUIDv7: порядок по id = порядок по времени создания
                                       board_id               uuid        NOT NULL REFERENCES boards (board_id),
    title                  text        NOT NULL CHECK (title <> ''),
    slug                   text        NOT NULL,

    author_id              uuid        NOT NULL,
    author_name            text        NOT NULL,                     -- снимок имени на момент создания

    status                 text        NOT NULL DEFAULT 'open'    CHECK (status IN ('open', 'locked', 'archived')),
    category               text        NOT NULL DEFAULT 'general' CHECK (category IN ('general', 'faq', 'flood')),
    visibility             text        NOT NULL DEFAULT 'public'  CHECK (visibility IN ('public', 'author', 'moderator')),
    pinned                 boolean     NOT NULL DEFAULT false,

    -- post_seq выдаёт номера постов (никогда не уменьшается), post_count считает только не удалённые.
    post_seq               int         NOT NULL DEFAULT 0 CHECK (post_seq >= 0),
    post_count             int         NOT NULL DEFAULT 0 CHECK (post_count >= 0),

    last_post_author_id    uuid        NOT NULL,
    last_post_author_name  text        NOT NULL,
    last_post_at           timestamptz NOT NULL DEFAULT now(),

    deleted_at             timestamptz,                              -- мягкое удаление
    deleted_by             uuid,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT threads_board_slug_key UNIQUE (board_id, slug)
    );

-- Список тем доски. Порядок: закреплённые, затем по ключу сортировки, затем по id (тай-брейк для курсора).
-- Запрос должен повторять порядок индекса: ORDER BY pinned DESC, last_post_at DESC, thread_id DESC
CREATE INDEX IF NOT EXISTS threads_board_bump_idx
    ON threads (board_id, pinned DESC, last_post_at DESC, thread_id DESC) WHERE deleted_at IS NULL;

-- Сортировка по созданию: id уже упорядочен по времени, отдельный столбец не нужен.
CREATE INDEX IF NOT EXISTS threads_board_created_idx
    ON threads (board_id, pinned DESC, thread_id DESC) WHERE deleted_at IS NULL;

-- Темы автора (профиль).
CREATE INDEX IF NOT EXISTS threads_author_idx
    ON threads (author_id, thread_id DESC) WHERE deleted_at IS NULL;

-- Посты.
CREATE TABLE IF NOT EXISTS posts (
                                     post_id                uuid        PRIMARY KEY,                  -- UUIDv7
                                     thread_id              uuid        NOT NULL REFERENCES threads (thread_id),
    number                 int         NOT NULL CHECK (number >= 1), -- порядковый номер в теме (#17), из threads.post_seq
    parent_post_id         uuid        REFERENCES posts (post_id),   -- один уровень вложенности и «родитель из той же темы» проверяет usecase

    author_id              uuid        NOT NULL,
    author_name            text        NOT NULL,                     -- снимок имени на момент написания

    body                   text        NOT NULL CHECK (body <> '' AND char_length(body) <= 20000), -- markdown. При удалении не стирается (нужен модераторам), наружу отдаётся пустым

    edited_at              timestamptz,
    deleted_at             timestamptz,                              -- мягкое удаление
    deleted_by             uuid,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT posts_thread_number_key UNIQUE (thread_id, number)
    );

-- Постраничный вывод темы: WHERE thread_id = $1 AND number > $cursor ORDER BY number LIMIT n.
-- Покрывается индексом UNIQUE (thread_id, number), отдельный индекс не нужен.

-- +goose Down
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS boards;
