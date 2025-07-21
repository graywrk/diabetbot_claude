import { ApiService } from '../api'
import axios from 'axios'

// Mock axios
jest.mock('axios', () => ({
  create: jest.fn(() => ({
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  })),
}))

const mockAxios = axios as jest.Mocked<typeof axios>
const mockAxiosInstance = {
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
}

mockAxios.create.mockReturnValue(mockAxiosInstance)

describe('ApiService', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('User methods', () => {
    test('getUser makes correct API call', async () => {
      const mockUser = {
        id: 1,
        telegram_id: 123456789,
        first_name: 'Test',
        username: 'testuser',
        is_active: true,
      }

      mockAxiosInstance.get.mockResolvedValue({ data: mockUser })

      const result = await ApiService.getUser(123456789)

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/user/123456789')
      expect(result).toEqual(mockUser)
    })

    test('updateDiabetesInfo makes correct API call', async () => {
      mockAxiosInstance.put.mockResolvedValue({ data: {} })

      await ApiService.updateDiabetesInfo(123456789, 2, 6.5)

      expect(mockAxiosInstance.put).toHaveBeenCalledWith(
        '/user/123456789/diabetes-info',
        {
          diabetes_type: 2,
          target_glucose: 6.5,
        }
      )
    })
  })

  describe('Glucose methods', () => {
    test('getGlucoseRecords makes correct API call with default days', async () => {
      const mockRecords = [
        { id: 1, user_id: 1, value: 6.5, measured_at: '2024-01-01T08:00:00Z' }
      ]

      mockAxiosInstance.get.mockResolvedValue({ data: mockRecords })

      const result = await ApiService.getGlucoseRecords(1)

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/glucose/1?days=30')
      expect(result).toEqual(mockRecords)
    })

    test('getGlucoseRecords makes correct API call with custom days', async () => {
      const mockRecords = []
      mockAxiosInstance.get.mockResolvedValue({ data: mockRecords })

      await ApiService.getGlucoseRecords(1, 7)

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/glucose/1?days=7')
    })

    test('createGlucoseRecord makes correct API call', async () => {
      const mockRecord = {
        id: 1,
        user_id: 1,
        value: 6.5,
        notes: 'Test note',
        measured_at: '2024-01-01T08:00:00Z'
      }

      mockAxiosInstance.post.mockResolvedValue({ data: mockRecord })

      const result = await ApiService.createGlucoseRecord(1, 6.5, 'Test note')

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/glucose', {
        user_id: 1,
        value: 6.5,
        notes: 'Test note',
      })
      expect(result).toEqual(mockRecord)
    })

    test('createGlucoseRecord handles empty notes', async () => {
      const mockRecord = { id: 1, user_id: 1, value: 6.5, notes: '' }
      mockAxiosInstance.post.mockResolvedValue({ data: mockRecord })

      await ApiService.createGlucoseRecord(1, 6.5)

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/glucose', {
        user_id: 1,
        value: 6.5,
        notes: '',
      })
    })

    test('updateGlucoseRecord makes correct API call', async () => {
      mockAxiosInstance.put.mockResolvedValue({ data: {} })

      await ApiService.updateGlucoseRecord(1, 1, 7.0, 'Updated note')

      expect(mockAxiosInstance.put).toHaveBeenCalledWith('/glucose/1', {
        user_id: 1,
        value: 7.0,
        notes: 'Updated note',
      })
    })

    test('deleteGlucoseRecord makes correct API call', async () => {
      mockAxiosInstance.delete.mockResolvedValue({ data: {} })

      await ApiService.deleteGlucoseRecord(1, 1)

      expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/glucose/1?user_id=1')
    })

    test('getGlucoseStats makes correct API call', async () => {
      const mockStats = { average: 6.2, min: 5.4, max: 7.1, count: 15 }
      mockAxiosInstance.get.mockResolvedValue({ data: mockStats })

      const result = await ApiService.getGlucoseStats(1, 7)

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/glucose/1/stats?days=7')
      expect(result).toEqual(mockStats)
    })

    test('getGlucoseStats uses default days', async () => {
      const mockStats = { average: 6.2, min: 5.4, max: 7.1, count: 15 }
      mockAxiosInstance.get.mockResolvedValue({ data: mockStats })

      await ApiService.getGlucoseStats(1)

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/glucose/1/stats?days=30')
    })
  })

  describe('Food methods', () => {
    test('getFoodRecords makes correct API call without type filter', async () => {
      const mockRecords = [
        { id: 1, user_id: 1, food_name: 'Овсянка', food_type: 'завтрак' }
      ]
      mockAxiosInstance.get.mockResolvedValue({ data: mockRecords })

      const result = await ApiService.getFoodRecords(1, 7)

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/food/1?days=7')
      expect(result).toEqual(mockRecords)
    })

    test('getFoodRecords makes correct API call with type filter', async () => {
      const mockRecords = []
      mockAxiosInstance.get.mockResolvedValue({ data: mockRecords })

      await ApiService.getFoodRecords(1, 7, 'завтрак')

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/food/1?days=7&type=завтрак')
    })

    test('createFoodRecord makes correct API call with all parameters', async () => {
      const mockRecord = {
        id: 1,
        user_id: 1,
        food_name: 'Овсянка',
        food_type: 'завтрак',
        carbs: 45.0,
        calories: 280,
        quantity: '1 порция',
        notes: 'Без сахара',
      }

      mockAxiosInstance.post.mockResolvedValue({ data: mockRecord })

      const result = await ApiService.createFoodRecord(
        1,
        'Овсянка',
        'завтрак',
        45.0,
        280,
        '1 порция',
        'Без сахара'
      )

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/food', {
        user_id: 1,
        food_name: 'Овсянка',
        food_type: 'завтрак',
        carbs: 45.0,
        calories: 280,
        quantity: '1 порция',
        notes: 'Без сахара',
      })
      expect(result).toEqual(mockRecord)
    })

    test('createFoodRecord handles optional parameters', async () => {
      const mockRecord = {
        id: 1,
        user_id: 1,
        food_name: 'Яблоко',
        food_type: 'перекус',
      }

      mockAxiosInstance.post.mockResolvedValue({ data: mockRecord })

      await ApiService.createFoodRecord(1, 'Яблоко', 'перекус')

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/food', {
        user_id: 1,
        food_name: 'Яблоко',
        food_type: 'перекус',
        carbs: null,
        calories: null,
        quantity: '',
        notes: '',
      })
    })

    test('updateFoodRecord makes correct API call', async () => {
      mockAxiosInstance.put.mockResolvedValue({ data: {} })

      const updates = {
        food_name: 'Обновленное название',
        carbs: 30.0,
      }

      await ApiService.updateFoodRecord(1, 1, updates)

      expect(mockAxiosInstance.put).toHaveBeenCalledWith('/food/1', {
        user_id: 1,
        ...updates,
      })
    })

    test('deleteFoodRecord makes correct API call', async () => {
      mockAxiosInstance.delete.mockResolvedValue({ data: {} })

      await ApiService.deleteFoodRecord(1, 1)

      expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/food/1?user_id=1')
    })
  })

  describe('Error handling', () => {
    test('handles API errors', async () => {
      const errorMessage = 'Network Error'
      mockAxiosInstance.get.mockRejectedValue(new Error(errorMessage))

      await expect(ApiService.getUser(123)).rejects.toThrow(errorMessage)
    })

    test('handles HTTP error responses', async () => {
      const errorResponse = {
        response: {
          status: 404,
          data: { error: 'User not found' }
        }
      }
      mockAxiosInstance.get.mockRejectedValue(errorResponse)

      await expect(ApiService.getUser(123)).rejects.toEqual(errorResponse)
    })
  })
})