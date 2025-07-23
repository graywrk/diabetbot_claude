package handlers

import (
	"net/http"
	"strconv"

	"diabetbot/internal/models"
	"diabetbot/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type APIHandler struct {
	userService    *services.UserService
	glucoseService *services.GlucoseService
	foodService    *services.FoodService
}

func NewAPIHandler(db *gorm.DB) *APIHandler {
	return &APIHandler{
		userService:    services.NewUserService(db),
		glucoseService: services.NewGlucoseService(db),
		foodService:    services.NewFoodService(db),
	}
}

// User endpoints
func (h *APIHandler) GetUser(c *gin.Context) {
	telegramIDStr := c.Param("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telegram_id"})
		return
	}

	user, err := h.userService.GetByTelegramID(telegramID)
	if err != nil {
		// Если пользователь не найден, создаем его из данных Telegram WebApp
		if err == gorm.ErrRecordNotFound {
			// Получаем данные из заголовков для создания пользователя
			username := c.GetHeader("X-Telegram-Username")
			firstName := c.GetHeader("X-Telegram-First-Name") 
			lastName := c.GetHeader("X-Telegram-Last-Name")
			languageCode := c.GetHeader("X-Telegram-Language-Code")
			
			// Создаем нового пользователя
			user, err = h.userService.GetOrCreateUser(telegramID, username, firstName, lastName, languageCode)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
			
			c.JSON(http.StatusCreated, user)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *APIHandler) UpdateDiabetesInfo(c *gin.Context) {
	telegramIDStr := c.Param("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telegram_id"})
		return
	}

	user, err := h.userService.GetByTelegramID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		DiabetesType  int     `json:"diabetes_type" binding:"required,min=1,max=2"`
		TargetGlucose float64 `json:"target_glucose" binding:"required,min=3,max=15"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateDiabetesInfo(user.ID, req.DiabetesType, req.TargetGlucose); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Diabetes info updated successfully"})
}

func (h *APIHandler) UpdateUser(c *gin.Context) {
	telegramIDStr := c.Param("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telegram_id"})
		return
	}

	user, err := h.userService.GetByTelegramID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		TargetGlucose *float64 `json:"target_glucose"`
		Notifications *bool    `json:"notifications"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем только переданные поля
	updates := make(map[string]interface{})
	if req.TargetGlucose != nil {
		updates["target_glucose"] = req.TargetGlucose
	}
	if req.Notifications != nil {
		updates["notifications"] = req.Notifications
	}

	if err := h.userService.UpdateUser(user.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Получаем обновленного пользователя
	updatedUser, err := h.userService.GetByTelegramID(telegramID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user"})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func (h *APIHandler) DeleteUserData(c *gin.Context) {
	telegramIDStr := c.Param("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telegram_id"})
		return
	}

	user, err := h.userService.GetByTelegramID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Удаляем все данные пользователя (glucose records, food records)
	if err := h.glucoseService.DeleteAllUserRecords(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete glucose records"})
		return
	}

	if err := h.foodService.DeleteAllUserRecords(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete food records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User data deleted successfully"})
}

// Glucose endpoints
func (h *APIHandler) GetGlucoseRecords(c *gin.Context) {
	telegramID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	// Получаем пользователя по telegram_id
	user, err := h.userService.GetByTelegramID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	days := 30 // по умолчанию 30 дней
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	records, err := h.glucoseService.GetUserRecords(user.ID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get glucose records"})
		return
	}

	c.JSON(http.StatusOK, records)
}

func (h *APIHandler) CreateGlucoseRecord(c *gin.Context) {
	var req struct {
		UserID int64   `json:"user_id" binding:"required"`
		Value  float64 `json:"value" binding:"required,min=1,max=30"`
		Notes  string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем пользователя по telegram_id
	user, err := h.userService.GetByTelegramID(req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	record, err := h.glucoseService.CreateRecord(user.ID, req.Value, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create glucose record"})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (h *APIHandler) UpdateGlucoseRecord(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	var req struct {
		UserID uint    `json:"user_id" binding:"required"`
		Value  float64 `json:"value" binding:"required,min=1,max=30"`
		Notes  string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.glucoseService.UpdateRecord(req.UserID, uint(recordID), req.Value, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update glucose record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Glucose record updated successfully"})
}

func (h *APIHandler) DeleteGlucoseRecord(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid user_id"})
		return
	}

	if err := h.glucoseService.DeleteRecord(uint(userID), uint(recordID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete glucose record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Glucose record deleted successfully"})
}

func (h *APIHandler) GetGlucoseStats(c *gin.Context) {
	telegramID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	// Получаем пользователя по telegram_id
	user, err := h.userService.GetByTelegramID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	stats, err := h.glucoseService.GetUserStats(user.ID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get glucose stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Food endpoints
func (h *APIHandler) GetFoodRecords(c *gin.Context) {
	telegramID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	// Получаем пользователя по telegram_id
	user, err := h.userService.GetByTelegramID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	var records []models.FoodRecord
	var serviceErr error

	if foodType := c.Query("type"); foodType != "" {
		records, serviceErr = h.foodService.GetRecordsByType(user.ID, foodType, days)
	} else {
		records, serviceErr = h.foodService.GetUserRecords(user.ID, days)
	}

	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get food records"})
		return
	}

	c.JSON(http.StatusOK, records)
}

func (h *APIHandler) CreateFoodRecord(c *gin.Context) {
	var req struct {
		UserID   int64    `json:"user_id" binding:"required"`
		FoodName string   `json:"food_name" binding:"required"`
		FoodType string   `json:"food_type" binding:"required"`
		Carbs    *float64 `json:"carbs"`
		Calories *int     `json:"calories"`
		Quantity string   `json:"quantity"`
		Notes    string   `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем пользователя по telegram_id
	user, err := h.userService.GetByTelegramID(req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	record, err := h.foodService.CreateRecord(
		user.ID, req.FoodName, req.FoodType,
		req.Carbs, req.Calories, req.Quantity, req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create food record"})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (h *APIHandler) UpdateFoodRecord(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	var req struct {
		UserID   uint     `json:"user_id" binding:"required"`
		FoodName string   `json:"food_name"`
		FoodType string   `json:"food_type"`
		Carbs    *float64 `json:"carbs"`
		Calories *int     `json:"calories"`
		Quantity string   `json:"quantity"`
		Notes    string   `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.FoodName != "" {
		updates["food_name"] = req.FoodName
	}
	if req.FoodType != "" {
		updates["food_type"] = req.FoodType
	}
	if req.Carbs != nil {
		updates["carbs"] = req.Carbs
	}
	if req.Calories != nil {
		updates["calories"] = req.Calories
	}
	if req.Quantity != "" {
		updates["quantity"] = req.Quantity
	}
	if req.Notes != "" {
		updates["notes"] = req.Notes
	}

	if err := h.foodService.UpdateRecord(req.UserID, uint(recordID), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update food record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Food record updated successfully"})
}

func (h *APIHandler) DeleteFoodRecord(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid user_id"})
		return
	}

	if err := h.foodService.DeleteRecord(uint(userID), uint(recordID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete food record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Food record deleted successfully"})
}