package services

import (
	"testing"
	"time"

	"diabetbot/internal/models"
	"diabetbot/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFoodService_CreateRecord(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewFoodService(db)
	user := testutils.CreateTestUser(db, 123)

	t.Run("CreateFullRecord", func(t *testing.T) {
		foodName := "Овсянка с ягодами"
		foodType := "завтрак"
		carbs := 45.0
		calories := 280
		quantity := "1 порция"
		notes := "Без сахара"

		record, err := service.CreateRecord(user.ID, foodName, foodType, &carbs, &calories, quantity, notes)
		
		require.NoError(t, err)
		assert.NotNil(t, record)
		assert.Equal(t, user.ID, record.UserID)
		assert.Equal(t, foodName, record.FoodName)
		assert.Equal(t, foodType, record.FoodType)
		assert.NotNil(t, record.Carbs)
		assert.Equal(t, carbs, *record.Carbs)
		assert.NotNil(t, record.Calories)
		assert.Equal(t, calories, *record.Calories)
		assert.Equal(t, quantity, record.Quantity)
		assert.Equal(t, notes, record.Notes)
		assert.NotZero(t, record.ID)
		assert.False(t, record.ConsumedAt.IsZero())
	})

	t.Run("CreateMinimalRecord", func(t *testing.T) {
		foodName := "Яблоко"
		foodType := "перекус"

		record, err := service.CreateRecord(user.ID, foodName, foodType, nil, nil, "", "")
		
		require.NoError(t, err)
		assert.Equal(t, foodName, record.FoodName)
		assert.Equal(t, foodType, record.FoodType)
		assert.Nil(t, record.Carbs)
		assert.Nil(t, record.Calories)
		assert.Equal(t, "", record.Quantity)
		assert.Equal(t, "", record.Notes)
	})
}

func TestFoodService_GetUserRecords(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewFoodService(db)
	user1 := testutils.CreateTestUser(db, 123)
	user2 := testutils.CreateTestUser(db, 456)

	// Создаем записи с явным временем (в пределах последних 7 дней)
	now := time.Now()
	record1 := &models.FoodRecord{
		UserID:     user1.ID,
		FoodName:   "Завтрак 1",
		FoodType:   "завтрак",
		ConsumedAt: now.Add(-1 * time.Hour),
	}
	require.NoError(t, db.Create(record1).Error)

	record2 := &models.FoodRecord{
		UserID:     user1.ID,
		FoodName:   "Обед 1",
		FoodType:   "обед",
		ConsumedAt: now.Add(-2 * time.Hour),
	}
	require.NoError(t, db.Create(record2).Error)

	// Запись другого пользователя
	record3 := &models.FoodRecord{
		UserID:     user2.ID,
		FoodName:   "Завтрак 2",
		FoodType:   "завтрак",
		ConsumedAt: now.Add(-3 * time.Hour),
	}
	require.NoError(t, db.Create(record3).Error)

	// Создаем старую запись (8 дней назад)
	oldRecord := &models.FoodRecord{
		UserID:     user1.ID,
		FoodName:   "Старая еда",
		FoodType:   "ужин",
		ConsumedAt: now.AddDate(0, 0, -8),
	}
	require.NoError(t, db.Create(oldRecord).Error)

	t.Run("GetRecordsLast7Days", func(t *testing.T) {
		records, err := service.GetUserRecords(user1.ID, 7)
		
		require.NoError(t, err)
		assert.Len(t, records, 2) // только записи последних 7 дней
		
		// Проверяем, что все записи принадлежат пользователю
		for _, record := range records {
			assert.Equal(t, user1.ID, record.UserID)
		}
	})

	t.Run("GetRecordsLast30Days", func(t *testing.T) {
		records, err := service.GetUserRecords(user1.ID, 30)
		
		require.NoError(t, err)
		assert.Len(t, records, 3) // включая старую запись
	})
}

func TestFoodService_GetRecordsByType(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewFoodService(db)
	user := testutils.CreateTestUser(db, 123)

	// Создаем записи разных типов с явным временем
	now := time.Now()
	records := []*models.FoodRecord{
		{
			UserID:     user.ID,
			FoodName:   "Завтрак 1",
			FoodType:   "завтрак",
			ConsumedAt: now.Add(-1 * time.Hour),
		},
		{
			UserID:     user.ID,
			FoodName:   "Завтрак 2",
			FoodType:   "завтрак",
			ConsumedAt: now.Add(-2 * time.Hour),
		},
		{
			UserID:     user.ID,
			FoodName:   "Обед 1",
			FoodType:   "обед",
			ConsumedAt: now.Add(-3 * time.Hour),
		},
		{
			UserID:     user.ID,
			FoodName:   "Перекус 1",
			FoodType:   "перекус",
			ConsumedAt: now.Add(-4 * time.Hour),
		},
	}
	
	for _, record := range records {
		require.NoError(t, db.Create(record).Error)
	}

	t.Run("GetBreakfastRecords", func(t *testing.T) {
		records, err := service.GetRecordsByType(user.ID, "завтрак", 7)
		
		require.NoError(t, err)
		assert.Len(t, records, 2)
		
		for _, record := range records {
			assert.Equal(t, "завтрак", record.FoodType)
		}
	})

	t.Run("GetLunchRecords", func(t *testing.T) {
		records, err := service.GetRecordsByType(user.ID, "обед", 7)
		
		require.NoError(t, err)
		assert.Len(t, records, 1)
		if len(records) > 0 {
			assert.Equal(t, "Обед 1", records[0].FoodName)
		}
	})

	t.Run("GetNonExistentType", func(t *testing.T) {
		records, err := service.GetRecordsByType(user.ID, "ночной_перекус", 7)
		
		require.NoError(t, err)
		assert.Len(t, records, 0)
	})
}

