import React from 'react'
import { LineChart, Line, ResponsiveContainer } from 'recharts'
import { GlucoseRecord } from '../types'

interface MiniChartProps {
  records: GlucoseRecord[]
}

const MiniChart: React.FC<MiniChartProps> = ({ records }) => {
  // Подготовка данных для мини-графика (последние 7 записей)
  const prepareChartData = () => {
    return records
      .slice(-7) // Берем последние 7 записей
      .sort((a, b) => new Date(a.measured_at).getTime() - new Date(b.measured_at).getTime())
      .map(record => ({
        value: record.value,
        time: new Date(record.measured_at).getTime()
      }))
  }

  const chartData = prepareChartData()

  if (chartData.length < 2) {
    return (
      <div style={{
        height: '60px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: '#9ca3af',
        fontSize: '12px'
      }}>
        Недостаточно данных
      </div>
    )
  }

  // Определяем цвет линии на основе тренда
  const firstValue = chartData[0].value
  const lastValue = chartData[chartData.length - 1].value
  const lineColor = lastValue > firstValue ? '#ef4444' : 
                   lastValue < firstValue ? '#10b981' : '#6b7280'

  return (
    <div style={{ height: '60px', width: '100%' }}>
      <ResponsiveContainer>
        <LineChart data={chartData}>
          <Line
            type="monotone"
            dataKey="value"
            stroke={lineColor}
            strokeWidth={2}
            dot={false}
            animationDuration={800}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

export default MiniChart