# syntax=docker/dockerfile:1
# Nuxt (Nitro) в production-режиме: node .output/server/index.mjs
FROM node:22-alpine AS base
WORKDIR /app
ENV CI=true

# --- зависимости: слой пересобирается только при смене манифестов ---
FROM base AS deps
COPY package.json pnpm-lock.yaml* yarn.lock* package-lock.json* .npmrc* ./
# кеш менеджеров пакетов лежит в cache-mount и в слой не попадает, поэтому очистка кеша не нужна
# hadolint ignore=DL3060
RUN --mount=type=cache,target=/root/.npm \
    --mount=type=cache,target=/root/.local/share/pnpm/store \
    --mount=type=cache,target=/usr/local/share/.cache/yarn \
    if [ -f pnpm-lock.yaml ]; then corepack enable && pnpm install --frozen-lockfile; \
    elif [ -f yarn.lock ]; then corepack enable && yarn install --frozen-lockfile; \
    elif [ -f package-lock.json ]; then npm ci; \
    else echo "нужен lock-файл (pnpm-lock.yaml, yarn.lock или package-lock.json)" >&2; exit 1; fi

# --- сборка ---
FROM deps AS build
COPY . .
RUN if [ -f pnpm-lock.yaml ]; then pnpm run build; \
    elif [ -f yarn.lock ]; then yarn build; \
    else npm run build; fi

# --- рантайм: только собранный сервер, без node_modules и исходников ---
FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production \
    NITRO_HOST=0.0.0.0 \
    NITRO_PORT=3000

COPY --from=build --chown=node:node /app/.output ./.output

USER node
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
