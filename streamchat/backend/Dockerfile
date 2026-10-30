FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api && CGO_ENABLED=0 go build -o /out/worker ./cmd/worker && CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.24.1 && cp /go/bin/goose /out/goose

FROM alpine:3.21 AS runtime
RUN apk add --no-cache ffmpeg yt-dlp ca-certificates && adduser -D -u 10001 app && mkdir -p /tmp/streamchat-media /app/data && chown -R app:app /tmp/streamchat-media /app/data
COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/worker /usr/local/bin/worker
COPY --from=build /out/goose /usr/local/bin/goose
USER app
WORKDIR /app

FROM runtime AS api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]

FROM runtime AS worker
ENTRYPOINT ["/usr/local/bin/worker"]

FROM runtime AS migrate
ENTRYPOINT ["/usr/local/bin/goose"]
