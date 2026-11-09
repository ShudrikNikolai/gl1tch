# shared

Общие пакеты всех Go-сервисов. Модуль `github.com/ShudrikNikolai/gl1tch/backend/shared`.

| Пакет | Что даёт |
|---|---|
| `pkg/apperr` | Прикладные ошибки с `Kind`, машинным `Code` и безопасным `Message`. |
| `pkg/apperr/grpcerr` | Перевод `apperr` ↔ gRPC status + `google.rpc.ErrorInfo` с `domain=gl1tch`. |
| `pkg/cache` + `pkg/cache/redis` | Интерфейсы KV/Set/PubSub и Redis-клиент с реконнектом подписки. |
| `pkg/config` | Загрузка env-конфига через `caarlos0/env`; `Base` с общими портами и логом. |
| `pkg/enum` | Доменные строковые enum: роль, статус, пол, порядок. |
| `pkg/grpcx` | gRPC-сервер/клиент с request id, identity, deadline, recover, health, OTel. |
| `pkg/healthcheck` | Self-check для Docker HEALTHCHECK в distroless-образах. |
| `pkg/jwtx` | Ed25519 signer, JWKS, verifier и in-memory revocations по Redis pub/sub. |
| `pkg/kafka` | Общий Sarama-конфиг, producer/consumer и middleware: retry, DLQ, dedupe, recover. |
| `pkg/logger` | `slog` с `request_id`, `user_id`, `trace_id` из контекста. |
| `pkg/metrics` | Минимальный Prometheus `/metrics` без внешних зависимостей. |
| `pkg/otelx` | Инициализация OTLP-трейсинга. |
| `pkg/protox` | Конвертация доменных enum ↔ protobuf enum. |
| `pkg/reqctx` | Заголовки/metadata и значения контекста: request id, user id, роли. |

## Proto

Сгенерированный код лежит в `backend/shared/proto/api`, каталог исключён из git. Генерация из корня репозитория: `task platform:proto:gen`.

## Тесты

```bash
cd backend/shared
go test ./...
```
