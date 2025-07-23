# DiabetBot 🩸

Telegram бот и веб-приложение для контроля диабета с ИИ рекомендациями от YandexGPT/GigaChat.

## Возможности

- 📊 **Контроль глюкозы**: Запись показаний глюкометра с автоматической статистикой
- 🍽️ **Дневник питания**: Учет приемов пищи с углеводами и калориями  
- 🤖 **ИИ рекомендации**: Персонализированные советы от YandexGPT или GigaChat API
  - Поддержка нескольких AI провайдеров (YandexGPT приоритетный)
  - Ограничение: 10 AI запросов на пользователя в день
  - Команда `/limits` для проверки оставшихся запросов
- 📱 **Telegram Mini App**: Полнофункциональное веб-приложение в Telegram
- 📈 **Аналитика**: Графики, статистика и тренды показателей

## Технологии

- **Backend**: Go 1.24, Gin, GORM, PostgreSQL
- **Frontend**: React 18, TypeScript, Vite, Recharts
- **Deployment**: Docker, Docker Compose, Nginx
- **Integrations**: Telegram Bot API, YandexGPT API, GigaChat API

## Быстрый старт

### Требования
- Docker и Docker Compose
- Токен Telegram бота (@BotFather)
- API ключ GigaChat (https://developers.sber.ru/)

### Настройка

1. **Клонируйте репозиторий**
```bash
git clone <repository_url>
cd diabetbot-claude
```

2. **Настройте переменные окружения**
```bash
cp .env.example .env
# Отредактируйте .env файл с вашими ключами
```

3. **Настройте SSL сертификаты**
```bash
mkdir ssl
# Поместите ваши SSL сертификаты в ssl/cert.pem и ssl/key.pem
```

4. **Запустите приложение**
```bash
docker-compose up -d
```

5. **Настройте webhook**
```bash
curl -X POST "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://yourdomain.com/webhook"}'
```

## Разработка

### Backend (Go)
```bash
# Установка зависимостей
go mod tidy

# Запуск для разработки
go run cmd/main.go

# Сборка
go build -o bin/diabetbot cmd/main.go

# Тестирование
make test-backend           # Запуск тестов
make test-coverage-backend  # Тесты с покрытием
```

### Frontend (React)
```bash
cd web
npm install
npm run dev     # Разработка на порту 3000
npm run build   # Сборка для продакшн
npm test        # Запуск тестов
npm run test:coverage  # Тесты с покрытием
```

## Тестирование

Проект имеет comprehensive test coverage с автоматизированным CI/CD:

### Локальное тестирование
```bash
# Все тесты
make test

# По типам
make test-backend     # Go unit + integration тесты  
make test-frontend    # React component тесты
make test-coverage    # Полный отчет о покрытии

# Watch режим для разработки
make test-watch
```

### Test Coverage
- **Backend**: 85%+ покрытие (services, handlers, bot logic)
- **Frontend**: 90%+ покрытие (components, API client, utils)
- **Integration**: API endpoints + database operations
- **Telegram Bot**: Mock-тестирование команд и сценариев

### CI/CD Testing
- ✅ Автоматические тесты в GitHub Actions
- ✅ PostgreSQL integration testing
- ✅ Docker build validation  
- ✅ Security scanning (Trivy, gosec)
- ✅ Coverage reports (Codecov)

Подробнее: [TESTING.md](TESTING.md)

## API Documentation

### REST API Endpoints

**Пользователи:**
- `GET /api/v1/user/{telegram_id}` - Получить пользователя
- `PUT /api/v1/user/{telegram_id}/diabetes-info` - Обновить информацию о диабете

**Показания глюкозы:**
- `GET /api/v1/glucose/{user_id}` - Получить записи
- `POST /api/v1/glucose` - Создать запись
- `PUT /api/v1/glucose/{id}` - Обновить запись
- `DELETE /api/v1/glucose/{id}` - Удалить запись
- `GET /api/v1/glucose/{user_id}/stats` - Статистика

**Питание:**
- `GET /api/v1/food/{user_id}` - Получить записи
- `POST /api/v1/food` - Создать запись
- `PUT /api/v1/food/{id}` - Обновить запись
- `DELETE /api/v1/food/{id}` - Удалить запись

### Telegram Bot Commands

- `/start` - Начать работу с ботом
- `/help` - Помощь и список команд
- `/glucose` - Записать уровень сахара
- `/food` - Записать прием пищи
- `/stats` - Показать статистику
- `/webapp` - Открыть веб-приложение

## Структура проекта

```
diabetbot-claude/
├── cmd/main.go              # Точка входа
├── internal/
│   ├── app/                 # Инициализация приложения
│   ├── config/              # Конфигурация
│   ├── database/            # Подключение к БД
│   ├── handlers/            # HTTP обработчики
│   ├── models/              # Модели данных
│   ├── services/            # Бизнес-логика
│   └── telegram/            # Telegram бот
├── web/                     # React приложение
│   ├── src/
│   │   ├── components/      # Компоненты
│   │   ├── pages/           # Страницы
│   │   ├── services/        # API клиент
│   │   └── utils/           # Утилиты
├── docker-compose.yml       # Docker Compose
├── Dockerfile              # Docker образ
├── nginx.conf              # Nginx конфигурация
└── .env.example            # Пример настроек
```

## Безопасность

- ✅ HTTPS обязателен для продакшн
- ✅ Rate limiting для API и webhook
- ✅ Валидация входных данных
- ✅ Безопасные заголовки HTTP
- ✅ Непривилегированный пользователь в контейнере

## Мониторинг

- Health check endpoint: `/health`
- Логи приложения в `./logs/`
- Docker логи: `docker-compose logs -f`

## Лицензия

MIT License

## Поддержка

Для вопросов и сообщений об ошибках создавайте Issues в репозитории.