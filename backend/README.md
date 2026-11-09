# backend

Go-workspace (`go.work`, Go 1.26.1). Каждый сервис — отдельный модуль, общий код — в `shared`.

## Модули

| Модуль | Назначение |
|---|---|
| `auth` | Логин/пароль, JWT (EdDSA), refresh-сессии в Redis, JWKS. |
| `users` | Профили, роли и статусы пользователей. Источник правды о роли/бане. |
| `boards` | Доски → темы → посты, права и курсорная пагинация. |
| `shared` | Общие пакеты: gRPC, ошибки, Redis, JWT, Kafka, логи, OTel. |
| `notifications` | Kafka-consumer `posts.events` + SSE-поток новых постов по темам. |

Важно: `go test ./...` и `go run ./...` запускайте **из каталога модуля**, не из `backend/`: корень workspace не является Go-модулем.

## Proto

Go-код генерируется в `backend/shared/proto/api` и **не хранится в git** (`shared/.gitignore`). Перед первой сборкой из корня репозитория:

```bash
task platform:proto:gen
```

Задача обновляет `buf.lock`, генерирует Go-типы/stubs и пересобирает `platform/gateway/envoy/api.binpb`. `task proto:build` делает только `buf generate` и дескриптор Envoy не обновляет.

## Запуск сервисов

```bash
cd backend/auth
cp .env.example .env
go run ./cmd/auth
```

Аналогично `backend/users` (`cmd/users`), `backend/boards` (`cmd/boards`) и `backend/notifications` (`.`). Все сервисы применяют встроенные goose-миграции при старте; отдельного шага миграций нет. Каждый сервис использует свою базу, чтобы не делить таблицу миграций.

Зависимости:

- `auth` — Postgres, Redis, gRPC `users`.
- `users` — Postgres.
- `boards` — Postgres, gRPC `users`, Kafka `posts.events`.
- `notifications` — Kafka `posts.events`, gRPC `boards`, JWKS auth.

Все сервисы внутри контейнеров слушают HTTP `8080` и gRPC `9090` (`shared/pkg/config.Base`). Совместный локальный запуск нескольких сервисов на хосте требует переопределения портов.

## Тесты и линт

```bash
cd backend/auth   # или users, boards, shared
go test ./...
golangci-lint run ./...
```

Корневые задачи: `task test`, `task lint`, `task smoke`. CI нет; перед выводами сверяйтесь с baseline и запускайте конкретный пакет:

```bash
go test ./internal/usecase/...
```

## Shared

Общие пакеты и соглашения: [shared/README.md](shared/README.md).
