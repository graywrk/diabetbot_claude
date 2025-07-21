import { useState } from 'react'
import { ApiService } from '../services/api'
import { User } from '../types'

interface Props {
  user: User
  setUser: (user: User) => void
}

function Settings({ user, setUser }: Props) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  
  // Форма настроек
  const [targetGlucose, setTargetGlucose] = useState(user.target_glucose?.toString() || '')
  const [notifications, setNotifications] = useState(user.notifications || false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    setSuccess(null)

    try {
      const updatedUser = await ApiService.updateUser(user.telegram_id, {
        target_glucose: targetGlucose ? parseFloat(targetGlucose) : null,
        notifications
      })
      
      setUser(updatedUser)
      setSuccess('Настройки сохранены')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteData = async () => {
    if (!window.confirm('Вы уверены, что хотите удалить все данные? Это действие нельзя отменить.')) {
      return
    }

    setLoading(true)
    setError(null)

    try {
      await ApiService.deleteUserData(user.telegram_id)
      setSuccess('Данные удалены')
      // Перезагрузить страницу чтобы пользователь начал заново
      setTimeout(() => window.location.reload(), 2000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка удаления данных')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <h2>Настройки</h2>
      
      {error && <div className="error">{error}</div>}
      {success && <div className="success">{success}</div>}
      
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Целевой уровень глюкозы (ммоль/л)</label>
          <input
            type="number"
            step="0.1"
            value={targetGlucose}
            onChange={(e) => setTargetGlucose(e.target.value)}
            placeholder="Например, 5.5"
          />
          <small>Оставьте пустым, если не хотите устанавливать цель</small>
        </div>

        <div className="form-group">
          <label>
            <input
              type="checkbox"
              checked={notifications}
              onChange={(e) => setNotifications(e.target.checked)}
            />
            Получать уведомления и советы
          </label>
        </div>

        <button type="submit" className="button" disabled={loading}>
          {loading ? 'Сохранение...' : 'Сохранить настройки'}
        </button>
      </form>

      <div className="section">
        <h3>Информация о пользователе</h3>
        <div className="user-info">
          <p><strong>Имя:</strong> {user.first_name} {user.last_name || ''}</p>
          <p><strong>Имя пользователя:</strong> @{user.username}</p>
          <p><strong>Дата регистрации:</strong> {new Date(user.created_at).toLocaleDateString()}</p>
        </div>
      </div>

      <div className="section danger-zone">
        <h3>Опасная зона</h3>
        <button 
          type="button" 
          className="button button-danger"
          onClick={handleDeleteData}
          disabled={loading}
        >
          Удалить все данные
        </button>
        <small>Это действие удалит все ваши записи и не может быть отменено</small>
      </div>
    </div>
  )
}

export default Settings