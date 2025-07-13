# Базовый образ
FROM golang:1.24-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git

# Рабочая директория
WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /calendar-app ./cmd/main.go

# Финальный образ
FROM alpine:latest

# Установка зависимостей для runtime
RUN apk --no-cache add ca-certificates

# Копируем бинарник из builder
COPY --from=builder /calendar-app /calendar-app
COPY config.yaml /config.yaml

# Создаем директорию для логов
RUN mkdir -p /var/log/calendar

# Порт приложения
EXPOSE 8080

# Команда запуска
CMD ["/calendar-app"]