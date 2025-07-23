import React from 'react'

interface PeriodSelectorProps {
  selectedPeriod: '7days' | '30days' | '90days'
  onPeriodChange: (period: '7days' | '30days' | '90days') => void
}

const PeriodSelector: React.FC<PeriodSelectorProps> = ({ selectedPeriod, onPeriodChange }) => {
  const periods = [
    { value: '7days' as const, label: '7 дней' },
    { value: '30days' as const, label: '30 дней' },
    { value: '90days' as const, label: '90 дней' }
  ]

  return (
    <div style={{
      display: 'flex',
      backgroundColor: '#f3f4f6',
      borderRadius: '8px',
      padding: '4px',
      gap: '2px'
    }}>
      {periods.map(period => (
        <button
          key={period.value}
          onClick={() => onPeriodChange(period.value)}
          style={{
            flex: 1,
            padding: '8px 16px',
            border: 'none',
            borderRadius: '6px',
            backgroundColor: selectedPeriod === period.value ? '#ffffff' : 'transparent',
            color: selectedPeriod === period.value ? '#374151' : '#6b7280',
            fontSize: '14px',
            fontWeight: selectedPeriod === period.value ? '500' : '400',
            cursor: 'pointer',
            transition: 'all 0.2s ease',
            boxShadow: selectedPeriod === period.value ? '0 1px 2px rgba(0, 0, 0, 0.05)' : 'none'
          }}
        >
          {period.label}
        </button>
      ))}
    </div>
  )
}

export default PeriodSelector