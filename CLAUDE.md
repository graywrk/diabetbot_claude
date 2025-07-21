# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DiabetBot - это Telegram бот и веб-приложение для контроля диабета, написанные на Go 1.24 и React. Проект использует GigaChat API для предоставления персонализированных рекомендаций пользователям.

## Architecture

**Backend (Go):**
- `cmd/main.go` - точка входа приложения
- `internal/app/` - инициализация и управление приложением
- `internal/config/` - конфигурация из переменных окружения
- `internal/database/` - подключение к PostgreSQL через GORM
- `internal/models/` - модели данных (User, GlucoseRecord, FoodRecord, AIRecommendation)
- `internal/services/` - бизнес-логика (UserService, GlucoseService, FoodService, GigaChatService)
- `internal/telegram/` - обработка Telegram Bot API
- `internal/handlers/` - HTTP API обработчики для веб-приложения

**Frontend (React + TypeScript):**
- `web/src/` - исходный код React приложения
- `web/src/pages/` - страницы приложения (Dashboard, GlucoseRecords, FoodRecords, etc.)
- `web/src/components/` - переиспользуемые компоненты
- `web/src/services/` - API клиент для общения с бэкендом
- `web/src/utils/` - утилиты для работы с Telegram WebApp

## Common Development Commands

```bash
# Backend development
go mod tidy                    # Обновить зависимости
go run cmd/main.go            # Запустить сервер для разработки
go build -o bin/diabetbot cmd/main.go  # Сборка бинарника

# Frontend development  
cd web
npm install                   # Установить зависимости
npm run dev                   # Запустить dev сервер
npm run build                 # Собрать для продакшна
npm run lint                  # Проверить код линтером

# Testing
make test                     # Запустить все тесты
make test-backend             # Только Go тесты
make test-frontend            # Только React тесты
make test-coverage            # Тесты с покрытием
make test-watch               # Тесты в watch режиме

# Database operations
# Миграции выполняются автоматически при запуске через GORM AutoMigrate

# Docker operations
docker-compose up -d          # Запустить все сервисы
docker-compose down           # Остановить сервисы
docker-compose logs -f app    # Просмотр логов приложения
docker-compose exec postgres psql -U diabetbot -d diabetbot  # Подключение к БД
```

## Environment Configuration

Создайте файл `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

Обязательные переменные:
- `TELEGRAM_BOT_TOKEN` - токен бота от @BotFather
- `GIGACHAT_API_KEY` - ключ API GigaChat от Сбера  
- `DB_PASSWORD` - пароль для PostgreSQL
- `TELEGRAM_WEBHOOK_URL` - URL вашего сервера для webhook

## Key Integrations

**Telegram Bot API:**
- Обработка команд: /start, /glucose, /food, /stats, /webapp, /help
- Inline клавиатуры для выбора типа приема пищи
- WebApp кнопка для открытия полного приложения
- Webhook endpoint: `/webhook`

**GigaChat API:**
- Авторизация через Bearer токен
- Персонализированные рекомендации по уровню сахара
- Советы по питанию на основе описания еды
- Ответы на общие вопросы о диабете

**PostgreSQL Database:**
- Автоматические миграции через GORM
- Мягкое удаление записей
- Индексы для оптимизации запросов

## Deployment

Приложение развертывается через Docker Compose с Nginx:

```bash
# Настройка SSL сертификатов
mkdir ssl
# Поместите cert.pem и key.pem в директорию ssl/

# Запуск в продакшн
docker-compose up -d

# Настройка webhook после деплоя
curl -X POST "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://yourdomain.com/webhook"}'
```

## Security Notes

- Все API endpoints защищены rate limiting
- SSL обязателен для продакшн развертывания
- Webhook endpoint имеет дополнительное ограничение по частоте запросов
- Приложение работает под непривилегированным пользователем в контейнере

## Medical Data Handling

Проект обрабатывает медицинские данные о диабете:
- Показания глюкозы в крови (ммоль/л)
- Записи о приемах пищи с углеводами и калориями
- ИИ рекомендации от GigaChat (только консультативные, не диагностические)
- Все данные привязаны к Telegram ID пользователя

## Testing

Проект имеет comprehensive test coverage:

**Backend Tests (Go):**
- Unit тесты для всех services (user, glucose, food, gigachat)
- Integration тесты для API handlers
- Тесты Telegram bot логики с mock API
- Покрытие > 85% для критических компонентов

**Frontend Tests (React):**
- Component тесты с React Testing Library
- API integration тесты с Mock Service Worker
- Утилиты для тестирования Telegram WebApp
- Покрытие > 90% для UI компонентов

**Test Commands:**
```bash
make test                    # Все тесты
make test-backend           # Go тесты
make test-frontend          # React тесты
make test-coverage          # С отчетом о покрытии
make test-watch             # Watch режим

# Отдельные тесты
go test ./internal/services/
go test -run TestUserService ./internal/services/
cd web && npm test -- --testNamePattern="Dashboard"
```

**CI/CD Testing:**
- Автоматический запуск тестов в GitHub Actions
- Тестирование в PostgreSQL контейнере
- Security scanning с Trivy и Gosec  
- Docker build тестирование
- Отчеты о покрытии в Codecov

Подробная документация по тестированию: `TESTING.md`