package services

import (
	"testing"

	"diabetbot/internal/models"
	"diabetbot/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_GetOrCreateUser(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewUserService(db)

	t.Run("CreateNewUser", func(t *testing.T) {
		telegramID := int64(123456789)
		username := "testuser"
		firstName := "Test"
		lastName := "User"
		languageCode := "ru"

		user, err := service.GetOrCreateUser(telegramID, username, firstName, lastName, languageCode)
		
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, telegramID, user.TelegramID)
		assert.Equal(t, username, user.Username)
		assert.Equal(t, firstName, user.FirstName)
		assert.Equal(t, lastName, user.LastName)
		assert.Equal(t, languageCode, user.LanguageCode)
		assert.True(t, user.IsActive)
	})

	t.Run("GetExistingUser", func(t *testing.T) {
		// Создаем пользователя
		existingUser := &models.User{
			TelegramID:   int64(987654321),
			Username:     "existing",
			FirstName:    "Existing",
			LastName:     "User",
			LanguageCode: "en",
			IsActive:     false, // будет обновлено
		}
		require.NoError(t, db.Create(existingUser).Error)

		// Получаем/обновляем пользователя
		user, err := service.GetOrCreateUser(
			existingUser.TelegramID,
			"updated_username",
			"Updated",
			"Name",
			"ru",
		)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, existingUser.ID, user.ID)
		assert.Equal(t, "updated_username", user.Username)
		assert.Equal(t, "Updated", user.FirstName)
		assert.Equal(t, "Name", user.LastName)
		assert.Equal(t, "ru", user.LanguageCode)
		assert.True(t, user.IsActive) // должен быть обновлен
	})
}

func TestUserService_GetByTelegramID(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewUserService(db)

	t.Run("UserExists", func(t *testing.T) {
		// Создаем пользователя
		existingUser := &models.User{
			TelegramID: int64(555666777),
			FirstName:  "Test",
			IsActive:   true,
		}
		require.NoError(t, db.Create(existingUser).Error)

		user, err := service.GetByTelegramID(existingUser.TelegramID)
		
		require.NoError(t, err)
		assert.Equal(t, existingUser.ID, user.ID)
		assert.Equal(t, existingUser.TelegramID, user.TelegramID)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		user, err := service.GetByTelegramID(int64(999999999))
		
		require.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserService_UpdateDiabetesInfo(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)
	
	service := NewUserService(db)

	t.Run("UpdateSuccessful", func(t *testing.T) {
		// Создаем пользователя
		user := &models.User{
			TelegramID: int64(111222333),
			FirstName:  "Test",
			IsActive:   true,
		}
		require.NoError(t, db.Create(user).Error)

		err := service.UpdateDiabetesInfo(user.ID, 2, 6.5)
		require.NoError(t, err)

		// Проверяем обновление
		var updatedUser models.User
		require.NoError(t, db.First(&updatedUser, user.ID).Error)
		assert.NotNil(t, updatedUser.DiabetesType)
		assert.Equal(t, 2, *updatedUser.DiabetesType)
		assert.NotNil(t, updatedUser.TargetGlucose)
		assert.Equal(t, 6.5, *updatedUser.TargetGlucose)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		err := service.UpdateDiabetesInfo(999999, 1, 5.5)
		// GORM не возвращает ошибку при обновлении несуществующей записи
		// но можно добавить проверку в сервис
		assert.NoError(t, err)
	})
}