FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -o /yoga ./cmd/yoga

FROM alpine:latest

RUN apk add --no-cache tzdata sqlite

WORKDIR /app

COPY --from=builder /yoga /app/yoga
COPY --from=builder /app/config /app/config
COPY --from=builder /app/web /app/web
COPY --from=builder /app/internal/infrastructure/notifier/templates /app/internal/infrastructure/notifier/templates

ENTRYPOINT ["/app/yoga"]
