package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"diabetbot/internal/models"
	"diabetbot/internal/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestRouter() (*gin.Engine, *APIHandler, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	
	db := testutils.SetupTestDB(&testing.T{})
	handler := NewAPIHandler(db)
	
	router := gin.New()
	
	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/user/:telegram_id", handler.GetUser)
		api.PUT("/user/:telegram_id/diabetes-info", handler.UpdateDiabetesInfo)
		
		api.GET("/glucose/:user_id", handler.GetGlucoseRecords)
		api.POST("/glucose", handler.CreateGlucoseRecord)
		api.PUT("/glucose/:id", handler.UpdateGlucoseRecord)
		api.DELETE("/glucose/:id", handler.DeleteGlucoseRecord)
		api.GET("/glucose/:user_id/stats", handler.GetGlucoseStats)
		
		api.GET("/food/:user_id", handler.GetFoodRecords)
		api.POST("/food", handler.CreateFoodRecord)
		api.PUT("/food/:id", handler.UpdateFoodRecord)
		api.DELETE("/food/:id", handler.DeleteFoodRecord)
	}
	
	return router, handler, db
}

func TestAPIHandler_GetUser(t *testing.T) {
	router, _, db := setupTestRouter()
	defer testutils.CleanupTestDB(db)

	t.Run("UserExists", func(t *testing.T) {
		// Создаем тестового пользователя
		user := testutils.CreateTestUser(db, 123456789)
		
		req := httptest.NewRequest("GET", "/api/v1/user/123456789", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.TelegramID, response.TelegramID)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/user/999999999", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "User not found", response["error"])
	})

	t.Run("InvalidTelegramID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/user/invalid", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAPIHandler_UpdateDiabetesInfo(t *testing.T) {
	router, _, db := setupTestRouter()
	defer testutils.CleanupTestDB(db)

	t.Run("ValidUpdate", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		
		updateData := map[string]interface{}{
			"diabetes_type":   2,
			"target_glucose": 6.5,
		}
		
		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/api/v1/user/123456789/diabetes-info", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Проверяем, что данные обновлены
		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.Equal(t, 2, *updatedUser.DiabetesType)
		assert.Equal(t, 6.5, *updatedUser.TargetGlucose)
	})

	t.Run("InvalidDiabetesType", func(t *testing.T) {
		testutils.CreateTestUser(db, 123456789)
		
		updateData := map[string]interface{}{
			"diabetes_type":   3, // некорректный тип
			"target_glucose": 6.5,
		}
		
		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/api/v1/user/123456789/diabetes-info", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		updateData := map[string]interface{}{
			"diabetes_type":   1,
			"target_glucose": 6.0,
		}
		
		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/api/v1/user/999999999/diabetes-info", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAPIHandler_CreateGlucoseRecord(t *testing.T) {
	router, _, db := setupTestRouter()
	defer testutils.CleanupTestDB(db)

	t.Run("ValidRecord", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		
		recordData := map[string]interface{}{
			"user_id": user.ID,
			"value":   6.5,
			"notes":   "После завтрака",
		}
		
		body, _ := json.Marshal(recordData)
		req := httptest.NewRequest("POST", "/api/v1/glucose", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response models.GlucoseRecord
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, user.ID, response.UserID)
		assert.Equal(t, 6.5, response.Value)
		assert.Equal(t, "После завтрака", response.Notes)
		assert.NotZero(t, response.ID)
	})

	t.Run("InvalidValue", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		
		recordData := map[string]interface{}{
			"user_id": user.ID,
			"value":   50.0, // слишком высокое значение
		}
		
		body, _ := json.Marshal(recordData)
		req := httptest.NewRequest("POST", "/api/v1/glucose", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MissingUserID", func(t *testing.T) {
		recordData := map[string]interface{}{
			"value": 6.5,
		}
		
		body, _ := json.Marshal(recordData)
		req := httptest.NewRequest("POST", "/api/v1/glucose", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAPIHandler_GetGlucoseRecords(t *testing.T) {
	router, _, db := setupTestRouter()
	defer testutils.CleanupTestDB(db)

	t.Run("GetUserRecords", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		
		// Создаем несколько записей
		testutils.CreateTestGlucoseRecord(db, user.ID, 6.0)
		testutils.CreateTestGlucoseRecord(db, user.ID, 6.5)
		testutils.CreateTestGlucoseRecord(db, user.ID, 7.0)
		
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/glucose/%d", user.ID), nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response []models.GlucoseRecord
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Len(t, response, 3)
		for _, record := range response {
			assert.Equal(t, user.ID, record.UserID)
		}
	})

	t.Run("WithDaysFilter", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 987654321)
		testutils.CreateTestGlucoseRecord(db, user.ID, 6.0)
		
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/glucose/%d?days=7", user.ID), nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/glucose/invalid", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAPIHandler_GetGlucoseStats(t *testing.T) {
	router, _, db := setupTestRouter()
	defer testutils.CleanupTestDB(db)

	t.Run("CalculateStats", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		
		// Создаем записи с известными значениями
		testutils.CreateTestGlucoseRecord(db, user.ID, 5.0)
		testutils.CreateTestGlucoseRecord(db, user.ID, 6.0)
		testutils.CreateTestGlucoseRecord(db, user.ID, 7.0)
		testutils.CreateTestGlucoseRecord(db, user.ID, 8.0)
		
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/glucose/%d/stats", user.ID), nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, float64(4), response["count"])
		assert.Equal(t, 6.5, response["average"]) // (5+6+7+8)/4
		assert.Equal(t, 5.0, response["min"])
		assert.Equal(t, 8.0, response["max"])
	})
}

func TestAPIHandler_CreateFoodRecord(t *testing.T) {
	router, _, db := setupTestRouter()
	defer testutils.CleanupTestDB(db)

	t.Run("ValidRecord", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		
		recordData := map[string]interface{}{
			"user_id":   user.ID,
			"food_name": "Овсянка с ягодами",
			"food_type": "завтрак",
			"carbs":     45.5,
			"calories":  280,
			"quantity":  "1 порция",
			"notes":     "Без сахара",
		}
		
		body, _ := json.Marshal(recordData)
		req := httptest.NewRequest("POST", "/api/v1/food", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response models.FoodRecord
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, user.ID, response.UserID)
		assert.Equal(t, "Овсянка с ягодами", response.FoodName)
		assert.Equal(t, "завтрак", response.FoodType)
		assert.NotNil(t, response.Carbs)
		assert.Equal(t, 45.5, *response.Carbs)
		assert.NotNil(t, response.Calories)
		assert.Equal(t, 280, *response.Calories)
	})

	t.Run("MinimalRecord", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456790)
		
		recordData := map[string]interface{}{
			"user_id":   user.ID,
			"food_name": "Яблоко",
			"food_type": "перекус",
		}
		
		body, _ := json.Marshal(recordData)
		req := httptest.NewRequest("POST", "/api/v1/food", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusCreated {
			t.Logf("Response body: %s", w.Body.String())
			t.Logf("Request body: %s", string(body))
		}
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("MissingFoodName", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456791)
		
		recordData := map[string]interface{}{
			"user_id":   user.ID,
			"food_type": "завтрак",
		}
		
		body, _ := json.Marshal(recordData)
		req := httptest.NewRequest("POST", "/api/v1/food", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAPIHandler_GetFoodRecords(t *testing.T) {
	router, _, db := setupTestRouter()
	defer testutils.CleanupTestDB(db)

	t.Run("GetAllUserRecords", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		
		testutils.CreateTestFoodRecord(db, user.ID, "Завтрак", "завтрак")
		testutils.CreateTestFoodRecord(db, user.ID, "Обед", "обед")
		testutils.CreateTestFoodRecord(db, user.ID, "Ужин", "ужин")
		
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/food/%d", user.ID), nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response []models.FoodRecord
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Len(t, response, 3)
	})

	t.Run("FilterByFoodType", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 987654321)
		
		testutils.CreateTestFoodRecord(db, user.ID, "Завтрак 1", "завтрак")
		testutils.CreateTestFoodRecord(db, user.ID, "Завтрак 2", "завтрак")
		testutils.CreateTestFoodRecord(db, user.ID, "Обед", "обед")
		
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/food/%d?type=завтрак", user.ID), nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response []models.FoodRecord
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Len(t, response, 2)
		for _, record := range response {
			assert.Equal(t, "завтрак", record.FoodType)
		}
	})
}

func TestAPIHandler_UpdateGlucoseRecord(t *testing.T) {
	router, _, db := setupTestRouter()
	defer testutils.CleanupTestDB(db)

	t.Run("ValidUpdate", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		record := testutils.CreateTestGlucoseRecord(db, user.ID, 6.0)
		
		updateData := map[string]interface{}{
			"user_id": user.ID,
			"value":   7.2,
			"notes":   "Исправленное значение",
		}
		
		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/glucose/%d", record.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAPIHandler_DeleteGlucoseRecord(t *testing.T) {
	router, _, db := setupTestRouter()
	defer testutils.CleanupTestDB(db)

	t.Run("ValidDelete", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		record := testutils.CreateTestGlucoseRecord(db, user.ID, 6.0)
		
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/glucose/%d?user_id=%d", record.ID, user.ID), nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("MissingUserID", func(t *testing.T) {
		user := testutils.CreateTestUser(db, 123456789)
		record := testutils.CreateTestGlucoseRecord(db, user.ID, 6.0)
		
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/glucose/%d", record.ID), nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}