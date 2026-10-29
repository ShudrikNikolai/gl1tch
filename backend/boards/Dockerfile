# syntax=docker/dockerfile:1
# Контекст сборки: каталог backend/ (нужен соседний модуль shared).
#   docker build -f boards/Dockerfile -t gl1tch/boards ./backend
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY shared ./shared
COPY boards ./boards

# boards и shared это отдельные модули, в образе связываем их через workspace
RUN go work init ./boards ./shared

WORKDIR /src/boards

# кеши модулей и сборки живут между сборками: повторный build занимает секунды, а не минуты
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/boards ./cmd/boards

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/boards /boards

# 8080: HTTP (пробы, jwks у auth), 9090: gRPC. Секреты монтируются снаружи, в образ не кладутся.
EXPOSE 8080 9090
ENTRYPOINT ["/boards"]
