import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ApiService } from '../services/api'
import { User, GlucoseStats, GlucoseRecord } from '../types'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'

interface DashboardProps {
  user: User
}

const Dashboard: React.FC<DashboardProps> = ({ user }) => {
  const [stats, setStats] = useState<GlucoseStats | null>(null)
  const [recentRecord, setRecentRecord] = useState<GlucoseRecord | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true)
        setError(null)

        const [statsData, recordsData] = await Promise.all([
          ApiService.getGlucoseStats(user.id, 7),
          ApiService.getGlucoseRecords(user.id, 1)
        ])

        setStats(statsData)
        if (recordsData.length > 0) {
          setRecentRecord(recordsData[0])
        }
      } catch (err) {
        setError('Ошибка загрузки данных')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [user.id])

  const getGlucoseStatus = (value: number): string => {
    if (value < 3.9) return 'glucose-low'
    if (value > 7.8) return 'glucose-high'
    return 'glucose-normal'
  }

  const getGlucoseStatusText = (value: number): string => {
    if (value < 3.9) return 'Низкий'
    if (value > 7.8) return 'Высокий'
    return 'Нормальный'
  }

  if (loading) {
    return (
      <div className="container">
        <div className="loading" data-testid="loading">
          Загрузка...
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container">
        <div className="error" data-testid="error">
          {error}
        </div>
      </div>
    )
  }

  return (
    <div className="container">
      <header style={{ marginBottom: '20px' }}>
        <h1 style={{ margin: 0, fontSize: '24px' }}>
          Привет, {user.first_name}! 👋
        </h1>
        <p style={{ color: 'var(--gray-color)', margin: '4px 0 0 0' }}>
          Контроль диабета
        </p>
      </header>

      {recentRecord && (
        <div className="card" data-testid="recent-record">
          <h3 style={{ margin: '0 0 12px 0', fontSize: '18px' }}>
            Последнее измерение
          </h3>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <div className={`record-value ${getGlucoseStatus(recentRecord.value)}`}>
                {recentRecord.value} ммоль/л
              </div>
              <div className="record-time">
                {format(new Date(recentRecord.measured_at), 'dd.MM.yyyy HH:mm', { locale: ru })}
              </div>
              <div style={{ fontSize: '14px', color: 'var(--gray-color)' }}>
                {getGlucoseStatusText(recentRecord.value)}
              </div>
            </div>
            <div>
              <Link to="/add/glucose" className="button" style={{ marginBottom: '0', width: 'auto', padding: '8px 16px' }}>
                + Добавить
              </Link>
            </div>
          </div>
        </div>
      )}

      {stats && stats.count > 0 && (
        <div className="card" data-testid="stats-card">
          <h3 style={{ margin: '0 0 16px 0', fontSize: '18px' }}>
            Статистика за 7 дней
          </h3>
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-value">{stats.average.toFixed(1)}</div>
              <div className="stat-label">Средний</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{stats.min.toFixed(1)}</div>
              <div className="stat-label">Минимум</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{stats.max.toFixed(1)}</div>
              <div className="stat-label">Максимум</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{stats.count}</div>
              <div className="stat-label">Измерений</div>
            </div>
          </div>
        </div>
      )}

      <div className="card">
        <h3 style={{ margin: '0 0 16px 0', fontSize: '18px' }}>
          Быстрые действия
        </h3>
        <div style={{ display: 'grid', gap: '12px' }}>
          <Link to="/add/glucose" className="button">
            🩸 Записать уровень сахара
          </Link>
          <Link to="/add/food" className="button button-secondary">
            🍽️ Записать прием пищи
          </Link>
          <Link to="/glucose" className="button button-secondary">
            📊 Посмотреть графики
          </Link>
        </div>
      </div>

      {(!stats || stats.count === 0) && (
        <div className="card" data-testid="empty-state">
          <div style={{ textAlign: 'center', padding: '20px' }}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>📊</div>
            <h3 style={{ margin: '0 0 8px 0' }}>Добро пожаловать!</h3>
            <p style={{ color: 'var(--gray-color)', margin: '0 0 20px 0' }}>
              Начните отслеживать уровень сахара, чтобы увидеть статистику
            </p>
            <Link to="/add/glucose" className="button">
              Первое измерение
            </Link>
          </div>
        </div>
      )}
    </div>
  )
}

export default Dashboard