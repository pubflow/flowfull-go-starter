# 01 - Initial Setup

## 📋 Prerequisites

- **Go 1.22+** - [Download](https://go.dev/dl/)
- **PostgreSQL 14+** or **MySQL 8+** or **SQLite 3**
- **Redis 7+** (optional but recommended)
- **Git**
- **Make** (optional, for Makefile commands)

---

## 🚀 Project Initialization

### 1. Create Project Structure

```bash
mkdir flowfull-go
cd flowfull-go

# Initialize Go module
go mod init github.com/yourusername/flowfull-go
```

### 2. Create Directory Structure

```bash
mkdir -p cmd/server
mkdir -p internal/{config,lib/{auth,cache,database,tokens,utils},models,routes}
mkdir -p tests/{unit,integration}
mkdir -p scripts
```

---

## 📦 go.mod Configuration

### go.mod

```go
module github.com/yourusername/flowfull-go

go 1.22

require (
    // Web Framework
    github.com/gofiber/fiber/v2 v2.52.0
    
    // Database & ORM
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
    gorm.io/driver/mysql v1.5.2
    gorm.io/driver/sqlite v1.5.4
    
    // Cache
    github.com/dgraph-io/ristretto v0.1.1
    github.com/redis/go-redis/v9 v9.3.0
    
    // Configuration
    github.com/spf13/viper v1.18.2
    github.com/go-playground/validator/v10 v10.16.0
    
    // Authentication & Security
    github.com/o1egl/paseto v1.0.0
    golang.org/x/crypto v0.17.0
    
    // HTTP Client
    github.com/go-resty/resty/v2 v2.11.0
    
    // Logging
    go.uber.org/zap v1.26.0
    
    // Utilities
    github.com/google/uuid v1.5.0
    golang.org/x/sync v0.5.0
)

require (
    // Development
    github.com/cosmtrek/air v1.49.0 // Live reload
    github.com/stretchr/testify v1.8.4 // Testing
    github.com/golang/mock v1.6.0 // Mocking
)
```

---

## 🔧 Configuration Files

### .env.example

```env
# ================================
# FLOWFULL-GO - Backend Template
# ================================

# Server Configuration
PORT=3001
HOST=0.0.0.0
ENVIRONMENT=development
BASE_URL=http://localhost:3001

# Database Configuration
DATABASE_URL=postgresql://username:password@localhost:5432/flowfull_db
# DATABASE_URL=mysql://username:password@tcp(localhost:3306)/flowfull_db
# DATABASE_URL=sqlite://./flowfull.db

# Database Pool Settings
DATABASE_MAX_IDLE_CONNS=10
DATABASE_MAX_OPEN_CONNS=100
DATABASE_CONN_MAX_LIFETIME=3600

# Flowless Integration
FLOWLESS_API_URL=http://localhost:3000
BRIDGE_VALIDATION_SECRET=your-shared-secret-key-here-min-32-chars
BRIDGE_VALIDATION_TIMEOUT=5000
BRIDGE_RETRY_ATTEMPTS=3

# Session Management
SESSION_VALIDATION_CACHE_TTL=300
SESSION_HEADER_NAME=X-Session-Id
SESSION_COOKIE_NAME=session_id

# ================================
# Authentication & Validation Mode
# ================================
AUTH_VALIDATION_MODE=STANDARD          # DISABLED | STANDARD | ADVANCED | STRICT
AUTH_ENABLE_VALIDATION_MODE=true
AUTH_IP_VALIDATION=true
AUTH_USER_AGENT_VALIDATION=true
AUTH_DEVICE_VALIDATION=false
AUTH_AUTO_INVALIDATE=false
AUTH_LOG_VIOLATIONS=true

# ================================
# HybridCache Configuration
# ================================
CACHE_ENABLED=true
CACHE_MAX_SIZE=50000
CACHE_NUM_COUNTERS=500000
REDIS_URL=redis://localhost:6379
# REDIS_URL=redis://:password@localhost:6379
# REDIS_URL=rediss://default:password@host:6379

# ================================
# Trust Tokens (PASETO v4)
# ================================
# Generate with: go run scripts/generate_paseto_key.go
PASETO_PRIVATE_KEY=

# Token TTL (hours)
TOKEN_TTL_HOURS=168
TOKEN_EMAIL_VERIFICATION_TTL_HOURS=24
TOKEN_PASSWORD_RESET_TTL_HOURS=1
TOKEN_INVITATION_TTL_HOURS=168

# ================================
# Security & CORS
# ================================
CORS_ORIGINS=http://localhost:3000,http://localhost:5173
CORS_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_HEADERS=Content-Type,Authorization,X-Session-Id
CORS_CREDENTIALS=true
CORS_MAX_AGE=86400

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60

# ================================
# Logging & Monitoring
# ================================
LOG_LEVEL=info
LOG_FORMAT=json
LOG_MODE=false

# ================================
# Development Settings
# ================================
DEV_MODE=true
DEV_CORS_RELAXED=true
DEV_LOG_REQUESTS=true
RELOAD=true
```

---

## 🛠️ Development Tools

### Makefile

```makefile
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
	docker build -t flowfull-go:latest .

docker-up: ## Start Docker Compose
	docker-compose up -d

docker-down: ## Stop Docker Compose
	docker-compose down

migrate: ## Run database migrations
	go run cmd/server/main.go migrate

generate-key: ## Generate PASETO key
	go run scripts/generate_paseto_key.go
```


### .air.toml (Live Reload)

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "tests"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

### .gitignore

```gitignore
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out
coverage.html

# Dependency directories
vendor/

# Go workspace file
go.work

# Environment files
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Temporary files
tmp/
*.log

# Database files
*.db
*.sqlite
*.sqlite3

# Air
.air.toml.local
```

### .golangci.yml (Linter Configuration)

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - unconvert
    - unparam
    - gosec
    - gocritic

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true

  govet:
    check-shadowing: true

  gofmt:
    simplify: true

  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

---

## 📝 Installation Steps

### 1. Install Dependencies

```bash
# Install all dependencies
go mod download
go mod tidy

# Verify installation
go mod verify
```

### 2. Install Development Tools

```bash
# Install Air (live reload)
go install github.com/cosmtrek/air@latest

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install gomock (for testing)
go install github.com/golang/mock/mockgen@latest
```

### 3. Setup Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your configuration
nano .env
```

### 4. Verify Setup

```bash
# Format code
make fmt

# Run linter
make lint

# Build project
make build

# Run tests (will fail until implementation)
make test
```

---

## 🚀 Development Commands

```bash
# Start development server with live reload
make dev

# Or without Makefile
air

# Build production binary
make build

# Run production binary
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint

# Generate PASETO key
make generate-key
```

---

## 📊 Project Structure Verification

After setup, your structure should look like:

```
flowfull-go/
├── cmd/
│   └── server/
│       └── main.go (to be created)
├── internal/
│   ├── config/
│   ├── lib/
│   ├── models/
│   └── routes/
├── tests/
│   ├── unit/
│   └── integration/
├── scripts/
├── .env.example
├── .gitignore
├── .air.toml
├── .golangci.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

**Next**: [02-CORE-CONCEPTS.md](./02-CORE-CONCEPTS.md) - Core concepts implementation

