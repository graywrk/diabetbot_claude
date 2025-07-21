import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import Dashboard from '../Dashboard'
import { User } from '../../types'
import { ApiService } from '../../services/api'

// Mock the API service
jest.mock('../../services/api')
const mockApiService = ApiService as jest.Mocked<typeof ApiService>

const mockUser: User = {
  id: 1,
  telegram_id: 123456789,
  first_name: 'Test',
  last_name: 'User',
  username: 'testuser',
  is_active: true,
  diabetes_type: 2,
  target_glucose: 6.0,
  language_code: 'ru',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z'
}

const mockStats = {
  average: 6.2,
  min: 5.4,
  max: 7.1,
  count: 15
}

const mockRecentRecord = {
  id: 1,
  user_id: 1,
  value: 6.5,
  measured_at: '2024-01-01T08:00:00Z',
  notes: 'После завтрака',
  created_at: '2024-01-01T08:00:00Z',
  updated_at: '2024-01-01T08:00:00Z'
}

const DashboardWrapper: React.FC<{ user: User }> = ({ user }) => (
  <BrowserRouter>
    <Dashboard user={user} />
  </BrowserRouter>
)

describe('Dashboard Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('displays loading state initially', () => {
    mockApiService.getGlucoseStats.mockImplementation(() => new Promise(() => {}))
    mockApiService.getGlucoseRecords.mockImplementation(() => new Promise(() => {}))

    render(<DashboardWrapper user={mockUser} />)
    
    expect(screen.getByTestId('loading')).toBeInTheDocument()
    expect(screen.getByText('Загрузка...')).toBeInTheDocument()
  })

  test('displays error state when API calls fail', async () => {
    mockApiService.getGlucoseStats.mockRejectedValue(new Error('API Error'))
    mockApiService.getGlucoseRecords.mockRejectedValue(new Error('API Error'))

    render(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(screen.getByTestId('error')).toBeInTheDocument()
      expect(screen.getByText('Ошибка загрузки данных')).toBeInTheDocument()
    })
  })

  test('displays welcome message with user name', async () => {
    mockApiService.getGlucoseStats.mockResolvedValue(mockStats)
    mockApiService.getGlucoseRecords.mockResolvedValue([mockRecentRecord])

    render(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(screen.getByText('Привет, Test! 👋')).toBeInTheDocument()
      expect(screen.getByText('Контроль диабета')).toBeInTheDocument()
    })
  })

  test('displays recent record when available', async () => {
    mockApiService.getGlucoseStats.mockResolvedValue(mockStats)
    mockApiService.getGlucoseRecords.mockResolvedValue([mockRecentRecord])

    render(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(screen.getByTestId('recent-record')).toBeInTheDocument()
      expect(screen.getByText('Последнее измерение')).toBeInTheDocument()
      expect(screen.getByText('6.5 ммоль/л')).toBeInTheDocument()
      expect(screen.getByText('После завтрака')).toBeInTheDocument()
    })
  })

  test('displays stats card when data is available', async () => {
    mockApiService.getGlucoseStats.mockResolvedValue(mockStats)
    mockApiService.getGlucoseRecords.mockResolvedValue([mockRecentRecord])

    render(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(screen.getByTestId('stats-card')).toBeInTheDocument()
      expect(screen.getByText('Статистика за 7 дней')).toBeInTheDocument()
      expect(screen.getByText('6.2')).toBeInTheDocument() // average
      expect(screen.getByText('5.4')).toBeInTheDocument() // min
      expect(screen.getByText('7.1')).toBeInTheDocument() // max
      expect(screen.getByText('15')).toBeInTheDocument() // count
    })
  })

  test('displays empty state when no records exist', async () => {
    const emptyStats = { average: 0, min: 0, max: 0, count: 0 }
    mockApiService.getGlucoseStats.mockResolvedValue(emptyStats)
    mockApiService.getGlucoseRecords.mockResolvedValue([])

    render(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(screen.getByTestId('empty-state')).toBeInTheDocument()
      expect(screen.getByText('Добро пожаловать!')).toBeInTheDocument()
      expect(screen.getByText('Начните отслеживать уровень сахара, чтобы увидеть статистику')).toBeInTheDocument()
      expect(screen.getByText('Первое измерение')).toBeInTheDocument()
    })
  })

  test('displays quick actions', async () => {
    mockApiService.getGlucoseStats.mockResolvedValue(mockStats)
    mockApiService.getGlucoseRecords.mockResolvedValue([mockRecentRecord])

    render(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(screen.getByText('Быстрые действия')).toBeInTheDocument()
      expect(screen.getByText('🩸 Записать уровень сахара')).toBeInTheDocument()
      expect(screen.getByText('🍽️ Записать прием пищи')).toBeInTheDocument()
      expect(screen.getByText('📊 Посмотреть графики')).toBeInTheDocument()
    })
  })

  test('correctly determines glucose status', async () => {
    // Test low glucose
    const lowRecord = { ...mockRecentRecord, value: 3.5 }
    mockApiService.getGlucoseStats.mockResolvedValue(mockStats)
    mockApiService.getGlucoseRecords.mockResolvedValue([lowRecord])

    const { rerender } = render(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(screen.getByText('3.5 ммоль/л')).toBeInTheDocument()
      expect(screen.getByText('Низкий')).toBeInTheDocument()
    })

    // Test high glucose
    const highRecord = { ...mockRecentRecord, value: 8.5 }
    mockApiService.getGlucoseRecords.mockResolvedValue([highRecord])
    
    rerender(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(screen.getByText('8.5 ммоль/л')).toBeInTheDocument()
      expect(screen.getByText('Высокий')).toBeInTheDocument()
    })

    // Test normal glucose
    mockApiService.getGlucoseRecords.mockResolvedValue([mockRecentRecord])
    
    rerender(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(screen.getByText('6.5 ммоль/л')).toBeInTheDocument()
      expect(screen.getByText('Нормальный')).toBeInTheDocument()
    })
  })

  test('has correct navigation links', async () => {
    mockApiService.getGlucoseStats.mockResolvedValue(mockStats)
    mockApiService.getGlucoseRecords.mockResolvedValue([mockRecentRecord])

    render(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      const addGlucoseLinks = screen.getAllByText(/Записать.*сахара?/i)
      expect(addGlucoseLinks.length).toBeGreaterThan(0)
      
      const addFoodLink = screen.getByText('🍽️ Записать прием пищи')
      expect(addFoodLink.closest('a')).toHaveAttribute('href', '/add/food')
      
      const chartsLink = screen.getByText('📊 Посмотреть графики')
      expect(chartsLink.closest('a')).toHaveAttribute('href', '/glucose')
    })
  })

  test('calls API with correct parameters', async () => {
    mockApiService.getGlucoseStats.mockResolvedValue(mockStats)
    mockApiService.getGlucoseRecords.mockResolvedValue([mockRecentRecord])

    render(<DashboardWrapper user={mockUser} />)
    
    await waitFor(() => {
      expect(mockApiService.getGlucoseStats).toHaveBeenCalledWith(1, 7)
      expect(mockApiService.getGlucoseRecords).toHaveBeenCalledWith(1, 1)
    })
  })
})