# Testing Guide

Этот документ описывает структуру тестов и процедуры тестирования для проекта DiabetBot.

## Обзор тестирования

Проект использует многоуровневое тестирование:

- **Unit тесты**: Тестируют отдельные компоненты в изоляции
- **Integration тесты**: Тестируют взаимодействие между компонентами
- **Component тесты**: Тестируют React компоненты
- **API тесты**: Тестируют REST API endpoints
- **End-to-end тесты**: Полноценное тестирование пользовательских сценариев

## Backend тестирование (Go)

### Структура тестов

```
internal/
├── services/
│   ├── user_service.go
│   ├── user_service_test.go
│   ├── glucose_service.go
│   ├── glucose_service_test.go
│   ├── food_service.go
│   ├── food_service_test.go
│   ├── gigachat_service.go
│   └── gigachat_service_test.go
├── handlers/
│   ├── api_handler.go
│   └── api_handler_test.go
├── telegram/
│   ├── bot.go
│   └── bot_test.go
└── testutils/
    └── database.go
```

### Запуск тестов

```bash
# Все тесты
make test-backend

# Тесты с покрытием
make test-coverage-backend

# Отдельные пакеты
go test ./internal/services/
go test ./internal/handlers/
go test ./internal/telegram/

# С подробным выводом
go test -v ./...

# С race detection
go test -race ./...
```

### Тестовая база данных

Тесты используют SQLite in-memory базу данных:

```go
func TestExample(t *testing.T) {
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(db)
    
    service := services.NewUserService(db)
    // тестирование...
}
```

### Mock объекты

Для внешних зависимостей используются mock объекты:

```go
// Mock для Telegram Bot API
type MockBotAPI struct {
    sentMessages []tgbotapi.Chattable
}

// Mock для HTTP сервера (GigaChat API)
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // mock ответ
}))
```

### Покрытие тестами

Текущее покрытие по компонентам:

- **Services**: 95%+ (user, glucose, food services)
- **Handlers**: 90%+ (API endpoints)
- **Telegram Bot**: 85%+ (команды и обработчики)
- **GigaChat Service**: 80%+ (с mock API)

## Frontend тестирование (React + TypeScript)

### Структура тестов

```
web/src/
├── components/
│   ├── Navigation.tsx
│   └── __tests__/
│       └── Navigation.test.tsx
├── pages/
│   ├── Dashboard.tsx
│   └── __tests__/
│       └── Dashboard.test.tsx
├── services/
│   ├── api.ts
│   └── __tests__/
│       └── api.test.ts
└── test/
    └── setup.ts
```

### Запуск тестов

```bash
# Все тесты
make test-frontend

# Тесты с покрытием
make test-coverage-frontend

# В режиме watch
make test-watch

# Jest напрямую
cd web && npm test
cd web && npm run test:coverage
```

### Testing Library

Используется React Testing Library для тестирования компонентов:

```tsx
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

test('renders dashboard with user name', async () => {
  render(<Dashboard user={mockUser} />)
  
  await waitFor(() => {
    expect(screen.getByText('Привет, Test!')).toBeInTheDocument()
  })
})
```

### Mock Service Worker (MSW)

API запросы мокаются через MSW:

```tsx
// setup.ts
export const server = setupServer(
  http.get('/api/v1/user/:telegram_id', () => {
    return HttpResponse.json(mockUser)
  })
)
```

### Покрытие тестами

- **Components**: 90%+ (Navigation, Dashboard)
- **Services**: 95%+ (API client)
- **Utils**: 85%+ (Telegram WebApp utils)

## Integration тестирование

### API Integration тесты

Тестируют полный цикл HTTP запросов:

```go
func TestAPIHandler_CreateGlucoseRecord(t *testing.T) {
    router, handler := setupTestRouter()
    defer testutils.CleanupTestDB(handler.userService.db)

    // Создаем пользователя
    user := testutils.CreateTestUser(handler.userService.db, 123)
    
    // Отправляем POST запрос
    body, _ := json.Marshal(map[string]interface{}{
        "user_id": user.ID,
        "value":   6.5,
        "notes":   "Test note",
    })
    
    req := httptest.NewRequest("POST", "/api/v1/glucose", bytes.NewBuffer(body))
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### Database Integration

Тестируют взаимодействие с реальной базой данных через тестовый PostgreSQL контейнер в CI.

## Continuous Integration

### GitHub Actions

Все тесты автоматически запускаются в CI:

```yaml
- name: Run backend tests
  run: go test -race -coverprofile=coverage.out ./...

- name: Run frontend tests  
  run: npm run test:coverage
```

### Тестовые окружения

- **Unit тесты**: SQLite in-memory
- **Integration тесты**: PostgreSQL container
- **End-to-end тесты**: Docker Compose stack

### Отчеты о покрытии

Отчеты загружаются в Codecov для отслеживания метрик:

- Backend coverage: `coverage.out`
- Frontend coverage: `coverage/lcov.info`

## Лучшие практики

### Naming Convention

```go
// Go tests
func TestServiceName_MethodName(t *testing.T) {
    t.Run("SpecificScenario", func(t *testing.T) {
        // тест
    })
}
```

```tsx
// React tests
describe('ComponentName', () => {
  test('should render with correct props', () => {
    // тест
  })
})
```

### Test Data

Используйте helper функции для создания тестовых данных:

```go
// testutils/database.go
func CreateTestUser(db *gorm.DB, telegramID int64) *models.User {
    user := &models.User{
        TelegramID: telegramID,
        FirstName:  fmt.Sprintf("Test User %d", telegramID),
        IsActive:   true,
    }
    db.Create(user)
    return user
}
```

### Assertions

Используйте четкие и информативные assertions:

```go
// Good
assert.Equal(t, expected, actual, "User ID should match")
assert.Len(t, records, 3, "Should have 3 glucose records")

// Bad  
assert.True(t, len(records) == 3)
```

### Test Isolation

Каждый тест должен быть независимым:

```go
func TestExample(t *testing.T) {
    // Setup
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(db)
    
    // Test
    // ...
    
    // Cleanup происходит автоматически через defer
}
```

## Отладка тестов

### Verbose режим

```bash
go test -v ./internal/services/
```

### Запуск отдельного теста

```bash
go test -run TestUserService_GetOrCreateUser ./internal/services/
```

### Debug в VS Code

Конфигурация для `.vscode/launch.json`:

```json
{
    "name": "Debug Test",
    "type": "go",
    "request": "launch",
    "mode": "test",
    "program": "${workspaceFolder}/internal/services",
    "args": ["-test.run", "TestUserService_GetOrCreateUser"]
}
```

## Метрики качества

### Критерии успешного прохождения

- Покрытие тестами > 85%
- Все тесты проходят без ошибок
- Нет race conditions
- Memory leaks отсутствуют
- Производительность тестов < 30 секунд

### Мониторинг

- **Codecov**: отслеживание покрытия
- **GitHub Actions**: автоматический запуск
- **Dependabot**: обновление зависимостей
- **Security scanning**: поиск уязвимостей

Тестирование является критически важной частью разработки DiabetBot, обеспечивая надежность и качество медицинского приложения.