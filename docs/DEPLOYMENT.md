# 🚀 Deployment Guide

This guide covers deploying the Flowfull Go Starter to various platforms.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Environment Variables](#environment-variables)
- [Docker Deployment](#docker-deployment)
- [Railway Deployment](#railway-deployment)
- [Render Deployment](#render-deployment)
- [Fly.io Deployment](#flyio-deployment)
- [AWS Deployment](#aws-deployment)
- [Production Checklist](#production-checklist)

---

## Prerequisites

Before deploying, ensure you have:

- ✅ Go 1.22+ installed
- ✅ Database (PostgreSQL recommended)
- ✅ Redis instance (optional but recommended)
- ✅ Flowless instance running
- ✅ Environment variables configured

---

## Environment Variables

### Required Variables

```env
# Server
PORT=3001
ENVIRONMENT=production

# Database
DATABASE_URL=postgresql://user:pass@host:5432/db

# Flowless
FLOWLESS_API_URL=https://your-flowless-instance.com
BRIDGE_VALIDATION_SECRET=your-secret-key-min-32-chars

# Auth
AUTH_VALIDATION_MODE=ADVANCED
```

### Optional but Recommended

```env
# Redis
REDIS_URL=redis://host:6379

# PASETO
PASETO_PRIVATE_KEY=your-generated-key

# CORS
CORS_ORIGINS=https://yourdomain.com

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Docker Deployment

### 1. Build Image

```bash
docker build -t flowfull-go-starter:latest .
```

### 2. Run Container

```bash
docker run -d \
  --name flowfull-go-starter \
  -p 3001:3001 \
  -e DATABASE_URL="postgresql://..." \
  -e FLOWLESS_API_URL="https://..." \
  -e BRIDGE_VALIDATION_SECRET="..." \
  flowfull-go-starter:latest
```

### 3. Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

---

## Railway Deployment

Railway supports automatic deployment from Git with Nixpacks.

### 1. Create Railway Project

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login
railway login

# Initialize project
railway init
```

### 2. Configure Environment

In Railway dashboard, add environment variables:

```
DATABASE_URL (from Railway PostgreSQL plugin)
REDIS_URL (from Railway Redis plugin)
FLOWLESS_API_URL
BRIDGE_VALIDATION_SECRET
AUTH_VALIDATION_MODE=ADVANCED
```

### 3. Deploy

```bash
# Link to project
railway link

# Deploy
railway up
```

Railway will automatically:
- Detect Go project
- Use `nixpacks.toml` configuration
- Build and deploy
- Provide a public URL

### 4. Add Database & Redis

```bash
# Add PostgreSQL
railway add postgresql

# Add Redis
railway add redis
```

---

## Render Deployment

### 1. Create `render.yaml`

```yaml
services:
  - type: web
    name: flowfull-go-starter
    env: go
    buildCommand: go build -o server ./cmd/server
    startCommand: ./server
    envVars:
      - key: PORT
        value: 3001
      - key: ENVIRONMENT
        value: production
      - key: DATABASE_URL
        fromDatabase:
          name: flowfull-db
          property: connectionString
      - key: FLOWLESS_API_URL
        value: https://your-flowless.onrender.com
      - key: BRIDGE_VALIDATION_SECRET
        generateValue: true
      - key: AUTH_VALIDATION_MODE
        value: ADVANCED

databases:
  - name: flowfull-db
    databaseName: flowfull
    user: flowfull
```

### 2. Deploy

1. Connect GitHub repository
2. Render auto-detects `render.yaml`
3. Click "Create Web Service"
4. Deployment starts automatically

---

## Fly.io Deployment

### 1. Install Fly CLI

```bash
# macOS
brew install flyctl

# Linux
curl -L https://fly.io/install.sh | sh

# Login
fly auth login
```

### 2. Initialize Fly App

```bash
fly launch
```

This creates `fly.toml`:

```toml
app = "flowfull-go-starter"
primary_region = "iad"

[build]
  builder = "paketobuildpacks/builder:base"

[env]
  PORT = "3001"
  ENVIRONMENT = "production"

[http_service]
  internal_port = 3001
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 0

[[vm]]
  cpu_kind = "shared"
  cpus = 1
  memory_mb = 256
```

### 3. Add PostgreSQL

```bash
fly postgres create
fly postgres attach <postgres-app-name>
```

### 4. Set Secrets

```bash
fly secrets set \
  FLOWLESS_API_URL="https://..." \
  BRIDGE_VALIDATION_SECRET="..." \
  AUTH_VALIDATION_MODE="ADVANCED"
```

### 5. Deploy

```bash
fly deploy
```

---

## AWS Deployment

### Option 1: ECS (Elastic Container Service)

1. **Build and push Docker image to ECR**

```bash
# Authenticate
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build
docker build -t flowfull-go-starter .

# Tag
docker tag flowfull-go-starter:latest \
  <account-id>.dkr.ecr.us-east-1.amazonaws.com/flowfull-go-starter:latest

# Push
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/flowfull-go-starter:latest
```

2. **Create ECS Task Definition**
3. **Create ECS Service**
4. **Configure Load Balancer**

### Option 2: Elastic Beanstalk

```bash
# Install EB CLI
pip install awsebcli

# Initialize
eb init -p go flowfull-go-starter

# Create environment
eb create production

# Deploy
eb deploy
```

---

## Production Checklist

### Security

- [ ] Use HTTPS/TLS
- [ ] Set `ENVIRONMENT=production`
- [ ] Use strong `BRIDGE_VALIDATION_SECRET` (32+ chars)
- [ ] Enable `AUTH_VALIDATION_MODE=ADVANCED` or `STRICT`
- [ ] Configure CORS properly
- [ ] Use environment variables (never commit secrets)
- [ ] Enable rate limiting
- [ ] Set up firewall rules

### Performance

- [ ] Enable Redis caching
- [ ] Set appropriate cache TTLs
- [ ] Configure database connection pooling
- [ ] Use CDN for static assets
- [ ] Enable gzip compression
- [ ] Monitor cache hit rates

### Reliability

- [ ] Set up health checks
- [ ] Configure auto-restart on failure
- [ ] Use managed database (RDS, Cloud SQL, etc.)
- [ ] Set up database backups
- [ ] Configure log aggregation
- [ ] Set up monitoring/alerts

### Monitoring

- [ ] Configure structured logging
- [ ] Set up error tracking (Sentry)
- [ ] Monitor application metrics
- [ ] Set up uptime monitoring
- [ ] Configure alerting

### Database

- [ ] Run migrations before deployment
- [ ] Use connection pooling
- [ ] Enable SSL for database connections
- [ ] Set up automated backups
- [ ] Configure read replicas (if needed)

---

## Health Checks

Configure your platform to use these endpoints:

```
Liveness:  GET /health
Readiness: GET /health/all
```

Example for Kubernetes:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 3001
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /health/all
    port: 3001
  initialDelaySeconds: 5
  periodSeconds: 10
```

---

## Scaling

### Horizontal Scaling

The application is stateless and can be scaled horizontally:

```bash
# Railway
railway scale --replicas 3

# Fly.io
fly scale count 3

# Kubernetes
kubectl scale deployment flowfull-go-starter --replicas=3
```

### Vertical Scaling

Increase resources per instance:

```bash
# Fly.io
fly scale vm shared-cpu-2x --memory 512

# Railway (via dashboard)
```

---

## Troubleshooting

### Application won't start

1. Check logs: `docker logs <container-id>`
2. Verify environment variables
3. Test database connection
4. Check port binding

### High latency

1. Check cache hit rates: `GET /health/cache`
2. Verify Redis connection
3. Monitor database query performance
4. Check network latency to Flowless

### Database connection errors

1. Verify `DATABASE_URL` format
2. Check firewall rules
3. Verify SSL settings
4. Test connection manually

---

## Support

For issues or questions:
- GitHub Issues: [github.com/yourrepo/issues](https://github.com/yourrepo/issues)
- Documentation: [docs](./README.md)

