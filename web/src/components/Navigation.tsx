import React from 'react'
import { useLocation, Link } from 'react-router-dom'

const Navigation: React.FC = () => {
  const location = useLocation()

  const navItems = [
    { path: '/', label: 'Главная', icon: '📊' },
    { path: '/glucose', label: 'Глюкоза', icon: '🩸' },
    { path: '/food', label: 'Питание', icon: '🍽️' },
    { path: '/settings', label: 'Настройки', icon: '⚙️' },
  ]

  return (
    <nav className="nav-tabs" style={{ marginTop: '20px' }}>
      {navItems.map((item) => (
        <Link
          key={item.path}
          to={item.path}
          className={`nav-tab ${location.pathname === item.path ? 'active' : ''}`}
          data-testid={`nav-${item.path.slice(1) || 'home'}`}
        >
          <span>{item.icon}</span>
          <span style={{ marginLeft: '4px' }}>{item.label}</span>
        </Link>
      ))}
    </nav>
  )
}

export default Navigation