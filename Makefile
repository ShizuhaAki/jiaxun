# Makefile for project with backend and frontend
# Friendly interface with colorized output and help display

# Variables
GOPATH := $(shell go env GOPATH)
BACKEND := jiaxun
FRONTEND := jiaxun-frontend
SWAG := $(GOPATH)/bin/swag

# Colors
GREEN := \033[0;32m
YELLOW := \033[1;33m
RESET := \033[0m

# Default: show help
.DEFAULT_GOAL := help

.PHONY: check-deps

check-deps: ## Check for required global dependencies
	@command -v go >/dev/null 2>&1 || { echo "$(YELLOW)[ERROR] Go is not installed.$(RESET)"; exit 1; }
	@go version | grep -q 'go1\.\(1[89]\|[2-9][0-9]\)' || { echo "$(YELLOW)[ERROR] Go version 1.18 or higher required.$(RESET)"; exit 1; }
	@command -v pnpm >/dev/null 2>&1 || { echo "$(YELLOW)[ERROR] pnpm is not installed.$(RESET)"; exit 1; }
	@echo "$(GREEN)All required dependencies are installed.$(RESET)"

# Help system
help:
	@echo "$(YELLOW)Available targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""

prepare: ## Generate Swagger spec and frontend client
	@echo "$(YELLOW)==> Generating Swagger documentation...$(RESET)"
	cd $(BACKEND) && $(SWAG) init -g cmd/server/main.go
	@echo "$(YELLOW)==> Generating frontend client code...$(RESET)"
	cd $(FRONTEND) && pnpm run generate:client

install-dependencies: ## Install frontend/backend dependencies
	@echo "$(YELLOW)==> Installing frontend dependencies...$(RESET)"
	cd $(FRONTEND) && pnpm install
	@echo "$(YELLOW)==> Installing backend dependencies...$(RESET)"
	cd $(BACKEND) && go install github.com/swaggo/swag/v2/cmd/swag@latest

backend: ## Run backend server
	@echo "$(YELLOW)==> Starting backend server...$(RESET)"
	cd $(BACKEND) && go run cmd/server/main.go

frontend: prepare ## Run frontend dev server
	@echo "$(YELLOW)==> Starting frontend development server...$(RESET)"
	cd $(FRONTEND) && pnpm run dev

dev: prepare ## Start full development environment (prepare + parallel backend/frontend)
	@echo "$(YELLOW)==> Preparing project...$(RESET)"
	$(MAKE) backend &
	$(MAKE) frontend

