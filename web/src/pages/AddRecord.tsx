import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ApiService } from '../services/api'
import { User } from '../types'

interface Props {
  user: User
}

function AddRecord({ user }: Props) {
  const { type } = useParams<{ type: string }>()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Состояние для записей глюкозы
  const [glucoseValue, setGlucoseValue] = useState('')
  const [glucosePeriod, setGlucosePeriod] = useState('before_meal')

  // Состояние для записей питания
  const [foodDescription, setFoodDescription] = useState('')
  const [foodCarbs, setFoodCarbs] = useState('')
  const [foodCalories, setFoodCalories] = useState('')
  const [mealType, setMealType] = useState('breakfast')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      if (type === 'glucose') {
        await ApiService.createGlucoseRecord(user.telegram_id, parseFloat(glucoseValue), '')
        navigate('/glucose')
      } else if (type === 'food') {
        await ApiService.createFoodRecord(
          user.telegram_id,
          foodDescription,
          mealType,
          parseFloat(foodCarbs),
          parseFloat(foodCalories),
          '',
          ''
        )
        navigate('/food')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
    } finally {
      setLoading(false)
    }
  }

  const renderGlucoseForm = () => (
    <form onSubmit={handleSubmit}>
      <div className="form-group">
        <label>Уровень глюкозы (ммоль/л)</label>
        <input
          type="number"
          step="0.1"
          value={glucoseValue}
          onChange={(e) => setGlucoseValue(e.target.value)}
          required
        />
      </div>

      <div className="form-group">
        <label>Период измерения</label>
        <select
          value={glucosePeriod}
          onChange={(e) => setGlucosePeriod(e.target.value)}
        >
          <option value="before_meal">До еды</option>
          <option value="after_meal">После еды</option>
          <option value="morning">Утром</option>
          <option value="night">Вечером</option>
        </select>
      </div>

      <button type="submit" className="button" disabled={loading}>
        {loading ? 'Сохранение...' : 'Сохранить'}
      </button>
    </form>
  )

  const renderFoodForm = () => (
    <form onSubmit={handleSubmit}>
      <div className="form-group">
        <label>Описание еды</label>
        <textarea
          value={foodDescription}
          onChange={(e) => setFoodDescription(e.target.value)}
          required
          placeholder="Опишите что вы съели..."
        />
      </div>

      <div className="form-group">
        <label>Углеводы (г)</label>
        <input
          type="number"
          step="0.1"
          value={foodCarbs}
          onChange={(e) => setFoodCarbs(e.target.value)}
          required
        />
      </div>

      <div className="form-group">
        <label>Калории</label>
        <input
          type="number"
          value={foodCalories}
          onChange={(e) => setFoodCalories(e.target.value)}
          required
        />
      </div>

      <div className="form-group">
        <label>Тип приема пищи</label>
        <select
          value={mealType}
          onChange={(e) => setMealType(e.target.value)}
        >
          <option value="breakfast">Завтрак</option>
          <option value="lunch">Обед</option>
          <option value="dinner">Ужин</option>
          <option value="snack">Перекус</option>
        </select>
      </div>

      <button type="submit" className="button" disabled={loading}>
        {loading ? 'Сохранение...' : 'Сохранить'}
      </button>
    </form>
  )

  return (
    <div className="page">
      <h2>
        {type === 'glucose' ? 'Добавить измерение глюкозы' : 'Добавить запись о питании'}
      </h2>
      
      {error && <div className="error">{error}</div>}
      
      {type === 'glucose' ? renderGlucoseForm() : renderFoodForm()}
    </div>
  )
}

export default AddRecord