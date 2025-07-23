import React from 'react'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Area,
  AreaChart,
  Dot,
  ReferenceLine
} from 'recharts'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'
import { GlucoseRecord } from '../types'

interface GlucoseChartProps {
  records: GlucoseRecord[]
  period: '7days' | '30days' | '90days'
}

const GlucoseChart: React.FC<GlucoseChartProps> = ({ records, period }) => {
  // Подготовка данных для графика
  const prepareChartData = () => {
    return records
      .sort((a, b) => new Date(a.measured_at).getTime() - new Date(b.measured_at).getTime())
      .map(record => ({
        time: new Date(record.measured_at).getTime(),
        value: record.value,
        formattedTime: format(new Date(record.measured_at), 'dd.MM HH:mm', { locale: ru }),
        fullDate: format(new Date(record.measured_at), 'dd MMMM yyyy, HH:mm', { locale: ru }),
        notes: record.notes || ''
      }))
  }

  const chartData = prepareChartData()

  // Определение цвета точки на основе значения
  const getDotColor = (value: number) => {
    if (value < 3.9) return '#f87171' // Красный для низких значений
    if (value > 7.8) return '#fb923c' // Оранжевый для высоких значений
    return '#34d399' // Зеленый для нормальных значений
  }

  // Кастомная точка на графике
  const CustomDot = (props: any) => {
    const { cx, cy, payload } = props
    return (
      <Dot
        cx={cx}
        cy={cy}
        r={6}
        fill={getDotColor(payload.value)}
        stroke="#fff"
        strokeWidth={2}
      />
    )
  }

  // Кастомный тултип
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload
      return (
        <div style={{
          backgroundColor: 'rgba(255, 255, 255, 0.95)',
          padding: '12px 16px',
          borderRadius: '8px',
          boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
          border: '1px solid rgba(0, 0, 0, 0.05)'
        }}>
          <p style={{ 
            margin: '0 0 8px 0',
            fontSize: '24px',
            fontWeight: 'bold',
            color: getDotColor(data.value)
          }}>
            {data.value} ммоль/л
          </p>
          <p style={{ margin: '0 0 4px 0', fontSize: '14px', color: '#6b7280' }}>
            {data.fullDate}
          </p>
          {data.notes && (
            <p style={{ margin: '8px 0 0 0', fontSize: '13px', color: '#374151' }}>
              {data.notes}
            </p>
          )}
          <div style={{ 
            marginTop: '8px',
            paddingTop: '8px',
            borderTop: '1px solid #e5e7eb',
            fontSize: '12px',
            color: '#9ca3af'
          }}>
            {data.value < 3.9 && '⚠️ Низкий уровень глюкозы'}
            {data.value >= 3.9 && data.value <= 7.8 && '✅ Нормальный уровень'}
            {data.value > 7.8 && '⚠️ Высокий уровень глюкозы'}
          </div>
        </div>
      )
    }
    return null
  }

  // Форматирование оси X в зависимости от периода
  const formatXAxis = (tickItem: number) => {
    const date = new Date(tickItem)
    if (period === '7days') {
      return format(date, 'dd.MM', { locale: ru })
    } else if (period === '30days') {
      return format(date, 'dd.MM', { locale: ru })
    }
    return format(date, 'MMM', { locale: ru })
  }

  if (chartData.length === 0) {
    return (
      <div style={{ 
        textAlign: 'center', 
        padding: '40px',
        backgroundColor: '#f9fafb',
        borderRadius: '12px'
      }}>
        <p style={{ color: '#6b7280' }}>Нет данных для отображения</p>
      </div>
    )
  }

  return (
    <div style={{ width: '100%', height: '400px' }}>
      <ResponsiveContainer>
        <AreaChart
          data={chartData}
          margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
        >
          <defs>
            <linearGradient id="colorGlucose" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#818cf8" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#818cf8" stopOpacity={0}/>
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
          <XAxis 
            dataKey="time"
            domain={['dataMin', 'dataMax']}
            type="number"
            tickFormatter={formatXAxis}
            stroke="#6b7280"
            style={{ fontSize: '12px' }}
          />
          <YAxis
            domain={[0, 'auto']}
            stroke="#6b7280"
            style={{ fontSize: '12px' }}
            tickFormatter={(value) => `${value}`}
          />
          <Tooltip content={<CustomTooltip />} />
          
          {/* Референсные линии для целевых значений */}
          <ReferenceLine 
            y={3.9} 
            stroke="#f87171" 
            strokeDasharray="5 5" 
            label={{ value: "Низкий", position: "right", style: { fill: '#f87171', fontSize: '12px' } }}
          />
          <ReferenceLine 
            y={7.8} 
            stroke="#fb923c" 
            strokeDasharray="5 5" 
            label={{ value: "Высокий", position: "right", style: { fill: '#fb923c', fontSize: '12px' } }}
          />
          
          <Area
            type="monotone"
            dataKey="value"
            stroke="#818cf8"
            strokeWidth={3}
            fillOpacity={1}
            fill="url(#colorGlucose)"
            animationDuration={1000}
            animationBegin={0}
            dot={<CustomDot />}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}

export default GlucoseChart