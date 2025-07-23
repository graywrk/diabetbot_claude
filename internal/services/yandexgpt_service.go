package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"diabetbot/internal/config"
	"diabetbot/internal/models"
)

type YandexGPTService struct {
	apiKey    string
	folderId  string
	client    *http.Client
}

type YandexGPTRequest struct {
	ModelURI          string                `json:"modelUri"`
	CompletionOptions YandexCompletionOptions `json:"completionOptions"`
	Messages          []YandexMessage       `json:"messages"`
}

type YandexCompletionOptions struct {
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
}

type YandexMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type YandexGPTResponse struct {
	Result YandexResult `json:"result"`
}

type YandexResult struct {
	Alternatives []YandexAlternative `json:"alternatives"`
	Usage        YandexUsage         `json:"usage"`
}

type YandexAlternative struct {
	Message YandexMessage `json:"message"`
	Status  string        `json:"status"`
}

type YandexUsage struct {
	InputTextTokens  interface{} `json:"inputTextTokens"`  // может быть string или int
	CompletionTokens interface{} `json:"completionTokens"` // может быть string или int
	TotalTokens      interface{} `json:"totalTokens"`      // может быть string или int
}

func NewYandexGPTService(cfg *config.YandexGPTConfig) *YandexGPTService {
	return &YandexGPTService{
		apiKey:   cfg.APIKey,
		folderId: cfg.FolderID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *YandexGPTService) sendRequest(messages []YandexMessage) (string, error) {
	if s.apiKey == "" || s.apiKey == "your_yandex_api_key_here" {
		return "", fmt.Errorf("YandexGPT API key not configured")
	}

	if s.folderId == "" {
		return "", fmt.Errorf("YandexGPT folder ID not configured")
	}

	modelURI := fmt.Sprintf("gpt://%s/yandexgpt-lite/latest", s.folderId)
	
	request := YandexGPTRequest{
		ModelURI: modelURI,
		CompletionOptions: YandexCompletionOptions{
			Stream:      false,
			Temperature: 0.7,
			MaxTokens:   1000,
		},
		Messages: messages,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("YandexGPT request to model: %s", modelURI)
	log.Printf("YandexGPT request data: %s", string(jsonData))

	req, err := http.NewRequest("POST", "https://llm.api.cloud.yandex.net/foundationModels/v1/completion", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+s.apiKey)

	log.Printf("YandexGPT request headers: %+v", req.Header)

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("YandexGPT request failed: %v", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("YandexGPT response status: %d", resp.StatusCode)
	log.Printf("YandexGPT response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response YandexGPTResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Result.Alternatives) == 0 {
		return "", fmt.Errorf("no alternatives in response")
	}

	return response.Result.Alternatives[0].Message.Text, nil
}

func (s *YandexGPTService) GetGlucoseRecommendation(user *models.User, record *models.GlucoseRecord) string {
	if s.apiKey == "" || s.apiKey == "your_yandex_api_key_here" {
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

	systemMessage := YandexMessage{
		Role: "system",
		Text: `Ты медицинский консультант-диабетолог. Дай короткую рекомендацию (до 150 слов) по показателю глюкозы крови. 
Учитывай: норма натощак 3.9-5.5 ммоль/л, через 2 часа после еды до 7.8 ммоль/л. 
Не ставь диагнозы, рекомендуй обращение к врачу при критических значениях.
Отвечай по-русски, дружелюбно и профессионально.`,
	}

	userMessage := YandexMessage{
		Role: "user",
		Text: fmt.Sprintf(`Пациент:
- Диабет: %s
- Целевая глюкоза: %s
- Текущий показатель: %.1f ммоль/л
- Время измерения: %s

Дай рекомендацию по этому показателю.`, 
			diabetesTypeText, 
			targetText, 
			record.Value, 
			record.MeasuredAt.Format("15:04 02.01.2006")),
	}

	messages := []YandexMessage{systemMessage, userMessage}

	response, err := s.sendRequest(messages)
	if err != nil {
		log.Printf("YandexGPT glucose recommendation error: %v", err)
		return "Не удалось получить рекомендацию от ИИ. Обратитесь к врачу для консультации."
	}

	return response
}

func (s *YandexGPTService) GetFoodRecommendation(user *models.User, foodDescription string) string {
	if s.apiKey == "" || s.apiKey == "your_yandex_api_key_here" {
		return "Рекомендации ИИ временно недоступны (не настроен API ключ). Следите за углеводами в рационе."
	}

	diabetesTypeText := "не указан"
	if user.DiabetesType != nil {
		diabetesTypeText = fmt.Sprintf("%d типа", *user.DiabetesType)
	}

	systemMessage := YandexMessage{
		Role: "system",
		Text: `Ты диетолог, специализирующийся на диабете. Дай короткую рекомендацию (до 150 слов) по питанию.
Оцени углеводность продуктов, влияние на сахар крови, дай советы по порциям или сочетанию с другими продуктами.
Отвечай по-русски, дружелюбно и практично.`,
	}

	userMessage := YandexMessage{
		Role: "user",
		Text: fmt.Sprintf(`Пациент с диабетом %s описал прием пищи:
"%s"

Дай рекомендацию по этой еде для контроля сахара в крови.`, diabetesTypeText, foodDescription),
	}

	messages := []YandexMessage{systemMessage, userMessage}

	response, err := s.sendRequest(messages)
	if err != nil {
		log.Printf("YandexGPT food recommendation error: %v", err)
		return "Не удалось получить рекомендацию от ИИ. Контролируйте количество углеводов в рационе."
	}

	return response
}

func (s *YandexGPTService) GetGeneralRecommendation(user *models.User, question string) string {
	if s.apiKey == "" || s.apiKey == "your_yandex_api_key_here" {
		return "Рекомендации ИИ временно недоступны (не настроен API ключ). Обратитесь к лечащему врачу за консультацией."
	}

	diabetesTypeText := "не указан"
	if user.DiabetesType != nil {
		diabetesTypeText = fmt.Sprintf("%d типа", *user.DiabetesType)
	}

	systemMessage := YandexMessage{
		Role: "system",
		Text: `Ты медицинский консультант по диабету. Отвечай на вопросы о диабете, питании, физической активности.
Давай практические советы (до 200 слов). Не ставь диагнозы, при серьезных симптомах рекомендуй врача.
Отвечай по-русски, понятно и дружелюбно.`,
	}

	userMessage := YandexMessage{
		Role: "user",
		Text: fmt.Sprintf(`Пациент с диабетом %s спрашивает:
"%s"

Дай полезный ответ по этому вопросу.`, diabetesTypeText, question),
	}

	messages := []YandexMessage{systemMessage, userMessage}

	response, err := s.sendRequest(messages)
	if err != nil {
		log.Printf("YandexGPT general recommendation error: %v", err)
		return "Не удалось получить ответ от ИИ. Рекомендую обратиться к лечащему врачу за консультацией."
	}

	return response
}