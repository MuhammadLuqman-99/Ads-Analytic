# =============================================================================
# Ads Analytics Platform - Makefile
# =============================================================================

.PHONY: help build run test clean docker-build docker-up docker-down docker-logs

# Variables
APP_NAME := ads-aggregator
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go settings
GOFLAGS := -ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Docker settings
DOCKER_COMPOSE := docker-compose
DOCKER_COMPOSE_DEV := $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml
DOCKER_COMPOSE_PROD := $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml

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
	@echo "  make test           Run all tests"
	@echo "  make lint           Run linters"
	@echo "  make fmt            Format code"
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
	@echo "Running tests..."
	go test -v -race -cover ./...

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
# All-in-one
# =============================================================================
dev: docker-dev

prod: docker-build docker-up

all: deps build test
