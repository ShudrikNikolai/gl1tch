# Контекст сборки: каталог backend/ (нужен соседний модуль shared).
#   docker build -f boards/Dockerfile -t gl1tch/boards ./backend
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY shared ./shared
COPY boards ./boards

# boards и shared это отдельные модули, в образе связываем их через workspace
RUN go work init ./boards ./shared \
 && cd boards \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/boards ./cmd/boards

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/boards /boards

# 8080: HTTP (/healthz, /readyz), 9090: gRPC. Миграции встроены в бинарник.
EXPOSE 8080 9090
ENTRYPOINT ["/boards"]
