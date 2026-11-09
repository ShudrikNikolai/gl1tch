# syntax=docker/dockerfile:1
# Контекст сборки: каталог backend/ (нужен соседний модуль shared).
#   docker build -f auth/Dockerfile -t gl1tch/auth ./backend
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY shared ./shared
COPY auth ./auth

# auth и shared это отдельные модули, в образе связываем их через workspace
RUN go work init ./auth ./shared

WORKDIR /src/auth

# кеши модулей и сборки живут между сборками: повторный build занимает секунды, а не минуты
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/auth ./cmd/auth

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/auth /auth

# 8080: HTTP (пробы, jwks у auth), 9090: gRPC. Секреты монтируются снаружи, в образ не кладутся.
EXPOSE 8080 9090
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=5 CMD ["/auth", "healthcheck"]
ENTRYPOINT ["/auth"]
