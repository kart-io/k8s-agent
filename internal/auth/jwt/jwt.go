package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// TokenTypeAccess represents access token type
	TokenTypeAccess = "access"
	// TokenTypeRefresh represents refresh token type
	TokenTypeRefresh = "refresh"
)

// Claims represents JWT custom claims.
type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"` // access or refresh
	jwt.RegisteredClaims
}

// TokenPair represents a pair of access and refresh tokens.
type TokenPair struct {
	AccessToken          string    `json:"access_token"`
	RefreshToken         string    `json:"refresh_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

// GenerateToken generates a JWT token for a user.
func GenerateToken(userID, username, secret string, expiresHours int) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(expiresHours) * time.Hour)

	claims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(), // JTI for token tracking
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateTokenPair generates both access and refresh tokens for a user.
// Access token expires in expiresHours, refresh token expires in 7 days.
func GenerateTokenPair(userID, username, secret string, expiresHours int) (*TokenPair, error) {
	now := time.Now()

	// Generate access token (e.g., 2 hours)
	accessExpiresAt := now.Add(time.Duration(expiresHours) * time.Hour)
	accessClaims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token (7 days)
	refreshExpiresAt := now.Add(7 * 24 * time.Hour)
	refreshClaims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:           accessTokenString,
		RefreshToken:          refreshTokenString,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

// ValidateToken validates a JWT token and returns claims.
func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshToken validates a refresh token specifically.
func ValidateRefreshToken(tokenString, secret string) (*Claims, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return nil, err
	}

	// Verify it's a refresh token
	if claims.TokenType != TokenTypeRefresh {
		return nil, fmt.Errorf("invalid token type: expected refresh, got %s", claims.TokenType)
	}

	return claims, nil
}
