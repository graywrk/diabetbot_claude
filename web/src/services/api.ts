import axios from 'axios'
import { User, GlucoseRecord, FoodRecord, GlucoseStats } from '../types'

const API_BASE_URL = '/api/v1'

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

export class ApiService {
  // User methods
  static async getUser(telegramId: number): Promise<User> {
    const response = await api.get(`/user/${telegramId}`)
    return response.data
  }

  static async updateDiabetesInfo(telegramId: number, diabetesType: number, targetGlucose: number): Promise<void> {
    await api.put(`/user/${telegramId}/diabetes-info`, {
      diabetes_type: diabetesType,
      target_glucose: targetGlucose,
    })
  }

  // Glucose methods
  static async getGlucoseRecords(userId: number, days = 30): Promise<GlucoseRecord[]> {
    const response = await api.get(`/glucose/${userId}?days=${days}`)
    return response.data
  }

  static async createGlucoseRecord(userId: number, value: number, notes = ''): Promise<GlucoseRecord> {
    const response = await api.post('/glucose', {
      user_id: userId,
      value,
      notes,
    })
    return response.data
  }

  static async updateGlucoseRecord(recordId: number, userId: number, value: number, notes = ''): Promise<void> {
    await api.put(`/glucose/${recordId}`, {
      user_id: userId,
      value,
      notes,
    })
  }

  static async deleteGlucoseRecord(recordId: number, userId: number): Promise<void> {
    await api.delete(`/glucose/${recordId}?user_id=${userId}`)
  }

  static async getGlucoseStats(userId: number, days = 30): Promise<GlucoseStats> {
    const response = await api.get(`/glucose/${userId}/stats?days=${days}`)
    return response.data
  }

  // Food methods
  static async getFoodRecords(userId: number, days = 30, type?: string): Promise<FoodRecord[]> {
    let url = `/food/${userId}?days=${days}`
    if (type) {
      url += `&type=${type}`
    }
    const response = await api.get(url)
    return response.data
  }

  static async createFoodRecord(
    userId: number,
    foodName: string,
    foodType: string,
    carbs?: number,
    calories?: number,
    quantity?: string,
    notes?: string
  ): Promise<FoodRecord> {
    const response = await api.post('/food', {
      user_id: userId,
      food_name: foodName,
      food_type: foodType,
      carbs: carbs || null,
      calories: calories || null,
      quantity: quantity || '',
      notes: notes || '',
    })
    return response.data
  }

  static async updateFoodRecord(
    recordId: number,
    userId: number,
    updates: {
      food_name?: string
      food_type?: string
      carbs?: number | null
      calories?: number | null
      quantity?: string
      notes?: string
    }
  ): Promise<void> {
    await api.put(`/food/${recordId}`, {
      user_id: userId,
      ...updates,
    })
  }

  static async deleteFoodRecord(recordId: number, userId: number): Promise<void> {
    await api.delete(`/food/${recordId}?user_id=${userId}`)
  }
}