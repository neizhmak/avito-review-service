# Builder
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Runner
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /go/bin/goose /bin/goose

COPY --from=builder /app/migrations ./migrations

CMD ["sh", "-c", "sleep 3 && goose -dir ./migrations postgres \"$DB_CONNECTION_STRING\" up && ./server"]