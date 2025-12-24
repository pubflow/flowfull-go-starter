# 00 - Architecture Overview

## 🏗️ Pubflow Ecosystem Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        PUBFLOW ECOSYSTEM                         │
└─────────────────────────────────────────────────────────────────┘

┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐
│   FLOWLESS       │      │   FLOWFULL       │      │  FLOWFULL-CLIENT │
│  (Auth Backend)  │◄────►│ (Custom Backend) │◄────►│   (Frontend)     │
└──────────────────┘      └──────────────────┘      └──────────────────┘
        │                          │                          │
        │                          │                          │
   Managed Service          Your Backend              React/Next.js/RN
   pubflow.com              Go/Python/Node.js          @pubflow/react
```

---

## 🔄 Authentication Flow

```
1. User authenticates in Flowless
   ↓
2. Flowless generates session_id
   ↓
3. Frontend sends session_id to Flowfull
   ↓
4. Flowfull validates session_id with Flowless (Bridge Validation)
   ↓
5. Flowless returns user data
   ↓
6. Flowfull caches session (HybridCache)
   ↓
7. Flowfull processes request with user context
```

---

## 🏛️ Flowfull-Go Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      FLOWFULL-GO                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │                    Fiber App                            │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐             │    │
│  │  │  Health  │  │   API    │  │  Custom  │             │    │
│  │  │  Routes  │  │  Routes  │  │  Routes  │             │    │
│  │  └──────────┘  └──────────┘  └──────────┘             │    │
│  └────────────────────────────────────────────────────────┘    │
│                           │                                      │
│  ┌────────────────────────┼────────────────────────────────┐   │
│  │         Auth Middleware (RequireAuth, OptionalAuth)     │   │
│  └────────────────────────┼────────────────────────────────┘   │
│                           │                                      │
│  ┌────────────────────────┼────────────────────────────────┐   │
│  │              Bridge Validator                           │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │  Validation Modes:                               │   │   │
│  │  │  • DISABLED  - No validation                     │   │   │
│  │  │  • STANDARD  - IP validation                     │   │   │
│  │  │  • ADVANCED  - IP + User-Agent                   │   │   │
│  │  │  • STRICT    - IP + User-Agent + Device ID       │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  └────────────────────────┼────────────────────────────────┘   │
│                           │                                      │
│  ┌────────────────────────┼────────────────────────────────┐   │
│  │              HybridCache (3-tier)                       │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐             │   │
│  │  │Ristretto │→ │  Redis   │→ │ Database │             │   │
│  │  │  Cache   │  │  Cache   │  │  Query   │             │   │
│  │  │  (Fast)  │  │ (Shared) │  │ (Source) │             │   │
│  │  └──────────┘  └──────────┘  └──────────┘             │   │
│  └────────────────────────┼────────────────────────────────┘   │
│                           │                                      │
│  ┌────────────────────────┼────────────────────────────────┐   │
│  │              Database Layer (GORM)                      │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐             │   │
│  │  │PostgreSQL│  │  MySQL   │  │  SQLite  │             │   │
│  │  └──────────┘  └──────────┘  └──────────┘             │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔀 Concurrency Model

```
┌─────────────────────────────────────────────────────────────┐
│                    Go Concurrency                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  HTTP Request → Goroutine                                   │
│                    │                                         │
│                    ├─→ Auth Validation (goroutine)          │
│                    │   ├─→ Cache Check (concurrent)         │
│                    │   └─→ Bridge Validation (if needed)    │
│                    │                                         │
│                    ├─→ Database Query (goroutine pool)      │
│                    │                                         │
│                    └─→ Response (channel)                   │
│                                                              │
│  Features:                                                   │
│  • Worker pools for database operations                     │
│  • Concurrent cache operations                              │
│  • Context-based cancellation                               │
│  • Graceful shutdown with sync.WaitGroup                    │
│  • Rate limiting with token bucket                          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 📦 Module Structure

