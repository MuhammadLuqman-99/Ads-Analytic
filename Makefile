# =============================================================================
# Ads Analytics Platform - Makefile
# =============================================================================

.PHONY: help build run test clean docker-build docker-up docker-down docker-logs \
        dev deploy logs migrate ssl-init ssl-renew \
        prod-local prod-local-down prod-local-logs prod-local-build \
        seed reset mock-error mock-slow mock-normal test-flow

# Variables
APP_NAME := ads-aggregator
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DOMAIN := $(shell grep DOMAIN .env 2>/dev/null | cut -d '=' -f2 || echo "localhost")

# Go settings
GOFLAGS := -ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Docker settings
DOCKER_COMPOSE := docker-compose
DOCKER_COMPOSE_DEV := $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml
DOCKER_COMPOSE_PROD := $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml
DOCKER_COMPOSE_LOCAL := $(DOCKER_COMPOSE) -f docker-compose.production-local.yml

# =============================================================================
# Help
# =============================================================================
help:
	@echo "Ads Analytics Platform - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make build          Build the Go backend"
	@echo "  make run            Run the API server locally"
	@echo "  make worker         Run the worker locally"
	@echo "  make lint           Run linters"
	@echo "  make fmt            Format code"
	@echo ""
	@echo "Testing:"
	@echo "  make test                Run all tests (backend + frontend)"
	@echo "  make test-backend        Run Go backend tests"
	@echo "  make test-backend-coverage  Run Go tests with coverage"
	@echo "  make test-frontend       Run frontend unit tests"
	@echo "  make test-frontend-watch Run frontend tests (watch mode)"
	@echo "  make test-frontend-coverage Run frontend tests with coverage"
	@echo "  make test-e2e            Run Playwright E2E tests"
	@echo "  make test-e2e-ui         Run E2E tests with UI"
	@echo "  make test-e2e-headed     Run E2E tests headed"
	@echo "  make test-e2e-install    Install Playwright browsers"
	@echo ""
	@echo "Docker (Development):"
	@echo "  make docker-dev     Start all services in dev mode"
	@echo "  make docker-dev-down  Stop all dev services"
	@echo "  make docker-logs    View docker logs"
	@echo ""
	@echo "Docker (Production):"
	@echo "  make docker-build   Build docker images"
	@echo "  make docker-up      Start production stack"
	@echo "  make docker-down    Stop production stack"
	@echo ""
	@echo "Production-Local Testing:"
	@echo "  make prod-local     Build and start production-like local environment"
	@echo "  make prod-local-down  Stop production-local environment"
	@echo "  make prod-local-logs  Tail all production-local service logs"
	@echo "  make seed           Seed database with test data"
	@echo "  make reset          Drop database and reseed"
	@echo "  make test-flow      Run smoke tests"
	@echo "  make mock-error     Toggle mock API to return errors"
	@echo "  make mock-slow      Toggle mock API to be slow (2-5s delay)"
	@echo "  make mock-normal    Reset mock API to normal mode"
	@echo ""
	@echo "Database:"
	@echo "  make migrate        Run database migrations"
	@echo "  make migrate-down   Rollback last migration"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean          Clean build artifacts"
	@echo "  make deps           Download dependencies"

# =============================================================================
# Development
# =============================================================================
build:
	@echo "Building $(APP_NAME)..."
	go build $(GOFLAGS) -o bin/api ./cmd/api
	go build $(GOFLAGS) -o bin/worker ./cmd/worker

run:
	@echo "Starting API server..."
	go run ./cmd/api

worker:
	@echo "Starting worker..."
	go run ./cmd/worker

test:
	@echo "Running all tests..."
	@$(MAKE) test-backend
	@$(MAKE) test-frontend

## Run Go backend tests
test-backend:
	@echo "Running Go backend tests..."
	go test -v -race -cover ./...

## Run Go tests with coverage report
test-backend-coverage:
	@echo "Running Go tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Run frontend unit tests
test-frontend:
	@echo "Running frontend tests..."
	cd frontend && npm run test:run

## Run frontend tests in watch mode
test-frontend-watch:
	@echo "Running frontend tests in watch mode..."
	cd frontend && npm run test:watch

## Run frontend tests with coverage
test-frontend-coverage:
	@echo "Running frontend tests with coverage..."
	cd frontend && npm run test:coverage

## Run E2E tests with Playwright
test-e2e:
	@echo "Running E2E tests..."
	cd frontend && npm run test:e2e

## Run E2E tests with UI
test-e2e-ui:
	@echo "Running E2E tests with UI..."
	cd frontend && npm run test:e2e:ui

