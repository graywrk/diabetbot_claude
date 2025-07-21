import { useState, useEffect } from 'react'
import { ApiService } from '../services/api'
import { User, GlucoseRecord } from '../types'

interface Props {
  user: User
}

function GlucoseRecords({ user }: Props) {
  const [records, setRecords] = useState<GlucoseRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchRecords = async () => {
      try {
        const data = await ApiService.getGlucoseRecords(user.telegram_id)
        setRecords(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Ошибка загрузки записей')
      } finally {
        setLoading(false)
      }
    }

    fetchRecords()
  }, [user.telegram_id])

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

  return (
    <div className="page">
      <h2>История измерений глюкозы</h2>
      
      {records.length === 0 ? (
        <div className="empty-state">
          <p>Записей пока нет</p>
          <button 
            className="button"
            onClick={() => window.history.pushState(null, '', '/webapp/add/glucose')}
          >
            Добавить первую запись
          </button>
        </div>
      ) : (
        <div className="records-list">
          {records.map((record) => (
            <div key={record.id} className="record-item">
              <div className="record-value">
                {record.value} ммоль/л
              </div>
              <div className="record-time">
                {new Date(record.measured_at).toLocaleString()}
              </div>
              <div className="record-notes">
                {record.notes}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default GlucoseRecords