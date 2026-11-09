# platform

Инфраструктура и шлюз. Приложения (auth, users, boards, frontend) живут отдельно в `apps/`.

## Состав

- `docker-compose.yaml` — подключает `edge`, `gateway`, `observability`, `kafka`.
- `compose/docker-compose.dev-infra.yaml` — Postgres 18 и Redis 8 для разработки, сеть `gl1tch-shared`.
- `compose/initdb/` — создание баз `auth`, `users`, `boards` только на пустом томе Postgres.
- `edge/` — Caddy: `/api/notifications/*` → notifications напрямую (SSE), остальное `/api/*` → Envoy без префикса, `/` → `frontend:3000`.
- `gateway/` — Envoy: JWT через JWKS auth, RBAC, gRPC-JSON transcoder, маршруты auth/users/boards.
- `observability/` — Prometheus, Grafana, Tempo, OTel Collector.
- `kafka/` — Kafka, `kafka-init` и Kafka UI; топик `posts.events` создаётся из `proto/topics.yaml`.

## Команды

```bash
task platform:create:network   # docker network create gl1tch-shared
task platform:dev              # пересоздать Postgres + Redis
task platform:up               # Caddy + Envoy + Kafka + observability
task platform:proto:gen        # buf: Go-код + api.binpb для Envoy
task platform:ps
task platform:logs -- gateway edge
task platform:down
task smoke                     # проверка запущенного стека
```

## Порты

| Что | Адрес |
|---|---|
| Caddy | `http://localhost:80` (`EDGE_HTTP_PORT`) |
| Envoy listener | `envoy:8080`, наружу не публикуется |
| notifications | `notifications:8080`, доступен через Caddy по `/api/notifications/*` |
| Envoy admin | `127.0.0.1:9901` (`ENVOY_ADMIN_PORT`) |
| Prometheus | `127.0.0.1:9090` |
| Grafana | `127.0.0.1:3001` |
| OTLP | `127.0.0.1:4317` (gRPC), `4318` (HTTP) |
| Postgres / Redis | `localhost:5432` / `localhost:6379` |
| Kafka UI (если включить Kafka) | `127.0.0.1:4040` |

## ВАЖНО!!!

- Все приложения и шлюз должны быть в одной внешней сети `gl1tch-shared`.
- Envoy маршрутизирует `auth`, `users` и `boards`; для новых сервисов нужно добавить cluster/route и пересобрать `api.binpb`.
- Go-сервисы отдают `/metrics`, `/healthz` и `/readyz` на `:8080`; Prometheus собирает auth, users, boards и notifications.
- Envoy-конфиги статические (`lds.yaml`, `cds.yaml`) и монтируются в контейнер; после правок нужен рестарт `envoy`.
- Для observability трейсы Tempo не сохраняются между пересозданиями контейнера (тома нет).