## Run E2E tests in headed mode
test-e2e-headed:
	@echo "Running E2E tests in headed mode..."
	cd frontend && npm run test:e2e:headed

## Install Playwright browsers
test-e2e-install:
	@echo "Installing Playwright browsers..."
	cd frontend && npx playwright install

## Show E2E test report
test-e2e-report:
	@echo "Showing E2E test report..."
	cd frontend && npm run test:e2e:report

## Run all tests with coverage
test-all-coverage:
	@echo "Running all tests with coverage..."
	@$(MAKE) test-backend-coverage
	@$(MAKE) test-frontend-coverage

lint:
	@echo "Running linters..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf tmp/
	go clean -cache

# =============================================================================
# Docker Development
# =============================================================================
docker-dev:
	@echo "Starting development stack..."
	$(DOCKER_COMPOSE_DEV) up --build

docker-dev-down:
	@echo "Stopping development stack..."
	$(DOCKER_COMPOSE_DEV) down

docker-dev-logs:
	$(DOCKER_COMPOSE_DEV) logs -f

# =============================================================================
# Docker Production
# =============================================================================
docker-build:
	@echo "Building production images..."
	$(DOCKER_COMPOSE_PROD) build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT)

docker-up:
	@echo "Starting production stack..."
	$(DOCKER_COMPOSE_PROD) up -d

docker-down:
	@echo "Stopping production stack..."
	$(DOCKER_COMPOSE_PROD) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f

docker-ps:
	$(DOCKER_COMPOSE) ps

docker-restart:
	$(DOCKER_COMPOSE) restart

docker-clean:
	@echo "Cleaning docker resources..."
	docker system prune -f
	docker volume prune -f

# =============================================================================
# Database
# =============================================================================
migrate:
	@echo "Running migrations..."
	@echo "TODO: Add migration command (e.g., golang-migrate)"

migrate-down:
	@echo "Rolling back migration..."
	@echo "TODO: Add rollback command"

# =============================================================================
# Frontend
# =============================================================================
frontend-dev:
	@echo "Starting frontend dev server..."
	cd frontend && npm run dev

frontend-build:
	@echo "Building frontend..."
	cd frontend && npm run build

frontend-install:
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

# =============================================================================
# All-in-one (Primary Commands)
# =============================================================================

## Run locally in development mode
dev:
	@echo "Starting development environment..."
	$(DOCKER_COMPOSE_DEV) up --build

## Build all docker images for production
build: docker-build
	@echo "Build complete!"

## Deploy to production (docker-compose up -d)
deploy:
	@echo "Deploying production stack..."
	$(DOCKER_COMPOSE_PROD) up -d
	@echo "Deployment complete! Waiting for health checks..."
	@sleep 10
	$(DOCKER_COMPOSE_PROD) ps

## Tail all service logs
logs:
	$(DOCKER_COMPOSE_PROD) logs -f --tail=100

## Run database migrations
migrate:
	@echo "Running database migrations..."
	$(DOCKER_COMPOSE_PROD) exec api /app/api migrate up
	@echo "Migrations complete!"

## Run all tests
all: deps build test

# =============================================================================
# SSL / Let's Encrypt
# =============================================================================

## Initialize SSL certificates with Let's Encrypt (first time setup)
ssl-init:
	@echo "Initializing SSL certificates for $(DOMAIN)..."
	@mkdir -p certbot/conf certbot/www
	@docker run -it --rm \
		-v "$(PWD)/certbot/conf:/etc/letsencrypt" \
		-v "$(PWD)/certbot/www:/var/www/certbot" \
		certbot/certbot certonly \
		--webroot \
		--webroot-path=/var/www/certbot \
		--email admin@$(DOMAIN) \
		--agree-tos \
		--no-eff-email \
		-d $(DOMAIN) \
		-d www.$(DOMAIN)
	@echo "SSL certificates generated!"
	@echo "Copying certificates to nginx ssl directory..."
	@mkdir -p deploy/nginx/ssl
	@cp certbot/conf/live/$(DOMAIN)/fullchain.pem deploy/nginx/ssl/
	@cp certbot/conf/live/$(DOMAIN)/privkey.pem deploy/nginx/ssl/
	@cp certbot/conf/live/$(DOMAIN)/chain.pem deploy/nginx/ssl/
	@echo "Done! Restart nginx: make docker-restart"

## Renew SSL certificates
ssl-renew:
	@echo "Renewing SSL certificates..."
	$(DOCKER_COMPOSE_PROD) run --rm certbot renew
	@echo "Certificates renewed. Reloading nginx..."
	$(DOCKER_COMPOSE_PROD) exec nginx nginx -s reload