func TestFoodService_UpdateRecord(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewFoodService(db)
	user := testutils.CreateTestUser(db, 123)
	record := testutils.CreateTestFoodRecord(db, user.ID, "Старое название", "завтрак")

	t.Run("UpdateAllFields", func(t *testing.T) {
		newCarbs := 30.0
		newCalories := 250
		updates := map[string]interface{}{
			"food_name": "Новое название",
			"food_type": "обед",
			"carbs":     &newCarbs,
			"calories":  &newCalories,
			"quantity":  "2 порции",
			"notes":     "Обновленные заметки",
		}

		err := service.UpdateRecord(user.ID, record.ID, updates)
		require.NoError(t, err)

		// Проверяем обновление
		var updated models.FoodRecord
		require.NoError(t, db.First(&updated, record.ID).Error)
		assert.Equal(t, "Новое название", updated.FoodName)
		assert.Equal(t, "обед", updated.FoodType)
		assert.NotNil(t, updated.Carbs)
		assert.Equal(t, newCarbs, *updated.Carbs)
		assert.NotNil(t, updated.Calories)
		assert.Equal(t, newCalories, *updated.Calories)
		assert.Equal(t, "2 порции", updated.Quantity)
		assert.Equal(t, "Обновленные заметки", updated.Notes)
	})

	t.Run("UpdatePartialFields", func(t *testing.T) {
		updates := map[string]interface{}{
			"food_name": "Частично обновлено",
		}

		err := service.UpdateRecord(user.ID, record.ID, updates)
		require.NoError(t, err)

		var updated models.FoodRecord
		require.NoError(t, db.First(&updated, record.ID).Error)
		assert.Equal(t, "Частично обновлено", updated.FoodName)
	})
}

func TestFoodService_DeleteRecord(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewFoodService(db)
	user := testutils.CreateTestUser(db, 123)
	record := testutils.CreateTestFoodRecord(db, user.ID, "Удаляемая еда", "завтрак")

	t.Run("DeleteSuccessful", func(t *testing.T) {
		err := service.DeleteRecord(user.ID, record.ID)
		require.NoError(t, err)

		// Проверяем soft delete
		var deleted models.FoodRecord
		err = db.Unscoped().First(&deleted, record.ID).Error
		require.NoError(t, err)
		assert.NotNil(t, deleted.DeletedAt)
	})
}

func TestFoodService_GetTodayCalories(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewFoodService(db)
	user := testutils.CreateTestUser(db, 123)

	// Создаем записи на сегодня
	today := time.Now().Truncate(24 * time.Hour)
	record1 := &models.FoodRecord{
		UserID:     user.ID,
		FoodName:   "Завтрак",
		FoodType:   "завтрак",
		Calories:   testutils.IntPtr(300),
		ConsumedAt: today.Add(8 * time.Hour),
	}
	require.NoError(t, db.Create(record1).Error)

	record2 := &models.FoodRecord{
		UserID:     user.ID,
		FoodName:   "Обед",
		FoodType:   "обед",
		Calories:   testutils.IntPtr(500),
		ConsumedAt: today.Add(13 * time.Hour),
	}
	require.NoError(t, db.Create(record2).Error)

	// Создаем запись на вчера
	yesterday := &models.FoodRecord{
		UserID:     user.ID,
		FoodName:   "Вчерашний ужин",
		FoodType:   "ужин",
		Calories:   testutils.IntPtr(400),
		ConsumedAt: today.Add(-12 * time.Hour),
	}
	require.NoError(t, db.Create(yesterday).Error)

	t.Run("CalculateTodayCalories", func(t *testing.T) {
		totalCalories, err := service.GetTodayCalories(user.ID)
		
		require.NoError(t, err)
		assert.Equal(t, 800, totalCalories) // 300 + 500
	})
}

func TestFoodService_GetTodayCarbs(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewFoodService(db)
	user := testutils.CreateTestUser(db, 123)

	// Создаем записи на сегодня
	today := time.Now().Truncate(24 * time.Hour)
	record1 := &models.FoodRecord{
		UserID:     user.ID,
		FoodName:   "Каша",
		FoodType:   "завтрак",
		Carbs:      testutils.FloatPtr(45.5),
		ConsumedAt: today.Add(8 * time.Hour),
	}
	require.NoError(t, db.Create(record1).Error)

	record2 := &models.FoodRecord{
		UserID:     user.ID,
		FoodName:   "Хлеб",
		FoodType:   "обед",
		Carbs:      testutils.FloatPtr(25.0),
		ConsumedAt: today.Add(13 * time.Hour),
	}
	require.NoError(t, db.Create(record2).Error)

	t.Run("CalculateTodayCarbs", func(t *testing.T) {
		totalCarbs, err := service.GetTodayCarbs(user.ID)
		
		require.NoError(t, err)
		assert.Equal(t, 70.5, totalCarbs) // 45.5 + 25.0
	})

	t.Run("NoRecordsToday", func(t *testing.T) {
		emptyUser := testutils.CreateTestUser(db, 789)
		totalCarbs, err := service.GetTodayCarbs(emptyUser.ID)
		
		require.NoError(t, err)
		assert.Equal(t, 0.0, totalCarbs)
	})
}