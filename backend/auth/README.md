# auth

Сервис аутентификации: логин и пароль (Argon2id), access-токен JWT (EdDSA), refresh-токен с ротацией.
Профиль пользователя (роль, статус) живёт в сервисе `users`.

## Потоки

- **Register**: валидация -> учётка в Postgres -> профиль в `users` (при ошибке учётка откатывается) -> токены.
- **Login**: проверка пароля (при устаревших параметрах хеш обновляется) -> роль и статус из `users` -> токены. Забаненный пользователь получает `USER_BANNED`.
- **Refresh**: refresh-токен одноразовый и заменяется новым. Повторное предъявление использованного токена
  (позже `REFRESH_REUSE_GRACE`) отзывает все сессии пользователя и публикует событие отзыва access-токенов.
- **Logout**: завершает текущую сессию, повторный вызов безопасен.

Refresh-токен хранится в Redis только в виде SHA-256 хеша. Роли берутся из `users` при каждой выдаче access-токена.

## Транспорт

- gRPC `:9090` (`AuthService`), наружу напрямую не публикуется, REST для клиентов делает Envoy.
- HTTP `:8080`: `/.well-known/jwks.json`, `/healthz`, `/readyz`.
- Если Envoy добавил `x-token-transport: cookie`, refresh-токен уходит в HttpOnly cookie (`Set-Cookie`),
  а в теле остаётся пустым. `Refresh` и `Logout` читают его из cookie, если в теле токена нет.

## Запуск локально

1. Создать ключ: `openssl genpkey -algorithm ed25519 -out secrets/jwt-private.pem` (каталог `secrets/` в git не попадает).
2. `cp .env.example .env`, поправить адреса.
3. Применить `cmd/migrations/*.up.sql` к Postgres.
4. `go run ./cmd/auth`

Тесты: `go test ./...`

## Структура

```
cmd/auth            точка входа
cmd/migrations      SQL-миграции
internal/app        запуск и DI (app/di)
internal/transport  gRPC и HTTP, конвертация protobuf <-> модели usecase
internal/usecase    сценарии (register, login, refresh, logout), свои модели
internal/modules    credentials, refresh_tokens (домен, repository, service)
internal/adapters   клиенты других сервисов (users)
internal/infrastructure  config, pg, password, uuid, validate
```
