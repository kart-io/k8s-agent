package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTokenPair(t *testing.T) {
	secret := "test-secret"
	userID := "user-123"
	username := "testuser"
	expiresHours := 2

	tokenPair, err := GenerateTokenPair(userID, username, secret, expiresHours)
	require.NoError(t, err)
	require.NotNil(t, tokenPair)

	// Verify access token
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)

	// Verify access token expiration (should be ~2 hours)
	accessDuration := time.Until(tokenPair.AccessTokenExpiresAt)
	assert.True(t, accessDuration > 1*time.Hour && accessDuration < 3*time.Hour,
		"Access token should expire in ~2 hours")

	// Verify refresh token expiration (should be ~7 days)
	refreshDuration := time.Until(tokenPair.RefreshTokenExpiresAt)
	assert.True(t, refreshDuration > 6*24*time.Hour && refreshDuration < 8*24*time.Hour,
		"Refresh token should expire in ~7 days")
}

func TestValidateAccessToken(t *testing.T) {
	secret := "test-secret"
	userID := "user-123"
	username := "testuser"
	expiresHours := 2

	tokenPair, err := GenerateTokenPair(userID, username, secret, expiresHours)
	require.NoError(t, err)

	// Validate access token
	claims, err := ValidateToken(tokenPair.AccessToken, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
	assert.NotEmpty(t, claims.ID) // JTI should be set
}

func TestValidateRefreshToken(t *testing.T) {
	secret := "test-secret"
	userID := "user-123"
	username := "testuser"
	expiresHours := 2

	tokenPair, err := GenerateTokenPair(userID, username, secret, expiresHours)
	require.NoError(t, err)

	// Validate refresh token
	claims, err := ValidateRefreshToken(tokenPair.RefreshToken, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, TokenTypeRefresh, claims.TokenType)
	assert.NotEmpty(t, claims.ID) // JTI should be set
}

func TestValidateRefreshToken_RejectsAccessToken(t *testing.T) {
	secret := "test-secret"
	userID := "user-123"
	username := "testuser"
	expiresHours := 2

	tokenPair, err := GenerateTokenPair(userID, username, secret, expiresHours)
	require.NoError(t, err)

	// Try to validate access token as refresh token (should fail)
	_, err = ValidateRefreshToken(tokenPair.AccessToken, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected refresh")
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	secret := "test-secret"
	wrongSecret := "wrong-secret"
	userID := "user-123"
	username := "testuser"
	expiresHours := 2

	tokenPair, err := GenerateTokenPair(userID, username, secret, expiresHours)
	require.NoError(t, err)

	// Try to validate with wrong secret
	_, err = ValidateToken(tokenPair.AccessToken, wrongSecret)
	assert.Error(t, err)
}

func TestValidateToken_MalformedToken(t *testing.T) {
	secret := "test-secret"
	malformedToken := "not.a.valid.jwt.token"

	_, err := ValidateToken(malformedToken, secret)
	assert.Error(t, err)
}

func TestGenerateToken_BackwardCompatibility(t *testing.T) {
	// Test the legacy GenerateToken function still works
	secret := "test-secret"
	userID := "user-123"
	username := "testuser"
	expiresHours := 2

	tokenString, expiresAt, err := GenerateToken(userID, username, secret, expiresHours)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify expiration time
	duration := time.Until(expiresAt)
	assert.True(t, duration > 1*time.Hour && duration < 3*time.Hour)

	// Validate the token
	claims, err := ValidateToken(tokenString, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
}
