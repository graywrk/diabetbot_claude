import { useEffect, useState } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { initTelegramWebApp, getTelegramUser } from './utils/telegram'
import { ApiService } from './services/api'
import Navigation from './components/Navigation'
import Dashboard from './pages/Dashboard'
import GlucoseRecords from './pages/GlucoseRecords'
import FoodRecords from './pages/FoodRecords'
import AddRecord from './pages/AddRecord'
import Settings from './pages/Settings'
import { User } from './types'

function App() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const initApp = async () => {
      try {
        console.log('=== WebApp Initialization Debug ===')
        console.log('window.Telegram:', window.Telegram)
        console.log('NODE_ENV:', process.env.NODE_ENV)
        console.log('window.location:', window.location.href)
        
        // Инициализация Telegram Web App
        const webApp = initTelegramWebApp()
        console.log('WebApp initialized:', webApp)
        console.log('WebApp initDataUnsafe:', webApp?.initDataUnsafe)
        console.log('WebApp initData (raw):', (webApp as any)?.initData)
        
        let telegramUser = webApp ? getTelegramUser(webApp) : null
        console.log('Telegram user from WebApp:', telegramUser)
        
        // Дополнительная отладочная информация
        if (webApp) {
          console.log('WebApp platform:', (webApp as any).platform)
          console.log('WebApp version:', (webApp as any).version)
          console.log('WebApp isExpanded:', (webApp as any).isExpanded)
        }
        
        // Fallback только для development режима, когда WebApp действительно недоступен
        if (!telegramUser && process.env.NODE_ENV === 'development') {
          console.log('No Telegram user found, using development fallback...')
          telegramUser = {
            id: 123456789,
            first_name: 'Test User',
            username: 'testuser',
            language_code: 'ru'
          }
          console.log('Using fallback user:', telegramUser)
        }
        
        if (!telegramUser) {
          throw new Error('Не удалось получить данные пользователя из Telegram')
        }

        // Получаем или создаем пользователя через API
        console.log('Fetching user data for ID:', telegramUser.id)
        try {
          const userData = await ApiService.getUser(telegramUser.id)
          console.log('User data received:', userData)
          setUser(userData)
        } catch (apiError) {
          console.error('API Error:', apiError)
          const errorMessage = apiError instanceof Error ? apiError.message : 'Неизвестная ошибка сервера'
          throw new Error(`Ошибка API: ${errorMessage}`)
        }
        
        // Настраиваем тему Telegram
        if (webApp?.themeParams) {
          document.documentElement.style.setProperty('--tg-theme-bg-color', webApp.themeParams.bg_color || '#ffffff')
          document.documentElement.style.setProperty('--tg-theme-text-color', webApp.themeParams.text_color || '#000000')
          document.documentElement.style.setProperty('--tg-theme-button-color', webApp.themeParams.button_color || '#007AFF')
          document.documentElement.style.setProperty('--tg-theme-button-text-color', webApp.themeParams.button_text_color || '#ffffff')
          document.documentElement.style.setProperty('--tg-theme-secondary-bg-color', webApp.themeParams.secondary_bg_color || '#f8f8f8')
        }

        // Показываем веб-приложение
        if (webApp) {
          webApp.ready()
          webApp.expand()
        }
        
      } catch (err) {
        console.error('Ошибка инициализации:', err)
        setError(err instanceof Error ? err.message : 'Неизвестная ошибка')
      } finally {
        setLoading(false)
      }
    }

    initApp()
  }, [])

  if (loading) {
    return (
      <div className="container">
        <div className="loading">
          <div>Загрузка...</div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container">
        <div className="error">
          <h3>Ошибка</h3>
          <p>{error}</p>
          <button className="button" onClick={() => window.location.reload()}>
            Попробовать снова
          </button>
        </div>
      </div>
    )
  }

  if (!user) {
    return (
      <div className="container">
        <div className="error">
          Не удалось загрузить данные пользователя
        </div>
      </div>
    )
  }

  return (
    <Router basename="/webapp">
      <div className="container">
        <Routes>
          <Route path="/" element={<Dashboard user={user} />} />
          <Route path="/glucose" element={<GlucoseRecords user={user} />} />
          <Route path="/food" element={<FoodRecords user={user} />} />
          <Route path="/add/:type" element={<AddRecord user={user} />} />
          <Route path="/settings" element={<Settings user={user} setUser={setUser} />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        <Navigation />
      </div>
    </Router>
  )
}

export default App