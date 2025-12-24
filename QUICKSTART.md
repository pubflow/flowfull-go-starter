# ⚡ Quick Start Guide

Get up and running with Flowfull Go Starter in 5 minutes!

## Prerequisites

- Go 1.22+ ([Download](https://go.dev/dl/))
- PostgreSQL or SQLite
- (Optional) Redis for caching

## Step 1: Clone & Install

```bash
# Clone the repository
git clone https://github.com/yourrepo/flowfull-go-starter.git
cd flowfull-go-starter

# Install dependencies
make install
```

## Step 2: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env file
nano .env  # or use your favorite editor
```

**Minimum required configuration:**

```env
# Server
PORT=3001
ENVIRONMENT=development

# Database (choose one)
DATABASE_URL=sqlite://./flowfull.db
# DATABASE_URL=postgresql://user:pass@localhost:5432/flowfull_dev

# Flowless Integration
FLOWLESS_API_URL=http://localhost:3000
BRIDGE_VALIDATION_SECRET=your-secret-key-must-be-at-least-32-characters-long

# Auth
AUTH_VALIDATION_MODE=STANDARD
```

## Step 3: Generate PASETO Key (Optional)

```bash
make generate-key
```

Copy the generated private key to your `.env`:

```env
PASETO_PRIVATE_KEY=your-generated-key-here
```

## Step 4: Run Development Server

```bash
# With live reload
make dev

# Or without live reload
make build
make run
```

Server starts at: **http://localhost:3001**

## Step 5: Test the API

### Health Check

```bash
curl http://localhost:3001/health
```

Response:
```json
{
  "status": "ok",
  "service": "flowfull-go-starter"
}
```

### Public Endpoint

```bash
curl http://localhost:3001/api/public
```

### Protected Endpoint (requires authentication)

```bash
curl -H "X-Session-Id: your-session-id" \
     http://localhost:3001/api/protected
```

## Using Docker Compose (Alternative)

If you prefer Docker:

```bash
# Start all services (app + PostgreSQL + Redis)
make docker-up

# View logs
docker-compose logs -f

# Stop services
make docker-down
```

## Next Steps

### 1. Explore the API

- **Health Checks**: `/health`, `/health/db`, `/health/cache`
- **Public Routes**: `/api/public`
- **Protected Routes**: `/api/protected`, `/api/profile`, `/api/tasks`

### 2. Read Documentation

- [README.md](./README.md) - Complete overview
- [ARCHITECTURE.md](./docs/ARCHITECTURE.md) - System architecture
- [CORE-CONCEPTS.md](./docs/CORE-CONCEPTS.md) - 7 core concepts explained
- [DEPLOYMENT.md](./docs/DEPLOYMENT.md) - Production deployment

### 3. Customize

- Add your own models in `internal/models/`
- Create routes in `internal/routes/`
- Add business logic in handlers
- Configure validation modes
- Set up caching strategy

### 4. Test

```bash
# Run tests
make test

# Run with coverage
make test-coverage

# Run linter
make lint
```

## Common Issues

### Port already in use

Change the port in `.env`:
```env
PORT=3002
```

### Database connection failed

For SQLite (easiest for development):
```env
DATABASE_URL=sqlite://./flowfull.db
```

For PostgreSQL:
```env
DATABASE_URL=postgresql://user:pass@localhost:5432/dbname?sslmode=disable
```

### Flowless connection failed

Make sure Flowless is running and accessible:
```env
FLOWLESS_API_URL=http://localhost:3000
```

Or use a deployed Flowless instance:
```env
FLOWLESS_API_URL=https://your-flowless.railway.app
```

## Development Workflow

```bash
# 1. Make changes to code
# 2. Live reload automatically restarts server (if using `make dev`)
# 3. Test your changes
curl http://localhost:3001/your-endpoint

# 4. Run tests
make test

# 5. Format code
make fmt

# 6. Run linter
make lint

# 7. Commit changes
git add .
git commit -m "feat: add new feature"
```

## Available Commands

```bash
make help           # Show all commands
make install        # Install dependencies
make dev            # Run with live reload
make build          # Build binary
make run            # Run binary
make test           # Run tests
make lint           # Run linter
make fmt            # Format code
make clean          # Clean artifacts
make generate-key   # Generate PASETO key
make docker-build   # Build Docker image
make docker-up      # Start Docker Compose
make docker-down    # Stop Docker Compose
```

## Project Structure

```
flowfull-go-starter/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration
│   ├── lib/             # Core libraries
│   │   ├── auth/        # Authentication
│   │   ├── cache/       # Caching
│   │   ├── database/    # Database
│   │   ├── tokens/      # PASETO tokens
│   │   └── utils/       # Utilities
│   ├── models/          # Database models
│   └── routes/          # HTTP routes
├── scripts/             # Utility scripts
├── docs/                # Documentation
└── .env                 # Environment config
```

## Need Help?

- 📖 Read the [full documentation](./README.md)
- 🐛 Report issues on [GitHub](https://github.com/yourrepo/issues)
- 💬 Join our [Discord](https://discord.gg/yourserver)
- 📧 Email: support@pubflow.com

---

**Happy coding! 🚀**

