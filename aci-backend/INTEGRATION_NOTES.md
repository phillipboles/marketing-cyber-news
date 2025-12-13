# Authentication Service Integration Notes

## Files Created

### 1. Repository Layer
- **`internal/repository/postgres/refresh_token_repo.go`**
  - Implements `RefreshTokenRepository` interface
  - Handles refresh token CRUD operations with token hashing
  - Methods: Create, GetByTokenHash, Revoke, RevokeAllForUser, DeleteExpired

### 2. Service Layer
- **`internal/service/auth_service.go`**
  - Core authentication business logic
  - Methods:
    - `Register(email, password, name)` - User registration with validation
    - `Login(email, password)` - User authentication
    - `Refresh(refreshToken)` - Token refresh with rotation
    - `Logout(refreshToken)` - Single device logout
    - `LogoutAll(userID)` - All devices logout
  - Password validation: min 8 chars, requires uppercase, lowercase, digit
  - Email validation with regex
  - Uses bcrypt for password hashing
  - Uses SHA-256 for token hashing

- **`internal/service/user_adapter.go`**
  - Interface adapter to work around domain.User vs entities.User mismatch
  - Defines `UserRepoInterface` used by AuthService

### 3. API Handlers
- **`internal/api/handlers/auth_handler.go`**
  - HTTP request handlers for authentication endpoints
  - Methods:
    - `Register(w, r)` - POST /v1/auth/register
    - `Login(w, r)` - POST /v1/auth/login
    - `Refresh(w, r)` - POST /v1/auth/refresh
    - `Logout(w, r)` - POST /v1/auth/logout
  - DTOs: RegisterRequest, LoginRequest, RefreshRequest, LogoutRequest
  - Response types: AuthResponse, TokenResponse, UserDTO

### 4. Middleware
- **`internal/api/middleware/auth.go`**
  - `Auth(jwtService)` - JWT validation middleware
  - `RequireRole(role)` - Role-based access control
  - `RequireAdmin()` - Admin-only access
  - `GetUserFromContext(ctx)` - Extract user claims from context

- **`internal/api/middleware/ratelimit.go`**
  - `AuthRateLimiter()` - 5 req/min for auth endpoints
  - `GlobalRateLimiter()` - 100 req/min general
  - `StrictRateLimiter()` - 3 req/min for sensitive ops

## Integration Steps

### 1. Wire up in `cmd/server/main.go`:

```go
import (
	"github.com/phillipboles/aci-backend/internal/api/handlers"
	"github.com/phillipboles/aci-backend/internal/api/middleware"
	"github.com/phillipboles/aci-backend/internal/repository/postgres"
	"github.com/phillipboles/aci-backend/internal/service"
	"github.com/phillipboles/aci-backend/internal/pkg/jwt"
)

func main() {
	// ... existing DB setup ...

	// Create JWT service
	jwtConfig := &jwt.Config{
		PrivateKeyPath: os.Getenv("JWT_PRIVATE_KEY_PATH"),
		PublicKeyPath:  os.Getenv("JWT_PUBLIC_KEY_PATH"),
		Issuer:         os.Getenv("JWT_ISSUER"),
	}
	jwtService, err := jwt.NewService(jwtConfig)
	if err != nil {
		log.Fatal("Failed to create JWT service:", err)
	}

	// Create repositories
	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)

	// Create service
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService)

	// Create handler
	authHandler := handlers.NewAuthHandler(authService)

	// Add to router (in internal/api/router.go or main.go)
	// See router integration below
}
```

### 2. Update `internal/api/router.go`:

```go
func SetupRoutes(
	r chi.Router,
	authHandler *handlers.AuthHandler,
	jwtService jwt.Service,
	// ... other handlers ...
) {
	// Public routes with rate limiting
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRateLimiter())

		r.Post("/v1/auth/register", authHandler.Register)
		r.Post("/v1/auth/login", authHandler.Login)
		r.Post("/v1/auth/refresh", authHandler.Refresh)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtService))

		r.Post("/v1/auth/logout", authHandler.Logout)

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin())

			// Admin endpoints here
		})
	})
}
```

