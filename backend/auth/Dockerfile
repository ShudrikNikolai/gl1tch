FROM golang:1.26-alpine AS build
WORKDIR /src

COPY shared ./shared
COPY auth ./auth

RUN go work init ./auth ./shared \
 && cd auth \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/auth ./cmd/auth

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/auth /auth

EXPOSE 8080 9090
ENTRYPOINT ["/auth"]
