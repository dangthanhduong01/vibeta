.PHONY: help build run-websocket run-worker run-all stop clean kafka-topics test docker-up docker-down

# Variables
APP_NAME=vibeta
WS_BINARY=ws-server
WORKER_BINARY=message-worker

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

# Default target
help: ## Show this help message
	@echo "$(GREEN)VibeTA Chat Application - Scalable with Kafka$(NC)"
	@echo ""
	@echo "$(YELLOW)Available commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build binaries
build: ## Build WebSocket server and Worker binaries
	@echo "$(GREEN)Building binaries...$(NC)"
	go build -o bin/$(WS_BINARY) ws/main.go
	go build -o bin/$(WORKER_BINARY) cmd/worker/main.go
	@echo "$(GREEN)Build completed!$(NC)"

# Start Docker infrastructure
docker-up: ## Start Kafka, PostgreSQL, and other infrastructure
	@echo "$(GREEN)Starting Docker infrastructure...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Infrastructure started!$(NC)"
	@echo "$(YELLOW)Kafka UI: http://localhost:8090$(NC)"
	@echo "$(YELLOW)PostgreSQL: localhost:5432$(NC)"

# Stop Docker infrastructure
docker-down: ## Stop Docker infrastructure
	@echo "$(GREEN)Stopping Docker infrastructure...$(NC)"
	docker-compose down
	@echo "$(GREEN)Infrastructure stopped!$(NC)"

# Create Kafka topics
kafka-topics: ## Create necessary Kafka topics
	@echo "$(GREEN)Creating Kafka topics...$(NC)"
	docker exec vibeta-kafka kafka-topics --bootstrap-server localhost:9092 --create --topic chat_messages --partitions 3 --replication-factor 1 --if-not-exists
	docker exec vibeta-kafka kafka-topics --bootstrap-server localhost:9092 --list
	@echo "$(GREEN)Kafka topics created!$(NC)"

# Run WebSocket server
run-websocket: build ## Run WebSocket server only
	@echo "$(GREEN)Starting WebSocket server...$(NC)"
	@echo "$(YELLOW)Server will be available at: http://localhost:8080$(NC)"
	./bin/$(WS_BINARY)

# Run Message Worker
run-worker: build ## Run Message Worker only  
	@echo "$(GREEN)Starting Message Worker...$(NC)"
	./bin/$(WORKER_BINARY)

# Run both WebSocket server and Worker
run-all: build ## Run both WebSocket server and Worker in background
	@echo "$(GREEN)Starting full application...$(NC)"
	@echo "$(YELLOW)Starting Message Worker...$(NC)"
	nohup ./bin/$(WORKER_BINARY) > logs/worker.log 2>&1 &
	@sleep 2
	@echo "$(YELLOW)Starting WebSocket server...$(NC)"
	@echo "$(YELLOW)Server will be available at: http://localhost:8080$(NC)"
	./bin/$(WS_BINARY)

# Development mode with live reload
dev: ## Start development mode (requires air for live reload)
	@echo "$(GREEN)Starting development mode...$(NC)"
	@if command -v air > /dev/null 2>&1; then \
		air; \
	else \
		echo "$(RED)Air not installed. Install with: go install github.com/cosmtrek/air@latest$(NC)"; \
		echo "$(YELLOW)Starting without live reload...$(NC)"; \
		make run-websocket; \
	fi

# Stop all processes
stop: ## Stop all application processes
	@echo "$(GREEN)Stopping application processes...$(NC)"
	pkill -f $(WS_BINARY) || true
	pkill -f $(WORKER_BINARY) || true
	@echo "$(GREEN)Processes stopped!$(NC)"

# Clean build artifacts
clean: ## Clean build artifacts and logs
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -rf logs/*.log
	@echo "$(GREEN)Clean completed!$(NC)"

# Setup development environment
setup: ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@mkdir -p bin logs
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	go mod download
	@echo "$(GREEN)Setup completed!$(NC)"

# Test the application
test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	go test ./...

# Full setup and start
start: docker-up setup kafka-topics run-all ## Full setup and start (infrastructure + app)

# View logs
logs-worker: ## View worker logs
	tail -f logs/worker.log

logs-kafka: ## View Kafka logs
	docker logs -f vibeta-kafka

logs-postgres: ## View PostgreSQL logs  
	docker logs -f vibeta-postgres

# Health check
health: ## Check application health
	@echo "$(GREEN)Checking application health...$(NC)"
	@curl -s http://localhost:8080/health | jq . || echo "$(RED)WebSocket server not responding$(NC)"

# Quick restart
restart: stop run-all ## Quick restart of application

# Production deployment simulation
deploy: docker-up setup build kafka-topics ## Prepare for production deployment
	@echo "$(GREEN)Production deployment ready!$(NC)"
	@echo "$(YELLOW)Run 'make run-all' to start the application$(NC)"
