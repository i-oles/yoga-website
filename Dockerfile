FROM golang:1.23-alpine AS builder

RUN apk add --no-cache make git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux make build

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/bin/yoga /app/yoga
COPY --from=builder /app/config /app/config

ENTRYPOINT ["/app/yoga"]
