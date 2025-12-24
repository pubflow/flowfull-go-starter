# 🎯 Core Concepts

This document explains the 7 core concepts of the Flowfull architecture implemented in this Go starter kit.

## 1. 🔐 Bridge Validation

**What is it?**
Bridge Validation is a distributed session validation mechanism that allows your backend to validate user sessions with Flowless without storing session data locally.

**How it works:**
```
Client → Backend → Flowless (Bridge API) → Validate Session → Return User Data
```

**Implementation:**
```go
// Create validator
validator := auth.NewBridgeValidator(cfg, logger)

// Validate session
session, err := validator.ValidateSession(ctx, sessionID, opts)
if err != nil {
    // Invalid session
}

// Access user data
userID := session.UserID
email := session.Email
```

**Benefits:**
- ✅ Centralized session management
- ✅ No local session storage needed
- ✅ Automatic session invalidation
- ✅ Multi-service session sharing

**Configuration:**
```env
FLOWLESS_API_URL=http://localhost:3000
BRIDGE_VALIDATION_SECRET=your-secret-key
BRIDGE_VALIDATION_TIMEOUT=5000
BRIDGE_RETRY_ATTEMPTS=3
```

---

## 2. 🛡️ Validation Modes

**What is it?**
Layered security validation that checks different aspects of a request to prevent session hijacking and unauthorized access.

**Modes:**

### DISABLED
```go
// No validation (development only)
AUTH_VALIDATION_MODE=DISABLED
```
- ⚠️ Use only in development
- No IP, User-Agent, or Device ID checks

### STANDARD
```go
// Basic IP validation
AUTH_VALIDATION_MODE=STANDARD
```
- ✅ IP address validation
- Prevents session use from different IPs

### ADVANCED (Recommended)
```go
// IP + User-Agent validation
AUTH_VALIDATION_MODE=ADVANCED
```
- ✅ IP address validation
- ✅ User-Agent validation
- Detects browser/device changes

### STRICT
```go
// Full validation
AUTH_VALIDATION_MODE=STRICT
```
- ✅ IP address validation
- ✅ User-Agent validation
- ✅ Device ID validation
- Maximum security for sensitive applications

**Implementation:**
```go
opts := &ValidationOptions{
    IP:        c.IP(),
    UserAgent: c.Get("User-Agent"),
    DeviceID:  c.Get("X-Device-ID"),
}

session, err := validator.ValidateSession(ctx, sessionID, opts)
```

---

## 3. ⚡ HybridCache

**What is it?**
A 3-tier caching system that combines in-memory (Ristretto) and distributed (Redis) caching for optimal performance.

**Cache Hierarchy:**
```
L1: Ristretto (in-memory)  → <1ms latency
L2: Redis (distributed)    → <5ms latency
L3: Database/API           → <50ms latency
```

**How it works:**
```go
// 1. Check Ristretto (L1)
if value, found := cache.Get(ctx, key); found {
    return value // <1ms
}

// 2. Check Redis (L2)
if value, found := redis.Get(ctx, key); found {
    cache.Set(ctx, key, value) // Backfill L1
    return value // <5ms
}

// 3. Fetch from source (L3)
value := fetchFromDatabase()
cache.Set(ctx, key, value) // Cache in L1 + L2
return value // <50ms
```

**Usage:**
```go
// Initialize
hybridCache, err := cache.NewHybridCache(
    maxSize,
    redisClient,
    logger,
    enabled,
)

// Get
value, found := hybridCache.Get(ctx, "session:123")

// Set with TTL
hybridCache.Set(ctx, "session:123", sessionData, 5*time.Minute)

// Delete
hybridCache.Delete(ctx, "session:123")

// Metrics
metrics := hybridCache.GetMetrics()
fmt.Printf("Ristretto hits: %d\n", metrics.RistrettoHits)
```

**Performance:**
- **95% cache hit rate** → <1ms average latency
- **5% cache miss** → <50ms latency
- **Overall**: 25x faster than no cache

---

## 4. 🔑 Trust Tokens (PASETO)

**What is it?**
Secure, stateless tokens using PASETO v4 (Platform-Agnostic Security Tokens) with Ed25519 public-key cryptography.

**Why PASETO over JWT?**
- ✅ No algorithm confusion attacks
- ✅ Built-in encryption
- ✅ Simpler API
- ✅ Modern cryptography (Ed25519)

**Generate Keys:**
```bash
make generate-key
```

