import { TelegramWebApp } from '../types'

declare global {
  interface Window {
    Telegram: {
      WebApp: TelegramWebApp
    }
  }
}

export function initTelegramWebApp(): TelegramWebApp | null {
  if (typeof window !== 'undefined' && window.Telegram?.WebApp) {
    return window.Telegram.WebApp
  }
  return null
}

export function getTelegramUser(webApp: TelegramWebApp) {
  return webApp.initDataUnsafe?.user || null
}

export function closeTelegramWebApp() {
  const webApp = initTelegramWebApp()
  if (webApp) {
    webApp.close()
  }
}

export function showMainButton(text: string, onClick: () => void) {
  const webApp = initTelegramWebApp()
  if (webApp) {
    webApp.MainButton.setText(text)
    webApp.MainButton.onClick(onClick)
    webApp.MainButton.show()
  }
}

export function hideMainButton() {
  const webApp = initTelegramWebApp()
  if (webApp) {
    webApp.MainButton.hide()
  }
}

export function showBackButton(onClick: () => void) {
  const webApp = initTelegramWebApp()
  if (webApp) {
    webApp.BackButton.onClick(onClick)
    webApp.BackButton.show()
  }
}

export function hideBackButton() {
  const webApp = initTelegramWebApp()
  if (webApp) {
    webApp.BackButton.hide()
  }
}