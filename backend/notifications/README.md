# notifications

Сервис live-уведомлений: читает Kafka-топик `posts.events` и отдаёт новые посты по SSE.

## Как работает

- `boards` после успешного создания поста публикует `events.v1.PostCreated` (best-effort).
- `notifications` читает событие consumer-группой `notifications`, декодирует protobuf и раскладывает по hub в памяти процесса.
- Клиент открывает `GET /v1/threads/{thread_id}/stream` с `Authorization: Bearer <access>`.
- Перед подпиской сервис проверяет через boards gRPC `HasUserPosted`: поток доступен только тем, кто уже писал в теме.
- Событие отдаётся как `event: post.created` с JSON в `data`.
- Heartbeat — каждые 15 секунд. История и replay не хранятся; при обрыве клиент переподключается.
- Сервис отдаёт `/healthz`, `/readyz`, `/metrics` и инициализирует OTel-трейсинг, если задан `OTEL_EXPORTER_OTLP_ENDPOINT`.

## Ограничения

- Один экземпляр сервиса: hub находится в памяти. Для нескольких реплик нужен Redis pub/sub fan-out или отдельная consumer-группа на каждую реплику.
- Kafka-доставка at-least-once, но публикация из boards не транзакционная. Потеря события при падении boards до отправки возможна; outbox пока не реализован.
- SSE идёт напрямую через Caddy (`/api/notifications/*` → `notifications:8080`), без Envoy и gRPC-транскодера.

## Запуск

```bash
cd backend/notifications
cp .env.example .env
go run .
```

Нужны Kafka, auth (JWKS) и boards (gRPC). Порты: HTTP `8080`.

## Проверка

```bash
go test ./...
golangci-lint run ./...
```

Интеграционный сценарий: создать тему, оставить в ней пост, открыть SSE и создать второй пост другим пользователем.