## Generate self-signed SSL for development
ssl-dev:
	@echo "Generating self-signed SSL certificates for development..."
	@mkdir -p deploy/nginx/ssl
	@openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout deploy/nginx/ssl/privkey.pem \
		-out deploy/nginx/ssl/fullchain.pem \
		-subj "/CN=localhost/O=Ads Analytics/C=US"
	@cp deploy/nginx/ssl/fullchain.pem deploy/nginx/ssl/chain.pem
	@echo "Self-signed certificates generated!"

# =============================================================================
# Shortcuts
# =============================================================================

## Quick status check
status:
	$(DOCKER_COMPOSE_PROD) ps
	@echo ""
	@echo "Service Health:"
	@curl -s http://localhost/health 2>/dev/null || echo "Nginx: DOWN"
	@curl -s http://localhost:8080/health 2>/dev/null || echo "API: DOWN"

## Stop all services
stop:
	$(DOCKER_COMPOSE_PROD) down

## Restart all services
restart:
	$(DOCKER_COMPOSE_PROD) restart

## View specific service logs
logs-api:
	$(DOCKER_COMPOSE_PROD) logs -f api

logs-frontend:
	$(DOCKER_COMPOSE_PROD) logs -f frontend

logs-nginx:
	$(DOCKER_COMPOSE_PROD) logs -f nginx

logs-worker:
	$(DOCKER_COMPOSE_PROD) logs -f worker

## Shell into containers
shell-api:
	$(DOCKER_COMPOSE_PROD) exec api sh

shell-postgres:
	$(DOCKER_COMPOSE_PROD) exec postgres psql -U $${DB_USER:-postgres} -d $${DB_NAME:-ads_aggregator}

shell-redis:
	$(DOCKER_COMPOSE_PROD) exec redis redis-cli

## Database backup
db-backup:
	@echo "Creating database backup..."
	@mkdir -p backups
	$(DOCKER_COMPOSE_PROD) exec postgres pg_dump -U $${DB_USER:-postgres} $${DB_NAME:-ads_aggregator} > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Backup created: backups/backup_$(shell date +%Y%m%d_%H%M%S).sql"

## Database restore
db-restore:
	@echo "Restoring database from $(file)..."
	@test -n "$(file)" || (echo "Usage: make db-restore file=backups/backup.sql" && exit 1)
	$(DOCKER_COMPOSE_PROD) exec -T postgres psql -U $${DB_USER:-postgres} $${DB_NAME:-ads_aggregator} < $(file)
	@echo "Database restored!"

## Full cleanup (WARNING: removes all data)
nuke:
	@echo "WARNING: This will remove all containers, volumes, and images!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	$(DOCKER_COMPOSE_PROD) down -v --rmi all
	docker system prune -af
	@echo "Cleanup complete!"

# =============================================================================
# Production-Like Local Environment
# =============================================================================

## Build and start production-like local environment
prod-local:
	@echo "Building and starting production-like local environment..."
	@echo "This simulates the full production stack with mock platform APIs"
	$(DOCKER_COMPOSE_LOCAL) build \
		--build-arg VERSION=local-test \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT)
	$(DOCKER_COMPOSE_LOCAL) up -d
	@echo ""
	@echo "Waiting for services to be healthy..."
	@sleep 15
	@echo ""
	@echo "========================================="
	@echo "Production-Local Environment Ready!"
	@echo "========================================="
	@echo ""
	@echo "Access URLs:"
	@echo "  Frontend:    http://localhost"
	@echo "  API:         http://localhost/api/v1"
	@echo "  Mock API:    http://localhost/mock-api"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make seed' to populate test data"
	@echo "  2. Run 'make test-flow' to run smoke tests"
	@echo ""
	@$(DOCKER_COMPOSE_LOCAL) ps

## Build production-local images only
prod-local-build:
	@echo "Building production-local images..."
	$(DOCKER_COMPOSE_LOCAL) build \
		--build-arg VERSION=local-test \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT)

## Stop production-local environment
prod-local-down:
	@echo "Stopping production-local environment..."
	$(DOCKER_COMPOSE_LOCAL) down

## View production-local logs
prod-local-logs:
	$(DOCKER_COMPOSE_LOCAL) logs -f --tail=100

## View specific service logs in prod-local
prod-local-logs-api:
	$(DOCKER_COMPOSE_LOCAL) logs -f api

prod-local-logs-frontend:
	$(DOCKER_COMPOSE_LOCAL) logs -f frontend

prod-local-logs-nginx:
	$(DOCKER_COMPOSE_LOCAL) logs -f nginx

