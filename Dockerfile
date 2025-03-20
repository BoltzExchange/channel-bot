FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:3.18.12

WORKDIR /app

RUN apk add --no-cache ca-certificates bash

COPY --from=builder /app/channel-bot /usr/local/bin/

RUN mkdir -p /app/config

ENV CONFIG_FILE="/app/config/channel-bot.toml"

ENTRYPOINT ["/usr/local/bin/channel-bot", "--configfile", "/app/config/channel-bot.toml"]