### 3. Required Environment Variables:

```bash
# JWT Keys (generate with openssl)
JWT_PRIVATE_KEY_PATH=/path/to/private.pem
JWT_PUBLIC_KEY_PATH=/path/to/public.pem
JWT_ISSUER=aci-backend

# Generate keys with:
# openssl genrsa -out private.pem 2048
# openssl rsa -in private.pem -pubout -out public.pem
```

### 4. Database Migration (if not exists):

```sql
-- Refresh tokens table should already exist based on domain model
-- Verify with:
SELECT * FROM information_schema.tables
WHERE table_name = 'refresh_tokens';

-- If needed:
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    last_used_at TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,

    INDEX idx_token_hash (token_hash),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
);
```

## API Endpoint Examples

### Register
```bash
POST /v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123",
  "name": "John Doe"
}

# Response: 201 Created
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "user",
      "email_verified": false
    },
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "expires_at": "2024-12-11T15:30:00Z"
  }
}
```

### Login
```bash
POST /v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123"
}

# Response: 200 OK
# Same format as Register
```

### Refresh
```bash
POST /v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}

# Response: 200 OK
{
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",  # New token (rotation)
    "expires_at": "2024-12-11T15:45:00Z"
  }
}
```

### Logout
```bash
POST /v1/auth/logout
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "refresh_token": "eyJhbGc...",  # Optional if all_devices=false
  "all_devices": false
}

# Response: 200 OK
{
  "data": null,
  "message": "Logged out successfully"
}
```

### Protected Endpoint Usage
```bash
GET /v1/some-protected-resource
Authorization: Bearer <access_token>

# Handler can access user with:
claims, ok := middleware.GetUserFromContext(r.Context())
if ok {
  userID := claims.UserID
  userEmail := claims.Email
  userRole := claims.Role
}
```

## Security Features Implemented

1. **Password Security**
   - Bcrypt hashing with cost factor 12
   - Minimum 8 characters
   - Requires uppercase, lowercase, and digit

2. **Token Security**
   - RS256 JWT signing
   - 15-minute access token expiry
   - 7-day refresh token expiry
   - Token rotation on refresh
   - SHA-256 hashing of refresh tokens in database

3. **Rate Limiting**
   - 5 requests/minute on auth endpoints
   - IP-based limiting

4. **Error Handling**
   - Generic "invalid credentials" to prevent email enumeration
   - Proper error wrapping with context
   - Domain-specific errors (ValidationError, ConflictError, etc.)

## Testing Checklist

- [ ] Register new user with valid data
- [ ] Register fails with weak password
- [ ] Register fails with invalid email
- [ ] Register fails with existing email
- [ ] Login with valid credentials
- [ ] Login fails with wrong password
- [ ] Login fails with non-existent email
- [ ] Refresh token works and rotates
- [ ] Refresh fails with invalid token
- [ ] Refresh fails with revoked token
- [ ] Logout revokes single token
- [ ] LogoutAll revokes all user tokens
- [ ] Auth middleware blocks unauthenticated requests
- [ ] Auth middleware allows valid JWT
- [ ] RequireAdmin blocks non-admin users
- [ ] Rate limiting triggers after 5 auth requests
- [ ] Concurrent login sessions work independently

## Notes

- **User Type Mismatch**: The codebase has both `domain.User` and `entities.User`. The repository interface defines `domain.User` but the implementation uses `entities.User`. Created `UserRepoInterface` adapter to work around this.
- **Token Hashing**: Refresh tokens are hashed with SHA-256 before storage - only the hash is stored in database.
- **Token Rotation**: On refresh, old token is revoked and new pair is issued for security.
- **Context Keys**: Used `authContextKey` type to avoid collision with `requestid` middleware's `contextKey`.

## Future Enhancements

1. Email verification flow
2. Password reset flow
3. Two-factor authentication (2FA)
4. OAuth2 social login
5. Session management with Redis
6. Audit logging for auth events
7. Brute force protection with exponential backoff
8. Device fingerprinting
9. Suspicious activity detection
10. Token blacklisting for immediate revocation
