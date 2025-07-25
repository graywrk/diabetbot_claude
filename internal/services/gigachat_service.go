package services

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"diabetbot/internal/config"
	"diabetbot/internal/models"
)

type GigaChatService struct {
	apiKey    string
	baseURL   string
	client    *http.Client
	authToken string
	tokenExp  time.Time
}

type AuthRequest struct {
	Scope string `json:"scope"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type ChatRequest struct {
	Model             string    `json:"model"`
	Messages          []Message `json:"messages"`
	Temperature       float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	N                int       `json:"n,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	RepetitionPenalty float64  `json:"repetition_penalty,omitempty"`
	UpdateInterval   int       `json:"update_interval,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Object  string   `json:"object"`
}

type Choice struct {
	Message      Message `json:"message"`
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewGigaChatService(cfg *config.GigaChatConfig) *GigaChatService {
	return &GigaChatService{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 30 * time.Second,
		},
	}
}

func (s *GigaChatService) authenticate() error {
	if time.Now().Before(s.tokenExp) && s.authToken != "" {
		return nil // токен еще действителен
	}

	// Используем локальный прокси для обхода Docker ограничений  
	authURL := "http://172.17.0.1:8888/oauth"
	formData := "scope=GIGACHAT_API_PERS"
	log.Printf("GigaChat auth URL (via proxy): %s", authURL)
	log.Printf("GigaChat auth data: %s", formData)
	log.Printf("GigaChat API key (first 20 chars): %s", s.apiKey[:20])
	
	req, err := http.NewRequest("POST", authURL, strings.NewReader(formData))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+s.apiKey)
	req.Header.Set("RqUID", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("User-Agent", "DiabetBot/1.0")
	
	log.Printf("Request headers: %+v", req.Header)

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Failed to send auth request: %v", err)
		return fmt.Errorf("failed to send auth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	log.Printf("Auth response status: %d", resp.StatusCode)
	log.Printf("Auth response headers: %+v", resp.Header)
	log.Printf("Auth response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to unmarshal auth response: %w", err)
	}

	s.authToken = authResp.AccessToken
	s.tokenExp = time.Now().Add(time.Duration(authResp.ExpiresIn-60) * time.Second) // обновляем за минуту до истечения

	return nil
}

func (s *GigaChatService) sendChatRequest(messages []Message) (string, error) {
	if err := s.authenticate(); err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	chatReq := ChatRequest{
		Model:             "GigaChat:latest",
		Messages:          messages,
		Temperature:       0.7,
		TopP:             0.9,
		N:                1,
		Stream:           false,
		MaxTokens:        1000,
		RepetitionPenalty: 1.0,
	}

	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal chat request: %w", err)
	}

	req, err := http.NewRequest("POST", s.baseURL+"/api/v2/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create chat request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.authToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send chat request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read chat response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("chat request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal chat response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in chat response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (s *GigaChatService) GetGlucoseRecommendation(user *models.User, record *models.GlucoseRecord) string {
	if s.apiKey == "" || s.apiKey == "your_gigachat_api_key_here" {
		return "Рекомендации ИИ временно недоступны (не настроен API ключ). Обратитесь к врачу для консультации."
	}

	diabetesTypeText := "не указан"
	if user.DiabetesType != nil {
		diabetesTypeText = fmt.Sprintf("%d типа", *user.DiabetesType)
	}

	targetText := "не указана"
	if user.TargetGlucose != nil {
		targetText = fmt.Sprintf("%.1f ммоль/л", *user.TargetGlucose)
	}

	systemPrompt := `Ты медицинский консультант-диабетолог. Дай короткую рекомендацию (до 150 слов) по показателю глюкозы крови. 
Учитывай: норма натощак 3.9-5.5 ммоль/л, через 2 часа после еды до 7.8 ммоль/л. 
Не ставь диагнозы, рекомендуй обращение к врачу при критических значениях.
Отвечай по-русски, дружелюбно и профессионально.`

	userPrompt := fmt.Sprintf(`Пациент:
- Диабет: %s
- Целевая глюкоза: %s
- Текущий показатель: %.1f ммоль/л
- Время измерения: %s

Дай рекомендацию по этому показателю.`, 
		diabetesTypeText, 
		targetText, 
		record.Value, 
		record.MeasuredAt.Format("15:04 02.01.2006"))

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := s.sendChatRequest(messages)
	if err != nil {
		log.Printf("GigaChat glucose recommendation error: %v", err)
		return "Не удалось получить рекомендацию от ИИ. Обратитесь к врачу для консультации."
	}

	return response
}

func (s *GigaChatService) GetFoodRecommendation(user *models.User, foodDescription string) string {
	if s.apiKey == "" || s.apiKey == "your_gigachat_api_key_here" {
		return "Рекомендации ИИ временно недоступны (не настроен API ключ). Следите за углеводами в рационе."
	}

	diabetesTypeText := "не указан"
	if user.DiabetesType != nil {
		diabetesTypeText = fmt.Sprintf("%d типа", *user.DiabetesType)
	}

	systemPrompt := `Ты диетолог, специализирующийся на диабете. Дай короткую рекомендацию (до 150 слов) по питанию.
Оцени углеводность продуктов, влияние на сахар крови, дай советы по порциям или сочетанию с другими продуктами.
Отвечай по-русски, дружелюбно и практично.`

	userPrompt := fmt.Sprintf(`Пациент с диабетом %s описал прием пищи:
"%s"

Дай рекомендацию по этой еде для контроля сахара в крови.`, diabetesTypeText, foodDescription)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := s.sendChatRequest(messages)
	if err != nil {
		log.Printf("GigaChat food recommendation error: %v", err)
		return "Не удалось получить рекомендацию от ИИ. Контролируйте количество углеводов в рационе."
	}

	return response
}

func (s *GigaChatService) GetGeneralRecommendation(user *models.User, question string) string {
	if s.apiKey == "" || s.apiKey == "your_gigachat_api_key_here" {
		return "Рекомендации ИИ временно недоступны (не настроен API ключ). Обратитесь к лечащему врачу за консультацией."
	}

	diabetesTypeText := "не указан"
	if user.DiabetesType != nil {
		diabetesTypeText = fmt.Sprintf("%d типа", *user.DiabetesType)
	}

	systemPrompt := `Ты медицинский консультант по диабету. Отвечай на вопросы о диабете, питании, физической активности.
Давай практические советы (до 200 слов). Не ставь диагнозы, при серьезных симптомах рекомендуй врача.
Отвечай по-русски, понятно и дружелюбно.`

	userPrompt := fmt.Sprintf(`Пациент с диабетом %s спрашивает:
"%s"

Дай полезный ответ по этому вопросу.`, diabetesTypeText, question)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := s.sendChatRequest(messages)
	if err != nil {
		log.Printf("GigaChat general recommendation error: %v", err)
		return "Не удалось получить ответ от ИИ. Рекомендую обратиться к лечащему врачу за консультацией."
	}

	return response
}