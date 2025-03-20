FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates bash

COPY --from=builder /app/channel-bot /app/

RUN mkdir -p /app/config /app/logs

ENV CONFIG_FILE="/app/config/channel-bot.toml"
ENV LOG_FILE="/app/logs/channel-bot.log"

ENTRYPOINT ["/app/channel-bot", "--configfile", "/app/config/channel-bot.toml", "--logfile", "/app/logs/channel-bot.log"]