prod-local-logs-worker:
	$(DOCKER_COMPOSE_LOCAL) logs -f worker

prod-local-logs-mock:
	$(DOCKER_COMPOSE_LOCAL) logs -f mock-api

## Production-local status
prod-local-status:
	$(DOCKER_COMPOSE_LOCAL) ps
	@echo ""
	@echo "Service Health:"
	@curl -s http://localhost/nginx-health 2>/dev/null && echo " - Nginx: UP" || echo " - Nginx: DOWN"
	@curl -s http://localhost/health 2>/dev/null | grep -q "healthy" && echo " - API: UP" || echo " - API: DOWN"
	@curl -s http://localhost/mock-api/health 2>/dev/null | grep -q "healthy" && echo " - Mock API: UP" || echo " - Mock API: DOWN"

## Seed database with test data
seed:
	@echo "Seeding database with test data..."
	$(DOCKER_COMPOSE_LOCAL) run --rm seed
	@echo ""
	@echo "Seed completed! Test accounts:"
	@echo "  admin@test.com / password123 (Business plan, all platforms)"
	@echo "  pro@test.com / password123 (Pro plan, Meta + TikTok)"
	@echo "  free@test.com / password123 (Free plan, Meta only)"

## Reset database and reseed
reset:
	@echo "WARNING: This will drop all data and reseed!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	@echo "Dropping database..."
	$(DOCKER_COMPOSE_LOCAL) exec postgres psql -U postgres -d ads_local -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "Re-running migrations..."
	$(DOCKER_COMPOSE_LOCAL) restart api
	@sleep 10
	@echo "Reseeding data..."
	$(DOCKER_COMPOSE_LOCAL) run --rm seed
	@echo "Reset complete!"

## Run smoke tests
test-flow:
	@echo "Running smoke tests..."
	@chmod +x scripts/smoke-test.sh
	@bash scripts/smoke-test.sh

## Toggle mock API to error mode (all requests fail)
mock-error:
	@echo "Setting mock API to ERROR mode..."
	@curl -s -X POST http://localhost/mock-api/control/mode \
		-H "Content-Type: application/json" \
		-d '{"mode":"error"}' | jq . 2>/dev/null || echo "Mock API mode set to error"

## Toggle mock API to slow mode (2-5 second delays)
mock-slow:
	@echo "Setting mock API to SLOW mode..."
	@curl -s -X POST http://localhost/mock-api/control/mode \
		-H "Content-Type: application/json" \
		-d '{"mode":"slow"}' | jq . 2>/dev/null || echo "Mock API mode set to slow"

## Toggle mock API to rate limited mode
mock-rate-limited:
	@echo "Setting mock API to RATE LIMITED mode..."
	@curl -s -X POST http://localhost/mock-api/control/mode \
		-H "Content-Type: application/json" \
		-d '{"mode":"rate_limited"}' | jq . 2>/dev/null || echo "Mock API mode set to rate_limited"

## Toggle mock API to token expired mode
mock-token-expired:
	@echo "Setting mock API to TOKEN EXPIRED mode..."
	@curl -s -X POST http://localhost/mock-api/control/mode \
		-H "Content-Type: application/json" \
		-d '{"mode":"token_expired"}' | jq . 2>/dev/null || echo "Mock API mode set to token_expired"

## Reset mock API to normal mode
mock-normal:
	@echo "Resetting mock API to NORMAL mode..."
	@curl -s -X POST http://localhost/mock-api/control/reset | jq . 2>/dev/null || echo "Mock API reset to normal"

## Check mock API current mode
mock-status:
	@echo "Current mock API configuration:"
	@curl -s http://localhost/mock-api/control/mode | jq . 2>/dev/null || echo "Could not get mock API status"

## Shell into prod-local containers
prod-local-shell-api:
	$(DOCKER_COMPOSE_LOCAL) exec api sh

prod-local-shell-postgres:
	$(DOCKER_COMPOSE_LOCAL) exec postgres psql -U postgres -d ads_local

prod-local-shell-redis:
	$(DOCKER_COMPOSE_LOCAL) exec redis redis-cli -a localredis123

## Clean up production-local environment (removes volumes)
prod-local-clean:
	@echo "Cleaning up production-local environment..."
	$(DOCKER_COMPOSE_LOCAL) down -v
	@echo "Cleanup complete!"

## Full production-local nuke (removes everything including images)
prod-local-nuke:
	@echo "WARNING: This will remove all prod-local containers, volumes, and images!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	$(DOCKER_COMPOSE_LOCAL) down -v --rmi all
	@echo "Nuke complete!"
