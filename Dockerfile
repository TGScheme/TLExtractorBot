FROM golang:1.26-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
COPY . .
RUN go get ./cmd/sqlgen \
 && go run ./cmd/sqlgen
RUN go get ./cmd/bot ./cmd/importjson \
 && go build -o /out/bot ./cmd/bot \
 && go build -o /out/importjson ./cmd/importjson

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache openjdk21-jre-headless postgresql-client curl unzip \
 && curl -sSL -o /tmp/jadx.zip https://github.com/skylot/jadx/releases/download/v1.5.0/jadx-1.5.0.zip \
 && mkdir -p /opt/jadx \
 && unzip -q /tmp/jadx.zip -d /opt/jadx \
 && rm /tmp/jadx.zip \
 && apk del unzip
COPY --from=builder /out/bot /usr/local/bin/bot
COPY --from=builder /out/importjson /usr/local/bin/importjson
CMD ["bot"]
