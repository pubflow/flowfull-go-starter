# 🏗️ Architecture Overview

## System Architecture

The Flowfull Go Starter follows a **layered architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Layer (Fiber)                    │
│                  Routes & Middleware                     │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                   Business Logic Layer                   │
│              Handlers & Service Logic                    │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                    Data Access Layer                     │
│                  GORM Models & Queries                   │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                   │
│        Database, Cache, External Services                │
└─────────────────────────────────────────────────────────┘
```

## Core Components

### 1. **Bridge Validator** (`internal/lib/auth/bridge_validator.go`)

Handles distributed session validation with Flowless.

**Flow:**
```
Client Request → Extract session_id → Check Cache → Validate with Flowless → Cache Result
```

**Features:**
- HTTP client with retry logic (Resty)
- Configurable timeout and retry attempts
- Structured logging with Zap
- Error handling and validation

### 2. **HybridCache** (`internal/lib/cache/hybrid_cache.go`)

3-tier caching system for maximum performance.

**Cache Hierarchy:**
```
Request → Ristretto (in-memory) → Redis (distributed) → Database
           <1ms                     <5ms                 <50ms
```

**Benefits:**
- **Ristretto**: Ultra-fast in-memory cache (sub-millisecond)
- **Redis**: Distributed cache for multi-instance deployments
- **Automatic backfill**: Redis hits populate Ristretto
- **Metrics tracking**: Monitor cache hit rates

### 3. **Auth Middleware** (`internal/lib/auth/middleware.go`)

Flexible authentication middleware with multiple modes.

**Middleware Types:**
- `RequireAuth()`: Enforces authentication
- `OptionalAuth()`: Authentication optional
- `RequireUserType(types...)`: Role-based access control

**Context Injection:**
```go
c.Locals("user_id")         // string
c.Locals("email")           // string
c.Locals("user_type")       // string
c.Locals("permissions")     // []string
c.Locals("session_data")    // *SessionData
```

### 4. **Validation Modes** (`internal/lib/auth/validation_mode.go`)

Layered security validation.

| Mode | IP | User-Agent | Device ID | Use Case |
|------|----|-----------|-----------| ---------|
| **DISABLED** | ❌ | ❌ | ❌ | Development |
| **STANDARD** | ✅ | ❌ | ❌ | Basic security |
| **ADVANCED** | ✅ | ✅ | ❌ | Recommended |
| **STRICT** | ✅ | ✅ | ✅ | High security |

### 5. **Trust Tokens** (`internal/lib/tokens/trust_tokens.go`)

PASETO v4 tokens for secure, stateless authentication.

**Features:**
- Ed25519 public-key cryptography
- Tamper-proof tokens
- Built-in expiration
- No database lookups needed

**Use Cases:**
- API keys
- Service-to-service auth
- Temporary access tokens
- Password reset tokens

### 6. **Database Connection** (`internal/lib/database/connection.go`)

Multi-database support with GORM.

**Supported Databases:**
- PostgreSQL (recommended)
- MySQL
- SQLite (development)

**Features:**
- Connection pooling
- Auto-migration
- Health checks
- Structured logging

## Request Flow

### Protected Endpoint Request

```
1. Client sends request with X-Session-Id header
   ↓
2. Auth Middleware extracts session_id
   ↓
3. Check Ristretto cache (in-memory)
   ├─ HIT → Return cached session (sub-ms)
   └─ MISS → Continue
   ↓
4. Check Redis cache (distributed)
   ├─ HIT → Backfill Ristretto → Return session (<5ms)
   └─ MISS → Continue
   ↓
5. Validate with Flowless Bridge API
   ├─ SUCCESS → Cache in Redis + Ristretto → Return session
   └─ FAIL → Return 401 Unauthorized
   ↓
6. Inject user data into Fiber context
   ↓
7. Execute route handler
   ↓
8. Return response
```

## Performance Characteristics

### Latency Breakdown

| Operation | Latency | Cache Tier |
|-----------|---------|------------|
| Ristretto hit | <1ms | L1 (in-memory) |
| Redis hit | <5ms | L2 (distributed) |
| Bridge validation | <50ms | L3 (external API) |
| Database query | <10ms | L4 (persistent) |

### Throughput

- **With cache (95% hit rate)**: >50,000 req/s
- **Without cache**: ~2,000 req/s
- **Cache speedup**: 25x improvement

## Security Model

### Defense in Depth

```
┌─────────────────────────────────────────┐
│  1. HTTPS/TLS Encryption                │
├─────────────────────────────────────────┤
│  2. CORS Policy                         │
├─────────────────────────────────────────┤
│  3. Rate Limiting                       │
├─────────────────────────────────────────┤
│  4. Session Validation (Bridge)         │
├─────────────────────────────────────────┤
│  5. Validation Modes (IP, UA, Device)   │
├─────────────────────────────────────────┤
│  6. Role-Based Access Control (RBAC)    │
├─────────────────────────────────────────┤
│  7. Input Validation                    │
└─────────────────────────────────────────┘
```

### Session Security

- **Bridge Validation**: Centralized session management
- **Cache TTL**: Automatic session expiration
- **IP Validation**: Prevent session hijacking
- **User-Agent Validation**: Detect suspicious activity
- **Device ID Validation**: Multi-device tracking

## Scalability

### Horizontal Scaling

```
┌──────────┐   ┌──────────┐   ┌──────────┐
│  App 1   │   │  App 2   │   │  App 3   │
└────┬─────┘   └────┬─────┘   └────┬─────┘
     │              │              │
     └──────────────┴──────────────┘
                    │
         ┌──────────┴──────────┐
         │                     │
    ┌────▼────┐          ┌────▼────┐
    │  Redis  │          │   DB    │
    │ (Cache) │          │ (GORM)  │
    └─────────┘          └─────────┘
```

**Benefits:**
- Stateless application servers
- Shared Redis cache
- Load balancer compatible
- Auto-scaling ready

### Caching Strategy

- **Write-through**: Update cache on write
- **TTL-based expiration**: Automatic cleanup
- **Cache warming**: Pre-populate on startup
- **Metrics monitoring**: Track hit rates

## Configuration Management

### Environment-based Config

```go
config.LoadConfig()
  ↓
Viper reads .env file
  ↓
Validates required fields
  ↓
Returns typed Config struct
```

### Config Validation

- Required fields checked at startup
- Type safety with Go structs
- Default values for optional fields
- Environment-specific overrides

## Error Handling

### Error Propagation

```go
Database Error → GORM Error → Service Error → HTTP Error → Client
```

### Error Response Format

```json
{
  "error": "human-readable message",
  "code": 400,
  "path": "/api/tasks",
  "method": "POST"
}
```

## Monitoring & Observability

### Structured Logging (Zap)

```go
logger.Info("session validated",
    zap.String("user_id", userID),
    zap.String("email", email),
    zap.Duration("latency", duration),
)
```

### Health Checks

- `/health` - Basic health
- `/health/db` - Database connectivity
- `/health/cache` - Redis connectivity
- `/health/all` - Complete system health

### Metrics

- Cache hit/miss rates
- Request latency
- Error rates
- Database connection pool stats

## Best Practices

1. **Always use middleware** for authentication
2. **Cache aggressively** with appropriate TTLs
3. **Log structured data** for easy querying
4. **Validate input** at the handler level
5. **Use transactions** for multi-step operations
6. **Handle errors gracefully** with proper HTTP codes
7. **Monitor cache metrics** to optimize hit rates
8. **Use connection pooling** for database efficiency

