import React from 'react'
import { GlucoseRecord } from '../types'

interface StatsCardsProps {
  records: GlucoseRecord[]
}

const StatsCards: React.FC<StatsCardsProps> = ({ records }) => {
  // Вычисление статистики
  const calculateStats = () => {
    if (records.length === 0) {
      return {
        average: 0,
        min: 0,
        max: 0,
        inRange: 0,
        lowCount: 0,
        highCount: 0
      }
    }

    const values = records.map(r => r.value)
    const average = values.reduce((sum, val) => sum + val, 0) / values.length
    const min = Math.min(...values)
    const max = Math.max(...values)
    
    const inRangeCount = values.filter(v => v >= 3.9 && v <= 7.8).length
    const lowCount = values.filter(v => v < 3.9).length
    const highCount = values.filter(v => v > 7.8).length

    return {
      average,
      min,
      max,
      inRange: Math.round((inRangeCount / values.length) * 100),
      lowCount,
      highCount
    }
  }

  const stats = calculateStats()

  const cardStyle: React.CSSProperties = {
    backgroundColor: '#ffffff',
    borderRadius: '12px',
    padding: '16px',
    border: '1px solid #e5e7eb',
    boxShadow: '0 1px 3px rgba(0, 0, 0, 0.05)',
    transition: 'transform 0.2s ease, box-shadow 0.2s ease'
  }

  const valueStyle: React.CSSProperties = {
    fontSize: '28px',
    fontWeight: 'bold',
    margin: '0 0 4px 0'
  }

  const labelStyle: React.CSSProperties = {
    fontSize: '14px',
    color: '#6b7280',
    margin: 0
  }

  if (records.length === 0) {
    return (
      <div style={{ 
        textAlign: 'center',
        padding: '20px',
        color: '#6b7280'
      }}>
        Недостаточно данных для статистики
      </div>
    )
  }

  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
      gap: '16px',
      marginBottom: '24px'
    }}>
      {/* Среднее значение */}
      <div style={cardStyle}>
        <div style={{ ...valueStyle, color: '#3b82f6' }}>
          {stats.average.toFixed(1)}
        </div>
        <div style={labelStyle}>Среднее</div>
        <div style={{ fontSize: '12px', color: '#9ca3af', marginTop: '4px' }}>
          ммоль/л
        </div>
      </div>

      {/* В норме */}
      <div style={cardStyle}>
        <div style={{ ...valueStyle, color: '#10b981' }}>
          {stats.inRange}%
        </div>
        <div style={labelStyle}>В норме</div>
        <div style={{ fontSize: '12px', color: '#9ca3af', marginTop: '4px' }}>
          3.9-7.8 ммоль/л
        </div>
      </div>

      {/* Минимум */}
      <div style={cardStyle}>
        <div style={{ 
          ...valueStyle, 
          color: stats.min < 3.9 ? '#ef4444' : '#6b7280'
        }}>
          {stats.min.toFixed(1)}
        </div>
        <div style={labelStyle}>Минимум</div>
        <div style={{ fontSize: '12px', color: '#9ca3af', marginTop: '4px' }}>
          ммоль/л
        </div>
      </div>

      {/* Максимум */}
      <div style={cardStyle}>
        <div style={{ 
          ...valueStyle,
          color: stats.max > 7.8 ? '#f59e0b' : '#6b7280'
        }}>
          {stats.max.toFixed(1)}
        </div>
        <div style={labelStyle}>Максимум</div>
        <div style={{ fontSize: '12px', color: '#9ca3af', marginTop: '4px' }}>
          ммоль/л
        </div>
      </div>

      {/* Низкие значения */}
      {stats.lowCount > 0 && (
        <div style={cardStyle}>
          <div style={{ ...valueStyle, color: '#ef4444' }}>
            {stats.lowCount}
          </div>
          <div style={labelStyle}>Низкие</div>
          <div style={{ fontSize: '12px', color: '#9ca3af', marginTop: '4px' }}>
            &lt; 3.9 ммоль/л
          </div>
        </div>
      )}

      {/* Высокие значения */}
      {stats.highCount > 0 && (
        <div style={cardStyle}>
          <div style={{ ...valueStyle, color: '#f59e0b' }}>
            {stats.highCount}
          </div>
          <div style={labelStyle}>Высокие</div>
          <div style={{ fontSize: '12px', color: '#9ca3af', marginTop: '4px' }}>
            &gt; 7.8 ммоль/л
          </div>
        </div>
      )}

      {/* Общее количество */}
      <div style={cardStyle}>
        <div style={{ ...valueStyle, color: '#6b7280' }}>
          {records.length}
        </div>
        <div style={labelStyle}>Измерений</div>
        <div style={{ fontSize: '12px', color: '#9ca3af', marginTop: '4px' }}>
          всего
        </div>
      </div>
    </div>
  )
}

export default StatsCards