```
flowfull-go/
│
├── cmd/server/main.go             # Entry point
│
├── internal/
│   ├── config/
│   │   └── environment.go         # Viper configuration
│   │
│   ├── lib/
│   │   ├── auth/
│   │   │   ├── bridge_validator.go    # Bridge Validation
│   │   │   ├── middleware.go          # Auth Middleware
│   │   │   └── validation_mode.go     # Validation Modes
│   │   │
│   │   ├── cache/
│   │   │   ├── hybrid_cache.go        # HybridCache
│   │   │   └── cache_instances.go     # Cache instances
│   │   │
│   │   ├── database/
│   │   │   ├── connection.go          # GORM setup
│   │   │   └── migrations.go          # Auto-migrations
│   │   │
│   │   ├── tokens/
│   │   │   └── trust_tokens.go        # PASETO tokens
│   │   │
│   │   └── utils/
│   │       └── logger.go              # Zap logger
│   │
│   ├── models/
│   │   └── user.go                # GORM models
│   │
│   └── routes/
│       ├── health.go              # Health routes
│       └── api.go                 # API routes
│
├── tests/
│   ├── integration/
│   └── unit/
│
├── go.mod
├── Dockerfile
└── Makefile
```

---

## 🔐 Request Flow with Authentication

```
1. Client Request
   │
   ├─ Headers: X-Session-Id: abc123...
   │
   ↓
2. Fiber Route
   │
   ├─ app.Get("/api/profile", middleware.RequireAuth(), handler)
   │
   ↓
3. Auth Middleware (RequireAuth)
   │
   ├─ Extract session_id from header
   │
   ↓
4. Check HybridCache (concurrent)
   │
   ├─ Ristretto Cache? → ✅ Return (sub-ms)
   ├─ Redis Cache? → ✅ Return + Backfill Ristretto (1-5ms)
   ├─ Cache Miss? → Continue to Bridge Validation
   │
   ↓
5. Bridge Validation (goroutine)
   │
   ├─ POST to Flowless: /api/bridge/validate
   ├─ Body: {session_id, bridge_secret, ip, user_agent}
   ├─ Retry logic with exponential backoff
   │
   ↓
6. Flowless Response
   │
   ├─ 200 OK → SessionData {user_id, email, name, ...}
   ├─ 401 Unauthorized → Reject request
   │
   ↓
7. Cache Session (concurrent writes)
   │
   ├─ Save to Ristretto Cache
   ├─ Save to Redis (TTL: 5 minutes)
   │
   ↓
8. Execute Route Handler (goroutine)
   │
   ├─ Access c.Locals("user_id"), c.Locals("email"), etc.
   ├─ Query database with user context
   ├─ Return response
   │
   ↓
9. Response to Client
```

---

## 🎯 The 7 Core Concepts - Mapping

| # | Concept | File | Description |
|---|---------|------|-------------|
| 1 | **Bridge Validation** | `internal/lib/auth/bridge_validator.go` | Distributed validation with Flowless |
| 2 | **Validation Modes** | `internal/lib/auth/validation_mode.go` | Layered security |
| 3 | **HybridCache** | `internal/lib/cache/hybrid_cache.go` | 3-tier cache |
| 4 | **Trust Tokens** | `internal/lib/tokens/trust_tokens.go` | PASETO v4 tokens |
| 5 | **Auth Middleware** | `internal/lib/auth/middleware.go` | Route protection |
| 6 | **Multi-Database** | `internal/lib/database/connection.go` | Multi-DB support |
| 7 | **Environment Config** | `internal/config/environment.go` | Validated config |

---

## 📊 Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| **Cache Hit Rate** | >95% | Ristretto + Redis |
| **Auth Latency (cached)** | <1ms | Ristretto hit |
| **Auth Latency (Redis)** | <5ms | Redis hit |
| **Auth Latency (Bridge)** | <50ms | Flowless validation |
| **Request Throughput** | >50k req/s | With cache |
| **Database Pool** | 25-100 connections | Configurable |
| **Memory Usage** | <100MB | Base + cache |
| **Goroutines** | 1000s concurrent | Lightweight |

---

**Next**: [README.md](./README.md) - Project overview

