package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test secret key
const testSecret = "test-secret-key"

// Mock session validator
type mockSessionValidator struct {
	isValid bool
	err     error
}

func (m *mockSessionValidator) ValidateSession(ctx context.Context, jti string) (bool, error) {
	return m.isValid, m.err
}

// Helper function to create test JWT middleware
func newTestJWTMiddleware(validator SessionValidator) *JWTMiddleware {
	return NewJWTMiddleware(&JWTConfig{
		Secret:           []byte(testSecret),
		SessionValidator: validator,
	})
}

// Helper function to generate test token
func generateTestToken(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))
	return tokenString
}

// Test valid token passes authentication
func TestJWTMiddleware_Auth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := newTestJWTMiddleware(nil)

	// Generate valid token
	claims := jwt.MapClaims{
		"user_id":  "user123",
		"username": "testuser",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := generateTestToken(claims)

	// Setup test server
	router := gin.New()
	router.Use(middleware.Auth())
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
		})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user123")
	assert.Contains(t, w.Body.String(), "testuser")
}

// Test invalid token is rejected
func TestJWTMiddleware_Auth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := newTestJWTMiddleware(nil)

	// Setup test server
	router := gin.New()
	router.Use(middleware.Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// Test expired token is rejected
func TestJWTMiddleware_Auth_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := newTestJWTMiddleware(nil)

	// Generate expired token
	claims := jwt.MapClaims{
		"user_id":  "user123",
		"username": "testuser",
		"exp":      time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := generateTestToken(claims)

	// Setup test server
	router := gin.New()
	router.Use(middleware.Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test missing authorization header is rejected
func TestJWTMiddleware_Auth_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := newTestJWTMiddleware(nil)

	// Setup test server
	router := gin.New()
	router.Use(middleware.Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request without Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing authorization header")
}

// Test wrong signing key is rejected
func TestJWTMiddleware_Auth_WrongSigningKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := newTestJWTMiddleware(nil)

	// Generate token with different secret
	claims := jwt.MapClaims{
		"user_id":  "user123",
		"username": "testuser",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret-key"))

	// Setup test server
	router := gin.New()
	router.Use(middleware.Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test missing user_id in claims is rejected
func TestJWTMiddleware_Auth_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := newTestJWTMiddleware(nil)

	// Generate token without user_id
	claims := jwt.MapClaims{
		"username": "testuser",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := generateTestToken(claims)

	// Setup test server
	router := gin.New()
	router.Use(middleware.Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid user_id in token")
}

// Test session validation - valid session
func TestJWTMiddleware_Auth_SessionValidation_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock validator that returns valid
	validator := &mockSessionValidator{isValid: true, err: nil}
	middleware := newTestJWTMiddleware(validator)

	// Generate token with JTI
	claims := jwt.MapClaims{
		"user_id":  "user123",
		"username": "testuser",
		"jti":      "session123",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := generateTestToken(claims)

	// Setup test server
	router := gin.New()
	router.Use(middleware.Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test session validation - revoked session
func TestJWTMiddleware_Auth_SessionValidation_Revoked(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock validator that returns invalid (revoked)
	validator := &mockSessionValidator{isValid: false, err: nil}
	middleware := newTestJWTMiddleware(validator)

	// Generate token with JTI
	claims := jwt.MapClaims{
		"user_id":  "user123",
		"username": "testuser",
		"jti":      "session123",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := generateTestToken(claims)

	// Setup test server
	router := gin.New()
	router.Use(middleware.Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "session has been terminated")
}

// Test session validation - validator error (SECURITY FIX: should reject)
func TestJWTMiddleware_Auth_SessionValidation_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock validator that returns error
	validator := &mockSessionValidator{isValid: false, err: assert.AnError}
	middleware := newTestJWTMiddleware(validator)

	// Generate token with JTI
	claims := jwt.MapClaims{
		"user_id":  "user123",
		"username": "testuser",
		"jti":      "session123",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := generateTestToken(claims)

	// Setup test server
	router := gin.New()
	router.Use(middleware.Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions - SECURITY FIX: should reject on validation error
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Session validation failed")
}

// Test OptionalAuth middleware - with valid token
func TestJWTMiddleware_OptionalAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := newTestJWTMiddleware(nil)

	// Generate valid token
	claims := jwt.MapClaims{
		"user_id":  "user123",
		"username": "testuser",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := generateTestToken(claims)

	// Setup test server
	router := gin.New()
	router.Use(middleware.OptionalAuth())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			c.JSON(http.StatusOK, gin.H{"user_id": userID, "authenticated": true})
		} else {
			c.JSON(http.StatusOK, gin.H{"authenticated": false})
		}
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user123")
	assert.Contains(t, w.Body.String(), "true")
}

// Test OptionalAuth middleware - without token (should allow)
func TestJWTMiddleware_OptionalAuth_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := newTestJWTMiddleware(nil)

	// Setup test server
	router := gin.New()
	router.Use(middleware.OptionalAuth())
	router.GET("/test", func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if exists {
			c.JSON(http.StatusOK, gin.H{"authenticated": true})
		} else {
			c.JSON(http.StatusOK, gin.H{"authenticated": false})
		}
	})

	// Make request without token
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"authenticated":false`)
}

// Test GenerateToken function
func TestJWTMiddleware_GenerateToken(t *testing.T) {
	middleware := newTestJWTMiddleware(nil)

	claims := jwt.MapClaims{
		"user_id":  "user123",
		"username": "testuser",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	}

	token, err := middleware.GenerateToken(claims)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	parsedClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, "user123", parsedClaims["user_id"])
}

// Test RefreshToken function
func TestJWTMiddleware_RefreshToken(t *testing.T) {
	middleware := newTestJWTMiddleware(nil)

	// Generate original token
	originalClaims := jwt.MapClaims{
		"user_id":  "user123",
		"username": "testuser",
		"jti":      "session123",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	originalToken := generateTestToken(originalClaims)

	// Refresh token
	newToken, err := middleware.RefreshToken(originalToken, 2)

	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, originalToken, newToken)

	// Verify new token
	parsedToken, err := jwt.Parse(newToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	newClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, "user123", newClaims["user_id"])
	assert.Equal(t, "session123", newClaims["jti"]) // JTI should be preserved
}
