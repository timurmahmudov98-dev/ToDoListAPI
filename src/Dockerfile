FROM golang:1.25-alpine AS builder

WORKDIR /app


# Сначала зависимости — чтобы Docker кэшировал этот слой
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Статический бинарник под Linux (main.exe из архива не нужен)
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Этап 2: запуск
FROM alpine:3.20

WORKDIR /app
COPY --from=builder /api /app/api

EXPOSE 8080

CMD ["/app/api"]
