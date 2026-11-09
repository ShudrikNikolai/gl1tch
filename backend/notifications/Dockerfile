# syntax=docker/dockerfile:1
# Контекст сборки: каталог backend/ (нужен соседний модуль shared).
#   docker build -f notifications/Dockerfile -t gl1tch/notifications ./backend
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY shared ./shared
COPY notifications ./notifications

RUN go work init ./notifications ./shared

WORKDIR /src/notifications

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/notifications .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/notifications /notifications

EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=5 CMD ["/notifications", "healthcheck"]
ENTRYPOINT ["/notifications"]
