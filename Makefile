.PHONY: help install dev build test lint clean docker-build docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	go mod download
	go mod tidy

dev: ## Run development server with live reload
	air

build: ## Build production binary
	go build -o bin/server cmd/server/main.go

run: ## Run production binary
	./bin/server

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...
	gofmt -s -w .

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

docker-build: ## Build Docker image
	docker build -t flowfull-go-starter:latest .

docker-up: ## Start Docker Compose
	docker-compose up -d

docker-down: ## Stop Docker Compose
	docker-compose down

migrate: ## Run database migrations
	go run cmd/server/main.go migrate

generate-key: ## Generate PASETO key
	go run scripts/generate_paseto_key.go

