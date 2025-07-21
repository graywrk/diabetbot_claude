# Многоступенчатая сборка
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

# Go builder
FROM golang:1.24-alpine AS backend-builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Копируем бинарник Go
COPY --from=backend-builder /app/main .

# Копируем собранный фронтенд
COPY --from=frontend-builder /app/web/dist ./web/dist

# Создаем директорию для логов
RUN mkdir -p logs

# Создаем пользователя для безопасности
RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 8080

CMD ["./main"]