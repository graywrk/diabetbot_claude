import { useState, useEffect } from 'react'
import { ApiService } from '../services/api'
import { User, GlucoseRecord } from '../types'
import GlucoseChart from '../components/GlucoseChart'
import PeriodSelector from '../components/PeriodSelector'
import StatsCards from '../components/StatsCards'

interface Props {
  user: User
}

function GlucoseRecords({ user }: Props) {
  const [records, setRecords] = useState<GlucoseRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedPeriod, setSelectedPeriod] = useState<'7days' | '30days' | '90days'>('7days')
  const [viewMode, setViewMode] = useState<'chart' | 'list'>('chart')

  useEffect(() => {
    const fetchRecords = async () => {
      try {
        // Загружаем больше записей для графиков
        const limit = selectedPeriod === '7days' ? 50 : selectedPeriod === '30days' ? 200 : 500
        const data = await ApiService.getGlucoseRecords(user.telegram_id, limit)
        setRecords(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Ошибка загрузки записей')
      } finally {
        setLoading(false)
      }
    }

    fetchRecords()
  }, [user.telegram_id, selectedPeriod])

  if (loading) {
    return (
      <div className="page">
        <div className="loading">Загрузка записей...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="page">
        <div className="error">{error}</div>
      </div>
    )
  }

  // Фильтрация записей по выбранному периоду
  const filterRecordsByPeriod = (records: GlucoseRecord[], period: '7days' | '30days' | '90days') => {
    const now = new Date()
    const daysMap = { '7days': 7, '30days': 30, '90days': 90 }
    const days = daysMap[period]
    const cutoffDate = new Date(now.getTime() - days * 24 * 60 * 60 * 1000)
    
    return records.filter(record => new Date(record.measured_at) >= cutoffDate)
  }

  const filteredRecords = filterRecordsByPeriod(records, selectedPeriod)

  return (
    <div className="page">
      <div style={{ marginBottom: '24px' }}>
        <h2 style={{ margin: '0 0 16px 0' }}>Показатели глюкозы</h2>
        
        {/* Селектор периода */}
        <div style={{ marginBottom: '16px' }}>
          <PeriodSelector 
            selectedPeriod={selectedPeriod}
            onPeriodChange={setSelectedPeriod}
          />
        </div>

        {/* Переключатель вида */}
        <div style={{ 
          display: 'flex', 
          gap: '8px',
          marginBottom: '16px'
        }}>
          <button
            onClick={() => setViewMode('chart')}
            style={{
              padding: '8px 16px',
              border: 'none',
              borderRadius: '6px',
              backgroundColor: viewMode === 'chart' ? '#3b82f6' : '#f3f4f6',
              color: viewMode === 'chart' ? '#ffffff' : '#6b7280',
              fontSize: '14px',
              cursor: 'pointer',
              transition: 'all 0.2s ease'
            }}
          >
            📊 График
          </button>
          <button
            onClick={() => setViewMode('list')}
            style={{
              padding: '8px 16px',
              border: 'none',
              borderRadius: '6px',
              backgroundColor: viewMode === 'list' ? '#3b82f6' : '#f3f4f6',
              color: viewMode === 'list' ? '#ffffff' : '#6b7280',
              fontSize: '14px',
              cursor: 'pointer',
              transition: 'all 0.2s ease'
            }}
          >
            📋 Список
          </button>
        </div>
      </div>
      
      {filteredRecords.length === 0 ? (
        <div className="empty-state">
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>📊</div>
            <h3>Нет данных за выбранный период</h3>
            <p style={{ color: '#6b7280', marginBottom: '20px' }}>
              Добавьте измерения глюкозы, чтобы увидеть красивые графики
            </p>
            <button 
              className="button"
              onClick={() => window.history.pushState(null, '', '/webapp/add/glucose')}
            >
              Добавить измерение
            </button>
          </div>
        </div>
      ) : (
        <>
          {viewMode === 'chart' ? (
            <div>
              {/* Статистические карточки */}
              <StatsCards records={filteredRecords} />
              
              {/* График */}
              <div style={{ 
                backgroundColor: '#ffffff',
                borderRadius: '12px',
                padding: '20px',
                marginBottom: '20px',
                border: '1px solid #e5e7eb'
              }}>
                <h3 style={{ margin: '0 0 20px 0', fontSize: '18px' }}>
                  Динамика уровня глюкозы
                </h3>
                <GlucoseChart records={filteredRecords} period={selectedPeriod} />
              </div>
            </div>
          ) : (
            <div className="records-list">
              {filteredRecords.map((record) => (
                <div 
                  key={record.id} 
                  className="record-item"
                  style={{
                    backgroundColor: '#ffffff',
                    borderRadius: '12px',
                    padding: '16px',
                    marginBottom: '12px',
                    border: '1px solid #e5e7eb',
                    boxShadow: '0 1px 3px rgba(0, 0, 0, 0.05)'
                  }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <div>
                      <div style={{ 
                        fontSize: '24px',
                        fontWeight: 'bold',
                        color: record.value < 3.9 ? '#ef4444' : 
                               record.value > 7.8 ? '#f59e0b' : '#10b981',
                        marginBottom: '4px'
                      }}>
                        {record.value} ммоль/л
                      </div>
                      <div style={{ 
                        fontSize: '14px',
                        color: '#6b7280',
                        marginBottom: '8px'
                      }}>
                        {new Date(record.measured_at).toLocaleString('ru-RU', {
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit'
                        })}
                      </div>
                      {record.notes && (
                        <div style={{ 
                          fontSize: '14px',
                          color: '#374151',
                          backgroundColor: '#f9fafb',
                          padding: '8px',
                          borderRadius: '6px'
                        }}>
                          {record.notes}
                        </div>
                      )}
                    </div>
                    <div style={{
                      fontSize: '12px',
                      padding: '4px 8px',
                      borderRadius: '12px',
                      backgroundColor: record.value < 3.9 ? '#fef2f2' : 
                                       record.value > 7.8 ? '#fefbf2' : '#f0fdf4',
                      color: record.value < 3.9 ? '#dc2626' : 
                             record.value > 7.8 ? '#d97706' : '#16a34a',
                      fontWeight: '500'
                    }}>
                      {record.value < 3.9 ? 'Низкий' : 
                       record.value > 7.8 ? 'Высокий' : 'Норма'}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}

export default GlucoseRecords