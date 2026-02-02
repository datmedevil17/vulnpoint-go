.PHONY: all help build run-server run-web test clean docker-up docker-down setup

# Configuration
APP_NAME=go-vuln
BUILD_DIR=./bin
MAIN_PATH=./cmd/server

# Default target
all: build

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# --- Application ---

build: ## Build the backend binary
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

run-server: ## Run the Go backend
	go run $(MAIN_PATH)/main.go

run-web: ## Run the React frontend (requires separate terminal)
	cd web && npm run dev

test: ## Run backend tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR) tmp coverage.out

# --- Infrastructure ---

docker-up: ## Start database & redis
	docker compose -f docker/docker-compose.yml up -d

docker-down: ## Stop database & redis
	docker compose -f docker/docker-compose.yml down

# --- Setup ---

deps: ## Update Go dependencies
	go mod tidy && go mod download

setup: docker-up deps ## Initialize project (start DB, install deps)
	@echo "Setup complete! Run 'make run-server' and 'make run-web' to start."
