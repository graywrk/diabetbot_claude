import '@testing-library/jest-dom';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';

// Mock Telegram WebApp
const mockTelegramWebApp = {
  initDataUnsafe: {
    user: {
      id: 123456789,
      first_name: 'Test',
      last_name: 'User',
      username: 'testuser',
      language_code: 'ru'
    }
  },
  themeParams: {
    bg_color: '#ffffff',
    text_color: '#000000',
    button_color: '#007AFF',
    button_text_color: '#ffffff',
    secondary_bg_color: '#f8f8f8'
  },
  ready: jest.fn(),
  expand: jest.fn(),
  close: jest.fn(),
  MainButton: {
    text: '',
    color: '#007AFF',
    textColor: '#ffffff',
    isVisible: false,
    isActive: true,
    show: jest.fn(),
    hide: jest.fn(),
    enable: jest.fn(),
    disable: jest.fn(),
    setText: jest.fn(),
    onClick: jest.fn(),
    offClick: jest.fn(),
  },
  BackButton: {
    isVisible: false,
    show: jest.fn(),
    hide: jest.fn(),
    onClick: jest.fn(),
    offClick: jest.fn(),
  }
};

// Mock window.Telegram
Object.defineProperty(window, 'Telegram', {
  value: {
    WebApp: mockTelegramWebApp
  },
  writable: true
});

// MSW server for API mocking
export const server = setupServer(
  // Mock API endpoints
  http.get('/api/v1/user/:telegram_id', () => {
    return HttpResponse.json({
      id: 1,
      telegram_id: 123456789,
      first_name: 'Test',
      last_name: 'User',
      username: 'testuser',
      is_active: true,
      diabetes_type: 2,
      target_glucose: 6.0,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z'
    });
  }),

  http.get('/api/v1/glucose/:user_id', () => {
    return HttpResponse.json([
      {
        id: 1,
        user_id: 1,
        value: 6.5,
        measured_at: '2024-01-01T08:00:00Z',
        notes: 'После завтрака',
        created_at: '2024-01-01T08:00:00Z',
        updated_at: '2024-01-01T08:00:00Z'
      },
      {
        id: 2,
        user_id: 1,
        value: 5.8,
        measured_at: '2024-01-01T12:00:00Z',
        notes: 'Перед обедом',
        created_at: '2024-01-01T12:00:00Z',
        updated_at: '2024-01-01T12:00:00Z'
      }
    ]);
  }),

  http.get('/api/v1/glucose/:user_id/stats', () => {
    return HttpResponse.json({
      average: 6.2,
      min: 5.8,
      max: 6.5,
      count: 2
    });
  }),

  http.get('/api/v1/food/:user_id', () => {
    return HttpResponse.json([
      {
        id: 1,
        user_id: 1,
        food_name: 'Овсянка с ягодами',
        food_type: 'завтрак',
        carbs: 45.0,
        calories: 280,
        quantity: '1 порция',
        consumed_at: '2024-01-01T08:00:00Z',
        notes: 'Без сахара',
        created_at: '2024-01-01T08:00:00Z',
        updated_at: '2024-01-01T08:00:00Z'
      }
    ]);
  }),

  http.post('/api/v1/glucose', () => {
    return HttpResponse.json({
      id: 3,
      user_id: 1,
      value: 7.0,
      measured_at: '2024-01-01T14:00:00Z',
      notes: 'Новая запись',
      created_at: '2024-01-01T14:00:00Z',
      updated_at: '2024-01-01T14:00:00Z'
    }, { status: 201 });
  }),

  http.post('/api/v1/food', () => {
    return HttpResponse.json({
      id: 2,
      user_id: 1,
      food_name: 'Салат',
      food_type: 'обед',
      carbs: 20.0,
      calories: 150,
      consumed_at: '2024-01-01T12:30:00Z',
      created_at: '2024-01-01T12:30:00Z',
      updated_at: '2024-01-01T12:30:00Z'
    }, { status: 201 });
  })
);

// Start server before all tests
beforeAll(() => server.listen());

// Reset handlers after each test
afterEach(() => server.resetHandlers());

// Clean up after all tests
afterAll(() => server.close());