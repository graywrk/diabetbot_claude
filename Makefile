.PHONY: help build run test clean docker-build docker-up docker-down migrate frontend

# Variables
BINARY_NAME=diabetbot
DOCKER_IMAGE=diabetbot:latest
DOCKER_COMPOSE_FILE=docker-compose.yml

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@egrep '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go binary
	@echo "Building $(BINARY_NAME)..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(BINARY_NAME) cmd/main.go

run: ## Run the application locally
	@echo "Running $(BINARY_NAME)..."
	@go run cmd/main.go

test: ## Run all tests
	@echo "Running all tests..."
	@$(MAKE) test-backend
	@$(MAKE) test-frontend

test-backend: ## Run backend tests
	@echo "Running backend tests..."
	@go test -v ./...

test-frontend: ## Run frontend tests
	@echo "Running frontend tests..."
	@cd web && npm test

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(MAKE) test-coverage-backend
	@$(MAKE) test-coverage-frontend

test-coverage-backend: ## Run backend tests with coverage
	@echo "Running backend tests with coverage..."
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-frontend: ## Run frontend tests with coverage
	@echo "Running frontend tests with coverage..."
	@cd web && npm run test:coverage

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -tags=integration ./...

test-watch: ## Run tests in watch mode
	@echo "Running tests in watch mode..."
	@cd web && npm run test:watch

test-ci: ## Run tests for CI environment
	@echo "Running CI tests..."
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@cd web && npm run test:coverage

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf web/dist/
	@rm -rf web/node_modules/

frontend: ## Build frontend
	@echo "Building frontend..."
	@cd web && npm install && npm run build

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-up: ## Start services with Docker Compose
	@echo "Starting services..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-down: ## Stop services
	@echo "Stopping services..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-logs: ## Show Docker logs
	@docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

migrate: ## Run database migrations (auto-migrates on startup)
	@echo "Migrations run automatically on application startup"

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Please edit .env file with your configuration"

prod-deploy: ## Deploy to production
	@echo "Deploying to production..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up -d --build

webhook: ## Set Telegram webhook (requires TELEGRAM_BOT_TOKEN and WEBHOOK_URL env vars)
	@echo "Setting Telegram webhook..."
	@curl -X POST "https://api.telegram.org/bot$(TELEGRAM_BOT_TOKEN)/setWebhook" \
		-H "Content-Type: application/json" \
		-d '{"url":"$(WEBHOOK_URL)/webhook"}'

webhook-info: ## Get current webhook info
	@curl "https://api.telegram.org/bot$(TELEGRAM_BOT_TOKEN)/getWebhookInfo"

lint-go: ## Lint Go code
	@echo "Linting Go code..."
	@golangci-lint run

lint-frontend: ## Lint frontend code
	@echo "Linting frontend..."
	@cd web && npm run lint

install-tools: ## Install development tools
	@echo "Installing tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest