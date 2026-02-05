.PHONY: help dev build run install clean docker-build docker-run docker-stop web-dev web-build server-build server-run

# Переменные
GO_CMD = go
BUN_CMD = bun
DOCKER_CMD = docker
COMPOSE_CMD = docker-compose
PORT ?= 8080
BINARY_NAME = server
WEB_DIR = web
BUILD_DIR = build

# Цвета для вывода
GREEN = \033[0;32m
YELLOW = \033[0;33m
NC = \033[0m # No Color

help: ## Показать справку по командам
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

install: ## Установить все зависимости (Go и Web)
	@echo "$(GREEN)Установка зависимостей Go...$(NC)"
	$(GO_CMD) mod download
	@echo "$(GREEN)Установка зависимостей Web...$(NC)"
	cd $(WEB_DIR) && $(BUN_CMD) install

dev: ## Запустить проект в режиме разработки (frontend + backend)
	@echo "$(GREEN)Запуск в режиме разработки...$(NC)"
	@echo "$(YELLOW)Frontend: http://localhost:5173$(NC)"
	@echo "$(YELLOW)Backend: http://localhost:$(PORT)$(NC)"
	@make -j2 web-dev server-run

web-dev: ## Запустить только frontend в режиме разработки
	@echo "$(GREEN)Запуск frontend dev сервера...$(NC)"
	cd $(WEB_DIR) && $(BUN_CMD) run dev

web-build: ## Собрать frontend
	@echo "$(GREEN)Сборка frontend...$(NC)"
	cd $(WEB_DIR) && $(BUN_CMD) run build

server-build: ## Собрать backend
	@echo "$(GREEN)Сборка backend...$(NC)"
	CGO_ENABLED=0 $(GO_CMD) build -o $(BINARY_NAME) ./cmd/server/main.go

server-run: ## Запустить только backend сервер
	@echo "$(GREEN)Запуск backend сервера на порту $(PORT)...$(NC)"
	PORT=$(PORT) $(GO_CMD) run ./cmd/server/main.go

build: web-build server-build ## Собрать весь проект (frontend + backend)
	@echo "$(GREEN)Сборка проекта завершена!$(NC)"

run: build ## Запустить собранный проект
	@echo "$(GREEN)Запуск проекта на порту $(PORT)...$(NC)"
	PORT=$(PORT) ./$(BINARY_NAME)

docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Сборка Docker образа...$(NC)"
	$(DOCKER_CMD) build -f $(BUILD_DIR)/Dockerfile -t avt0x/docker-dashboard:latest .

docker-run: ## Запустить проект через Docker Compose
	@echo "$(GREEN)Запуск через Docker Compose...$(NC)"
	$(COMPOSE_CMD) up -d

docker-stop: ## Остановить Docker Compose
	@echo "$(GREEN)Остановка Docker Compose...$(NC)"
	$(COMPOSE_CMD) down

docker-logs: ## Показать логи Docker Compose
	$(COMPOSE_CMD) logs -f

clean: ## Очистить артефакты сборки
	@echo "$(GREEN)Очистка артефактов...$(NC)"
	rm -f $(BINARY_NAME)
	rm -rf $(WEB_DIR)/node_modules
	rm -rf $(WEB_DIR)/dist
	rm -rf $(WEB_DIR)/public/assets
	$(GO_CMD) clean

test: ## Запустить тесты
	@echo "$(GREEN)Запуск тестов...$(NC)"
	$(GO_CMD) test ./...

fmt: ## Форматировать Go код
	@echo "$(GREEN)Форматирование Go кода...$(NC)"
	$(GO_CMD) fmt ./...

