# Makefile for Subscription Service

DOCKER_COMPOSE = docker-compose
GO = go
APP_NAME = subscriptions-app
DB_NAME = subscriptions-postgres
IMAGE_NAME = subservice
VERSION ?= latest

## Help
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
lint: ## Run linter
	golangci-lint run ./... --config=./.golangci.yml

test: ## Run tests
	$(GO) test -v ./...

test-coverage: ## Run tests with coverage
	$(GO) test ./... -cover

##@ Docker Operations
build: ## Build Docker image
	docker build -t $(IMAGE_NAME):$(VERSION) .

start: ## Start all services with Docker Compose
	$(DOCKER_COMPOSE) up -d

stop: ## Stop and remove all containers
	$(DOCKER_COMPOSE) down

restart: ## Restart all services
	$(DOCKER_COMPOSE) restart

status: ## Show status of all containers
	@echo "$(GREEN)Container status:$(NC)"
	$(DOCKER_COMPOSE) ps

##@ Health & Monitoring
health: ## Check health of services
	@echo "$(GREEN)Checking service health...$(NC)"
	@curl -f http://localhost:8080/api/v1/subscriptions?page=1&page_size=1 > /dev/null 2>&1 && echo "$(GREEN)✓ API is healthy$(NC)" || echo "$(RED)✗ API is not responding$(NC)"
	@docker exec $(DB_NAME) pg_isready -U user -d subscriptionsdb > /dev/null 2>&1 && echo "$(GREEN)✓ Database is healthy$(NC)" || echo "$(RED)✗ Database is not responding$(NC)"

##@ Development Tools
swagger: ## Open Swagger documentation
	@which open > /dev/null && open http://localhost:8080/swagger/index.html || echo "Open http://localhost:8080/swagger/index.html in your browser"

##@ Information
version: ## Show application version
	@echo "Application: $(IMAGE_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Go version: $(shell $(GO) version)"
	@echo "Docker Compose version: $(shell $(DOCKER_COMPOSE) version --short)"

ports: ## Show used ports
	@echo "  Application: http://localhost:8080"
	@echo "  Database: localhost:5432"
	@echo "  Swagger UI: http://localhost:8080/swagger/index.html"
