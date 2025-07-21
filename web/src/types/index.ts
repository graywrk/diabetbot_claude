export interface User {
  id: number
  telegram_id: number
  username?: string
  first_name: string
  last_name?: string
  language_code?: string
  is_active: boolean
  diabetes_type?: number
  target_glucose?: number
  notifications?: boolean
  created_at: string
  updated_at: string
}

export interface GlucoseRecord {
  id: number
  user_id: number
  value: number
  measured_at: string
  notes?: string
  created_at: string
  updated_at: string
}

export interface FoodRecord {
  id: number
  user_id: number
  food_name: string
  food_type: string
  carbs?: number
  calories?: number
  quantity?: string
  consumed_at: string
  notes?: string
  created_at: string
  updated_at: string
}

export interface GlucoseStats {
  average: number
  min: number
  max: number
  count: number
}

export interface TelegramWebApp {
  initDataUnsafe: {
    user?: {
      id: number
      first_name: string
      last_name?: string
      username?: string
      language_code?: string
    }
  }
  themeParams?: {
    bg_color?: string
    text_color?: string
    hint_color?: string
    link_color?: string
    button_color?: string
    button_text_color?: string
    secondary_bg_color?: string
  }
  ready: () => void
  expand: () => void
  close: () => void
  MainButton: {
    text: string
    color: string
    textColor: string
    isVisible: boolean
    isActive: boolean
    show: () => void
    hide: () => void
    enable: () => void
    disable: () => void
    setText: (text: string) => void
    onClick: (callback: () => void) => void
    offClick: (callback: () => void) => void
  }
  BackButton: {
    isVisible: boolean
    show: () => void
    hide: () => void
    onClick: (callback: () => void) => void
    offClick: (callback: () => void) => void
  }
}

export type FoodType = 'завтрак' | 'обед' | 'ужин' | 'перекус' | 'неопределено'