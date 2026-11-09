# GL1TCH

Форум в эстетике старого терминала: Nuxt-клиент с xterm.js, Go-микросервисы, gRPC внутри и REST/JSON снаружи через Envoy.

## Стек

- **Backend:** Go 1.26, gRPC, PostgreSQL (pgx + goose), Redis.
- **Gateway:** Envoy (JWT, RBAC, gRPC-JSON transcoder), Caddy (edge).
- **Frontend:** Nuxt 4 (`ssr: false`), Pinia, xterm.js.
- **Контракты:** `proto/` + Buf; Go-код и дескриптор для Envoy генерируются.
- **Инфраструктура:** Docker Compose, Prometheus, Grafana, Tempo, OpenTelemetry Collector. Манифесты Kafka есть, но в `platform/docker-compose.yaml` не подключены.

## Архитектура

```text
браузер
  └─ Caddy :80
       ├─ /                       -> frontend:3000 (Nuxt)
       ├─ /api/notifications/*    -> notifications:8080 (SSE, без Envoy)
       └─ /api/*                  -> Envoy :8080 (префикс /api срезается)
                                      ├─ /auth.v1.AuthService/*   -> auth:9090
                                      ├─ /users.v1.UserService/*  -> users:9090
                                      └─ /boards.v1.*Service/*    -> boards:9090
```

REST-пути задаются аннотациями `google.api.http` в `proto/`; Envoy транскодирует JSON в gRPC по дескриптору `platform/gateway/envoy/api.binpb`. Личность пользователя сервисы получают из заголовков `x-user-id` и `x-roles`, которые выставляет Envoy после проверки JWT.

`boards` публикует `events.v1.PostCreated` в Kafka (`posts.events`) после создания поста. `notifications` читает событие и отдаёт его по SSE подписчикам темы, которые уже писали в неё. Поток живёт в памяти одного экземпляра сервиса и не хранит историю.

## Структура

- `backend/auth` — регистрация, логин, JWT, refresh-сессии.
- `backend/users` — профили, роли и статусы пользователей.
- `backend/boards` — доски, темы, посты и права.
- `backend/shared` — общие пакеты Go: gRPC, ошибки, Redis, JWT, Kafka, логи, OTel.
- `backend/notifications` — Kafka-consumer + SSE-поток новых постов по темам.
- `frontend` — Nuxt SPA с терминальным интерфейсом и командами.
- `proto` — исходные контракты и конфигурация Buf.
- `platform` — Caddy, Envoy, observability, dev-инфраструктура.
- `apps` — compose-сборка приложений и Dockerfile-обвязка.

## Быстрый старт (Docker)

Нужны Docker Compose и [Task](https://taskfile.dev/). Для локальной разработки также понадобятся Go 1.26, Node 22 + pnpm и Buf.

```bash
task platform:create:network      # общая сеть gl1tch-shared, один раз
task platform:dev                 # Postgres + Redis
task apps:secrets:jwt             # backend/auth/secrets/jwt-private.pem
cp apps/.env.example apps/.env
task platform:up                  # Caddy, Envoy, Kafka, observability
task apps:up                      # auth, users, boards, notifications, frontend
task smoke                        # проверка HTTP-сценария
```

Открыть <http://localhost>. Если том Postgres уже существовал, базы создаются идемпотентно: `task apps:db:create`.

Остановка: `task apps:down`, затем `task platform:down`.

`task apps:up` сам генерирует proto-код и пересобирает `api.binpb`. Для frontend используется отслеживаемый `pnpm-lock.yaml`.

## Разработка

```bash
# proto: Go-код + дескриптор Envoy (из корня репозитория)
task platform:proto:gen

# backend: команды выполняются из каталога модуля
cd backend/auth
cp .env.example .env
go run ./cmd/auth
go test ./...

# frontend
cd frontend
pnpm install
pnpm dev
```

## Документация

- [backend/README.md](backend/README.md) — устройство Go-workspace и команды.
- [backend/auth/README.md](backend/auth/README.md), [backend/users/README.md](backend/users/README.md), [backend/boards/README.md](backend/boards/README.md), [backend/notifications/README.md](backend/notifications/README.md).
- [frontend/README.md](frontend/README.md) — клиент, SSE и терминальные команды.
- [platform/README.md](platform/README.md) — инфраструктура, шлюз и порты.
- [apps/readme.md](apps/readme.md) — сборка приложений через compose.
