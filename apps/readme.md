# apps: сборка и запуск приложений через docker compose

Здесь живут auth, users, boards, notifications и frontend. Инфраструктура (Postgres, Redis, Kafka) и шлюз (Caddy, Envoy) остаются в `platform/`.
Все контейнеры сидят в общей сети `gl1tch-shared`, поэтому Envoy находит `auth`, `users`, `boards` по имени, Caddy — `frontend` и `notifications`.

```text
apps/
  docker-compose.yaml      сеть и include
  .env.example             пароли БД и теги (копируется в .env)
  Taskfile.yaml            build, up, down, ps, logs, secrets:jwt, db:create
  backend/                 backend.yaml -> auth.yaml, users.yaml, boards.yaml, notifications.yaml
  frontend/frontend.yaml   Nuxt-клиент
```

## Первый запуск

1. Сеть (один раз): `docker network create gl1tch-shared` или `task platform:create:network`.
2. Поднять dev-инфраструктуру: `task platform:dev`.
3. Ключ подписи JWT: `task apps:secrets:jwt` (создаст `backend/auth/secrets/jwt-private.pem`; ключ не должен попадать в git).
4. `cp apps/.env.example apps/.env`.
5. Шлюз: `task platform:up`.
6. `task apps:up` и открыть <http://localhost>.

Если том Postgres уже существует, скрипт `platform/compose/initdb/init.sql` сам не выполнится. Из корня репозитория создайте базы вручную или идемпотентной задачей:

```bash
task apps:db:create
```

`task apps:up` сам генерирует proto-код и пересобирает `api.binpb`.

## Команды

```bash
task apps:build
task apps:up
task apps:down
task apps:ps
task apps:logs -- auth boards
```

## ВАЖНО!!!

- `auth`, `users`, `boards`, `notifications` собираются из контекста `backend/`, потому что им нужен соседний модуль `shared`.
- `frontend` собирается из `frontend/`; Dockerfile использует отслеживаемый `pnpm-lock.yaml` и `pnpm-workspace.yaml`.
- Все Go-сервисы внутри контейнеров слушают HTTP `8080` и gRPC `9090`.
- `boards` публикует события в Kafka (`KAFKA_BROKERS: kafka:9092`), `notifications` читает их и отдаёт SSE через Caddy.
- Порядок старта задан через `depends_on`: `users` → `auth`/`boards` → `notifications`.
- Сервисы падают, если БД ещё недоступна, и перезапускаются сами (`restart: unless-stopped`). `depends_on` между разными compose-проектами не работает.
- У Go-контейнеров есть Docker healthcheck: бинарник сам вызывает `/healthz`. `depends_on` ждёт `service_healthy`.
- Go-сервисы отдают `/metrics` на `:8080`; Prometheus собирает auth, users, boards и notifications.
- JWT-ключ монтируется как файл. На Linux с правами `0600` и владельцем root контейнер (uid 65532) его не прочитает; `secrets:jwt` ставит `644`.
