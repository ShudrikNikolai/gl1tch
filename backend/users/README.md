# users

Источник правды о профиле пользователя: имя, роль, статус, бан и дата регистрации. `auth` читает роль и статус при логине/выдаче токена, `boards` — имя и статус при записи.

## Транспорт

- gRPC `UserService` (`users.v1`): `CreateUser`, `GetUser`, `ListUsers`, `Me`.
- HTTP только для проб: `/healthz`, `/readyz`.
- REST для клиентов делает Envoy; у `CreateUser` нет HTTP-аннотации — его вызывает `auth` по gRPC.
- HTTP `8080`, gRPC `9090` — одинаково локально и в `apps` compose.
- `Me` берёт пользователя из `x-user-id` (его выставляет Envoy).

Роль хранится в Postgres массивом `user_role[]` и читается в `model.Roles` (`[]enum.Role`). Миграции `00003`/`00004` добавляют и канонизируют значение `moderator`; историческое `moder` конвертер тоже понимает. Новые пользователи создаются со статусом `active` (миграция `00002_active_default.sql`).

## Запуск

```bash
cd backend/users
cp .env.example .env
go run ./cmd/users
```

Нужен Postgres. Миграции goose встроены и применяются при старте; таблица миграций — `users_schema_migrations`.

## Тесты и линт

```bash
go test ./...
golangci-lint run ./...
```
