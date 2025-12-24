# 🚀 Flowfull Go Starter Kit

A production-ready Go backend starter template built with **Fiber**, **GORM**, and the **Flowfull/Flowless** ecosystem. This starter kit implements all 7 core concepts of Flowfull architecture.

## ✨ Features

- 🔐 **Bridge Validation** - Distributed session validation with Flowless
- 🛡️ **Validation Modes** - Layered security (DISABLED, STANDARD, ADVANCED, STRICT)
- ⚡ **HybridCache** - 3-tier caching (Ristretto → Redis → Database)
- 🔑 **Trust Tokens** - PASETO v4 tokens for secure authentication
- 🎯 **Auth Middleware** - Flexible route protection (RequireAuth, OptionalAuth, RequireUserType)
- 💾 **Multi-Database** - Support for PostgreSQL, MySQL, and SQLite
- ⚙️ **Environment Config** - Validated configuration with Viper

## 📦 Tech Stack

- **Web Framework**: [Fiber v2](https://gofiber.io/) - Fast HTTP framework
- **ORM**: [GORM](https://gorm.io/) - Multi-database ORM
- **Cache**: [Ristretto](https://github.com/dgraph-io/ristretto) + [Redis](https://redis.io/)
- **Config**: [Viper](https://github.com/spf13/viper) - Configuration management
- **Logging**: [Zap](https://github.com/uber-go/zap) - Structured logging
- **Tokens**: [PASETO](https://github.com/o1egl/paseto) - Secure tokens
- **Validation**: [validator](https://github.com/go-playground/validator)

## 🏗️ Project Structure

```
flowfull-go-starter/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── environment.go       # Configuration with Viper
│   ├── lib/
│   │   ├── auth/
│   │   │   ├── bridge_validator.go    # Bridge Validation
│   │   │   ├── middleware.go          # Auth Middleware
│   │   │   ├── validation_mode.go     # Validation Modes
│   │   │   └── types.go               # Auth types
│   │   ├── cache/
│   │   │   └── hybrid_cache.go        # HybridCache implementation
│   │   ├── database/
│   │   │   └── connection.go          # Database connection
│   │   ├── tokens/
│   │   │   └── trust_tokens.go        # PASETO tokens
│   │   └── utils/
│   │       └── logger.go              # Zap logger
│   ├── models/
│   │   └── task.go              # Example GORM model
│   └── routes/
│       ├── health.go            # Health check routes
│       └── api.go               # API routes
├── scripts/
│   └── generate_paseto_key.go   # Generate PASETO keys
├── .env.example                 # Environment variables template
├── .gitignore
├── .air.toml                    # Live reload config
├── go.mod
├── Makefile
└── README.md
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.22+** - [Download](https://go.dev/dl/)
- **PostgreSQL/MySQL/SQLite** - Choose your database
- **Redis** (optional but recommended)

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/flowfull-go-starter.git
cd flowfull-go-starter

# Install dependencies
make install

# Copy environment file
cp .env.example .env
```

### 2. Configure Environment

Edit `.env` with your settings:

```env
# Server
PORT=3001
ENVIRONMENT=development

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/mydb

# Flowless Integration
FLOWLESS_API_URL=http://localhost:3000
BRIDGE_VALIDATION_SECRET=your-secret-key-min-32-chars

# Redis (optional)
REDIS_URL=redis://localhost:6379

# Auth Validation Mode
AUTH_VALIDATION_MODE=STANDARD  # DISABLED | STANDARD | ADVANCED | STRICT
```

### 3. Generate PASETO Key (Optional)

```bash
make generate-key
# Copy the generated private key to .env
```

### 4. Run Development Server

```bash
# With live reload
make dev

# Or without live reload
make build
make run
```

The server will start at `http://localhost:3001`

## 📚 API Endpoints

### Health Checks

```bash
GET /health          # Basic health check
GET /health/db       # Database health
GET /health/cache    # Cache health
GET /health/all      # Complete health check
```

### Public Routes

```bash
GET /                # Service info
GET /api/public      # Public endpoint (no auth)
```

### Protected Routes (Require Authentication)

```bash
GET /api/protected   # Protected endpoint
GET /api/profile     # User profile
GET /api/tasks       # List user tasks
POST /api/tasks      # Create task
GET /api/tasks/:id   # Get task
PUT /api/tasks/:id   # Update task
DELETE /api/tasks/:id # Delete task
```

### Optional Auth Routes

```bash
GET /api/optional    # Works with or without auth
```

## 🔐 Authentication

This starter uses **Bridge Validation** with Flowless for authentication.

### How it works:

1. User authenticates with Flowless
2. Flowless generates a `session_id`
3. Frontend sends `session_id` in header: `X-Session-Id`
4. Backend validates with Flowless via Bridge Validation
5. Session is cached in HybridCache (Ristretto + Redis)
6. Subsequent requests use cached session (sub-millisecond latency)

### Example Request:

```bash
curl -H "X-Session-Id: your-session-id" \
     http://localhost:3001/api/protected
```

## 🛡️ Validation Modes

Configure security level in `.env`:

```env
AUTH_VALIDATION_MODE=STANDARD
```

| Mode | IP Check | User-Agent | Device ID | Use Case |
|------|----------|------------|-----------|----------|
| **DISABLED** | ❌ | ❌ | ❌ | Development only |
| **STANDARD** | ✅ | ❌ | ❌ | Basic security |
| **ADVANCED** | ✅ | ✅ | ❌ | Recommended |
| **STRICT** | ✅ | ✅ | ✅ | High security |

## ⚡ Performance

- **Cache Hit Rate**: >95% (Ristretto + Redis)
- **Auth Latency (cached)**: <1ms (Ristretto)
- **Auth Latency (Redis)**: <5ms
- **Auth Latency (Bridge)**: <50ms
- **Throughput**: >50k req/s with cache

## 🧪 Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Format code
make fmt
```

## 📦 Available Commands

```bash
make help           # Show all commands
make install        # Install dependencies
make dev            # Run with live reload
make build          # Build binary
make run            # Run binary
make test           # Run tests
make lint           # Run linter
make clean          # Clean artifacts
make generate-key   # Generate PASETO key
```

## 🐳 Docker Support

```bash
# Build image
make docker-build

# Run with docker-compose
make docker-up

# Stop
make docker-down
```

## 📖 Documentation

For detailed documentation, see the `/docs` folder or visit:
- [Architecture Overview](./docs/ARCHITECTURE.md)
- [Core Concepts](./docs/CORE-CONCEPTS.md)
- [Deployment Guide](./docs/DEPLOYMENT.md)

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](./LICENSE) for details.

## 🙏 Acknowledgments

Built with the [Flowfull/Flowless](https://pubflow.com) ecosystem.

---

**Made with ❤️ by the Pubflow Team**

