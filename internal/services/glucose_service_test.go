package services

import (
	"testing"
	"time"

	"diabetbot/internal/models"
	"diabetbot/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlucoseService_CreateRecord(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewGlucoseService(db)
	user := testutils.CreateTestUser(db, 123)

	t.Run("CreateValidRecord", func(t *testing.T) {
		value := 6.5
		notes := "После завтрака"

		record, err := service.CreateRecord(user.ID, value, notes)
		
		require.NoError(t, err)
		assert.NotNil(t, record)
		assert.Equal(t, user.ID, record.UserID)
		assert.Equal(t, value, record.Value)
		assert.Equal(t, notes, record.Notes)
		assert.NotZero(t, record.ID)
		assert.False(t, record.MeasuredAt.IsZero())
	})

	t.Run("CreateRecordWithoutNotes", func(t *testing.T) {
		value := 5.2

		record, err := service.CreateRecord(user.ID, value, "")
		
		require.NoError(t, err)
		assert.Equal(t, "", record.Notes)
	})
}

func TestGlucoseService_GetUserRecords(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewGlucoseService(db)
	user1 := testutils.CreateTestUser(db, 123)
	user2 := testutils.CreateTestUser(db, 456)

	// Создаем записи для разных пользователей с явным временем
	now := time.Now()
	record1 := &models.GlucoseRecord{
		UserID:     user1.ID,
		Value:      6.0,
		MeasuredAt: now.Add(-1 * time.Hour),
	}
	require.NoError(t, db.Create(record1).Error)

	record2 := &models.GlucoseRecord{
		UserID:     user1.ID,
		Value:      7.2,
		MeasuredAt: now.Add(-2 * time.Hour),
	}
	require.NoError(t, db.Create(record2).Error)

	// Запись другого пользователя
	record3 := &models.GlucoseRecord{
		UserID:     user2.ID,
		Value:      5.5,
		MeasuredAt: now.Add(-3 * time.Hour),
	}
	require.NoError(t, db.Create(record3).Error)

	// Создаем старую запись (8 дней назад)
	oldRecord := &models.GlucoseRecord{
		UserID:     user1.ID,
		Value:      8.0,
		MeasuredAt: now.AddDate(0, 0, -8),
	}
	require.NoError(t, db.Create(oldRecord).Error)

	t.Run("GetRecordsLast7Days", func(t *testing.T) {
		records, err := service.GetUserRecords(user1.ID, 7)
		
		require.NoError(t, err)
		assert.Len(t, records, 2) // только записи последних 7 дней
		
		// Проверяем порядок (сначала новые)
		if len(records) >= 2 {
			assert.True(t, records[0].MeasuredAt.After(records[1].MeasuredAt) || 
				records[0].MeasuredAt.Equal(records[1].MeasuredAt))
		}
	})

	t.Run("GetRecordsLast30Days", func(t *testing.T) {
		records, err := service.GetUserRecords(user1.ID, 30)
		
		require.NoError(t, err)
		assert.Len(t, records, 3) // включая старую запись
	})

	t.Run("NoRecordsFound", func(t *testing.T) {
		emptyUser := testutils.CreateTestUser(db, 789)
		records, err := service.GetUserRecords(emptyUser.ID, 7)
		
		require.NoError(t, err)
		assert.Len(t, records, 0)
	})
}

func TestGlucoseService_GetUserStats(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewGlucoseService(db)
	user := testutils.CreateTestUser(db, 123)

	// Создаем записи с разными значениями и явным временем (в пределах последних 7 дней)
	now := time.Now()
	records := []*models.GlucoseRecord{
		{
			UserID:     user.ID,
			Value:      5.0,
			MeasuredAt: now.Add(-1 * time.Hour),
		},
		{
			UserID:     user.ID,
			Value:      6.0,
			MeasuredAt: now.Add(-2 * time.Hour),
		},
		{
			UserID:     user.ID,
			Value:      7.0,
			MeasuredAt: now.Add(-3 * time.Hour),
		},
		{
			UserID:     user.ID,
			Value:      8.0,
			MeasuredAt: now.Add(-4 * time.Hour),
		},
	}
	
	for _, record := range records {
		require.NoError(t, db.Create(record).Error)
	}

	t.Run("CalculateStats", func(t *testing.T) {
		stats, err := service.GetUserStats(user.ID, 7)
		
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, int64(4), stats.Count)
		assert.Equal(t, 6.5, stats.Average) // (5+6+7+8)/4 = 6.5
		assert.Equal(t, 5.0, stats.Min)
		assert.Equal(t, 8.0, stats.Max)
	})

	t.Run("NoRecords", func(t *testing.T) {
		emptyUser := testutils.CreateTestUser(db, 789)
		stats, err := service.GetUserStats(emptyUser.ID, 7)
		
		require.NoError(t, err)
		assert.Equal(t, int64(0), stats.Count)
	})
}

func TestGlucoseService_GetRecentRecord(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewGlucoseService(db)
	user := testutils.CreateTestUser(db, 123)

	t.Run("GetMostRecentRecord", func(t *testing.T) {
		// Создаем записи в разное время
		older := &models.GlucoseRecord{
			UserID:     user.ID,
			Value:      5.0,
			MeasuredAt: time.Now().Add(-2 * time.Hour),
		}
		require.NoError(t, db.Create(older).Error)

		newer := &models.GlucoseRecord{
			UserID:     user.ID,
			Value:      6.5,
			MeasuredAt: time.Now().Add(-1 * time.Hour),
		}
		require.NoError(t, db.Create(newer).Error)

		record, err := service.GetRecentRecord(user.ID)
		
		require.NoError(t, err)
		assert.Equal(t, newer.ID, record.ID)
		assert.Equal(t, 6.5, record.Value)
	})

	t.Run("NoRecords", func(t *testing.T) {
		emptyUser := testutils.CreateTestUser(db, 789)
		record, err := service.GetRecentRecord(emptyUser.ID)
		
		require.Error(t, err)
		assert.Nil(t, record)
	})
}

func TestGlucoseService_UpdateRecord(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewGlucoseService(db)
	user := testutils.CreateTestUser(db, 123)
	record := testutils.CreateTestGlucoseRecord(db, user.ID, 5.0)

	t.Run("UpdateSuccessful", func(t *testing.T) {
		newValue := 7.2
		newNotes := "Исправленное значение"

		err := service.UpdateRecord(user.ID, record.ID, newValue, newNotes)
		require.NoError(t, err)

		// Проверяем обновление
		var updated models.GlucoseRecord
		require.NoError(t, db.First(&updated, record.ID).Error)
		assert.Equal(t, newValue, updated.Value)
		assert.Equal(t, newNotes, updated.Notes)
	})
}

func TestGlucoseService_DeleteRecord(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewGlucoseService(db)
	user := testutils.CreateTestUser(db, 123)
	record := testutils.CreateTestGlucoseRecord(db, user.ID, 5.0)

	t.Run("DeleteSuccessful", func(t *testing.T) {
		err := service.DeleteRecord(user.ID, record.ID)
		require.NoError(t, err)

		// Проверяем soft delete
		var deleted models.GlucoseRecord
		err = db.Unscoped().First(&deleted, record.ID).Error
		require.NoError(t, err)
		assert.NotNil(t, deleted.DeletedAt)
	})
}