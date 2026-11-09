# frontend

Клиент GL1TCH: Nuxt 4 в режиме SPA и терминальный интерфейс на xterm.js. SSR выключен (`ssr: false`), потому что xterm работает только в браузере.

## Что внутри

- `app/terminal/` — xterm, тема, звук, эффекты, boot-анимация и редактор строки.
- `app/commands/` — реестр команд: парсер, права по роли, подсказки, обработчики.
- `app/api/` — HTTP-клиент к Envoy, refresh-токены, SSE, типы досок и ошибки.
- `app/stores/` — Pinia-состояние сессии, досок, live-постов и пользователей.
- `app/components/` — `TerminalView`, `BanOverlay`.
- `nuxt.config.ts` — SPA, Pinia, CSS и dev-proxy `/api/**`.

## Запуск

```bash
cd frontend
pnpm install
pnpm dev      # http://localhost:3000
pnpm build
pnpm preview
```

`npm *` запрещён в `opencode.json`; используйте pnpm. Для сборки образа нужны `pnpm-lock.yaml` и `pnpm-workspace.yaml` — оба хранятся в git.

## API и авторизация

- `API_BASE = '/api'`. В compose Caddy срезает `/api` и отправляет запрос в `envoy:8080`.
- В dev Nuxt проксирует `/api/**` на `http://localhost:9091/**` (`routeRules`). Это не порт Envoy из compose: listener `:8080` наружу не публикуется, поэтому для локального клиента нужен свой шлюз или правка `routeRules`.
- Access-токен хранится только в памяти, refresh — в HttpOnly cookie. В `localStorage` лежит лишь флаг возможной сессии.
- При `401` клиент делает один общий refresh для всех параллельных запросов; при бане показывает `BanOverlay` и ждёт `banned_until`.
- Ответы об ошибках разбираются из `google.rpc.ErrorInfo` (`USER_BANNED` и т.п.).
- При входе в тему открывается SSE `fetch`-поток `/api/notifications/v1/threads/{id}/stream` с Bearer-токеном. Поток доступен только авторам, уже писавшим в теме; после своего поста клиент перезапускает подписку.

## Команды

Команды регистрируются из экспортов `app/commands/handlers/*` в `app/commands/index.ts`. Чтобы добавить команду, экспортируйте объект `Command` с `name`, `description` и `run`; новый handler-модуль подключите в `index.ts`. `minRole` ограничивает доступ (`guest < user < moderator < admin`). Список доступных команд — `/help`.

Основные группы: `/login`, `/register`, `/logout`, `/whoami`; `/boards`, `/board`, `/threads`, `/thread`, `/post`, `/reply`, `/edit`, `/rm`; модерация `/lock`, `/unlock`, `/pin`, `/unpin`, `/mkboard`, `/rmthread`.

## Проверки

В `package.json` нет скриптов `test` и `lint`. Основная проверка — `pnpm build` (Nuxt + TypeScript). Перед коммитом также полезно вручную открыть терминал и пройти логин/refresh.
