.PHONY: help build run stop restart logs clean test dev prod monitoring-up monitoring-down monitoring-logs metrics prometheus grafana

# Detect if using podman or docker
DOCKER := $(shell command -v podman 2> /dev/null || command -v docker 2> /dev/null)
COMPOSE := $(shell command -v podman-compose 2> /dev/null || command -v docker-compose 2> /dev/null)

# Colors for output
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
NC     := \033[0m # No Color

help: ## Show this help message
	@echo '$(GREEN)E-Learning Platform - Makefile Commands$(NC)'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# Development
dev: ## Run in development mode
	@echo "$(GREEN)Starting development environment...$(NC)"
	go run cmd/api/main.go

build: ## Build the application
	@echo "$(GREEN)Building application...$(NC)"
	go build -o bin/api cmd/api/main.go

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v ./...

# Docker/Podman commands
up: ## Start all services (backend + monitoring)
	@echo "$(GREEN)Starting all services with $(COMPOSE)...$(NC)"
	$(COMPOSE) up -d
	@echo "$(GREEN)Services started!$(NC)"
	@echo "$(YELLOW)Backend API:    http://localhost:8080$(NC)"
	@echo "$(YELLOW)Prometheus:     http://localhost:9090$(NC)"
	@echo "$(YELLOW)Grafana:        http://localhost:3001 (admin/admin123)$(NC)"
	@echo "$(YELLOW)PostgreSQL:     localhost:5432$(NC)"

down: ## Stop all services
	@echo "$(RED)Stopping all services...$(NC)"
	$(COMPOSE) down

stop: down ## Alias for down

restart: ## Restart all services
	@echo "$(YELLOW)Restarting services...$(NC)"
	$(COMPOSE) restart

logs: ## Show logs from all services
	$(COMPOSE) logs -f

logs-backend: ## Show backend logs
	$(COMPOSE) logs -f backend

logs-db: ## Show database logs
	$(COMPOSE) logs -f postgres

logs-notification: ## Show notification service logs
	$(COMPOSE) logs -f notification-service

# Monitoring specific commands
monitoring-up: ## Start only monitoring services (Prometheus + Grafana)
	@echo "$(GREEN)Starting monitoring services...$(NC)"
	$(COMPOSE) up -d prometheus grafana node-exporter
	@echo "$(GREEN)Monitoring started!$(NC)"
	@echo "$(YELLOW)Prometheus: http://localhost:9090$(NC)"
	@echo "$(YELLOW)Grafana:    http://localhost:3001 (admin/admin123)$(NC)"

monitoring-down: ## Stop monitoring services
	@echo "$(RED)Stopping monitoring services...$(NC)"
	$(COMPOSE) stop prometheus grafana node-exporter

monitoring-restart: ## Restart monitoring services
	@echo "$(YELLOW)Restarting monitoring services...$(NC)"
	$(COMPOSE) restart prometheus grafana node-exporter

monitoring-logs: ## Show monitoring logs
	$(COMPOSE) logs -f prometheus grafana

prometheus: ## Open Prometheus UI
	@echo "$(GREEN)Opening Prometheus UI...$(NC)"
	@echo "$(YELLOW)http://localhost:9090$(NC)"
	@command -v xdg-open > /dev/null && xdg-open http://localhost:9090 || echo "Open http://localhost:9090 in your browser"

grafana: ## Open Grafana UI
	@echo "$(GREEN)Opening Grafana UI...$(NC)"
	@echo "$(YELLOW)http://localhost:3001 (admin/admin123)$(NC)"
	@command -v xdg-open > /dev/null && xdg-open http://localhost:3001 || echo "Open http://localhost:3001 in your browser"

metrics: ## Check metrics endpoint
	@echo "$(GREEN)Fetching metrics from backend...$(NC)"
	@curl -s http://localhost:8080/metrics | head -n 50
	@echo ""
	@echo "$(YELLOW)Full metrics: http://localhost:8080/metrics$(NC)"

# Backend specific commands
backend-up: ## Start only backend services (without monitoring)
	@echo "$(GREEN)Starting backend services...$(NC)"
	$(COMPOSE) up -d postgres notification-service backend

backend-restart: ## Restart backend
	$(COMPOSE) restart backend

backend-rebuild: ## Rebuild and restart backend
	@echo "$(YELLOW)Rebuilding backend...$(NC)"
	$(COMPOSE) build backend
	$(COMPOSE) up -d backend

# Database commands
db-shell: ## Connect to PostgreSQL shell
	$(DOCKER) exec -it elearning-postgres psql -U postgres -d elearning

db-migrate: ## Run database migrations (if you have migrations)
	@echo "$(YELLOW)Running migrations...$(NC)"
	# Add your migration command here

db-backup: ## Backup database
	@echo "$(GREEN)Backing up database...$(NC)"
	$(DOCKER) exec elearning-postgres pg_dump -U postgres elearning > backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Backup completed!$(NC)"

# Cleanup commands
clean: ## Stop and remove all containers, networks, and volumes
	@echo "$(RED)Cleaning up...$(NC)"
	$(COMPOSE) down -v
	rm -rf bin/

clean-logs: ## Clean log files
	@echo "$(YELLOW)Cleaning logs...$(NC)"
	rm -rf logs/*.log

prune: ## Remove unused Docker/Podman resources
	@echo "$(RED)Pruning unused resources...$(NC)"
	$(DOCKER) system prune -af --volumes

# Production commands
prod: ## Start in production mode
	@echo "$(GREEN)Starting in production mode...$(NC)"
	GIN_MODE=release $(COMPOSE) up -d

prod-build: ## Build for production
	@echo "$(GREEN)Building for production...$(NC)"
	$(COMPOSE) build --no-cache

# Health checks
health: ## Check service health
	@echo "$(GREEN)Checking service health...$(NC)"
	@echo -n "Backend:        "
	@curl -sf http://localhost:8080/health && echo "$(GREEN)✓ OK$(NC)" || echo "$(RED)✗ DOWN$(NC)"
	@echo -n "Prometheus:     "
	@curl -sf http://localhost:9090/-/healthy && echo "$(GREEN)✓ OK$(NC)" || echo "$(RED)✗ DOWN$(NC)"
	@echo -n "Grafana:        "
	@curl -sf http://localhost:3001/api/health && echo "$(GREEN)✓ OK$(NC)" || echo "$(RED)✗ DOWN$(NC)"

status: ## Show running services
	@echo "$(GREEN)Running services:$(NC)"
	$(COMPOSE) ps

# Quick commands
quick-start: ## Quick start (pull, build, up)
	@echo "$(GREEN)Quick starting...$(NC)"
	$(COMPOSE) pull
	$(COMPOSE) build
	$(COMPOSE) up -d
	@make health

quick-restart: down up ## Quick restart (down + up)

# Development helpers
watch: ## Watch and rebuild on changes (requires entr or similar)
	@echo "$(YELLOW)Watching for changes...$(NC)"
	ls **/*.go | entr -r make dev

install-deps: ## Install Go dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	go mod download
	go mod tidy

# Notification service
notification-logs: ## Show notification service logs
	$(COMPOSE) logs -f notification-service

notification-restart: ## Restart notification service
	$(COMPOSE) restart notification-service