**Usage:**
```go
// Create token manager
tokenManager, err := tokens.NewTrustTokenManager(privateKey, logger)

// Create token
token, err := tokenManager.CreateToken(
    userID,
    email,
    "access_token",
    24*time.Hour,
    map[string]interface{}{
        "role": "admin",
    },
)

// Verify token
claims, err := tokenManager.VerifyToken(token)
if err != nil {
    // Invalid or expired token
}

// Access claims
userID := claims.UserID
email := claims.Email
```

**Use Cases:**
- API keys
- Service-to-service authentication
- Temporary access tokens
- Password reset tokens
- Email verification tokens

---

## 5. 🎯 Auth Middleware

**What is it?**
Flexible Fiber middleware for protecting routes with different authentication requirements.

**Middleware Types:**

### RequireAuth
```go
// Requires valid session
api.Get("/protected", authMiddleware.RequireAuth(), handler)
```

### OptionalAuth
```go
// Works with or without auth
api.Get("/optional", authMiddleware.OptionalAuth(), handler)
```

### RequireUserType
```go
// Requires specific user type/role
api.Get("/admin", 
    authMiddleware.RequireAuth(),
    authMiddleware.RequireUserType("admin", "superadmin"),
    handler,
)
```

**Accessing User Data:**
```go
func handler(c *fiber.Ctx) error {
    // Get user data from context
    userID := c.Locals("user_id").(string)
    email := c.Locals("email").(string)
    userType := c.Locals("user_type").(string)
    permissions := c.Locals("permissions").([]string)
    
    // Or get full session data
    session := c.Locals("session_data").(*auth.SessionData)
    
    return c.JSON(fiber.Map{
        "user_id": userID,
        "email": email,
    })
}
```

---

## 6. 💾 Multi-Database Support

**What is it?**
GORM-based database abstraction supporting multiple database engines.

**Supported Databases:**
- PostgreSQL (recommended for production)
- MySQL
- SQLite (development only)

**Configuration:**
```env
# PostgreSQL
DATABASE_URL=postgresql://user:pass@localhost:5432/db

# MySQL
DATABASE_URL=mysql://user:pass@localhost:3306/db

# SQLite
DATABASE_URL=sqlite://./app.db
```

**Usage:**
```go
// Connect
db, err := database.NewConnection(cfg, logger)

// Auto-migrate models
db.AutoMigrate(&models.Task{})

// Query
var tasks []models.Task
db.DB.Where("user_id = ?", userID).Find(&tasks)

// Create
task := models.Task{Title: "New Task"}
db.DB.Create(&task)

// Update
db.DB.Save(&task)

// Delete
db.DB.Delete(&task)
```

---

## 7. ⚙️ Environment Configuration

**What is it?**
Type-safe configuration management using Viper with validation.

**Features:**
- ✅ Environment variable loading
- ✅ `.env` file support
- ✅ Type safety
- ✅ Validation
- ✅ Default values

**Usage:**
```go
// Load config
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal(err)
}

// Access config
port := cfg.Port
dbURL := cfg.DatabaseURL
isDev := cfg.IsDevelopment()
```

**Configuration Structure:**
```go
type Config struct {
    // Server
    Port        int
    Host        string
    Environment string
    
    // Database
    DatabaseURL string
    
    // Cache
    CacheEnabled bool
    RedisURL     string
    
    // Auth
    FlowlessAPIURL          string
    BridgeValidationSecret  string
    AuthValidationMode      string
    
    // ... more fields
}
```

---

## Putting It All Together

Here's how all 7 concepts work together in a typical request:

```
1. Request arrives with X-Session-Id header
   ↓
2. Auth Middleware (Concept #5) extracts session
   ↓
3. HybridCache (Concept #3) checks for cached session
   ├─ Cache hit → Return session (fast path)
   └─ Cache miss → Continue
   ↓
4. Bridge Validator (Concept #1) validates with Flowless
   ↓
5. Validation Mode (Concept #2) checks IP/UA/Device
   ↓
6. Session cached in HybridCache for future requests
   ↓
7. User data injected into Fiber context
   ↓
8. Handler accesses Multi-Database (Concept #6) using GORM
   ↓
9. Response returned to client
```

**Configuration** (Concept #7) and **Trust Tokens** (Concept #4) support the entire flow.

---

## Next Steps

- Read [ARCHITECTURE.md](./ARCHITECTURE.md) for system design details
- Read [DEPLOYMENT.md](./DEPLOYMENT.md) for production deployment
- Check the [README.md](../README.md) for quick start guide

