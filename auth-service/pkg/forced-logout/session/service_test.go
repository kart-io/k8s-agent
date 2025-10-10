package session

import (
	"context"
	"errors"
	"testing"

	"github.com/kart/k8s-agent/auth-service/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) StoreSession(ctx context.Context, session *types.SessionInfo) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) GetSession(ctx context.Context, jti string) (*types.SessionInfo, error) {
	args := m.Called(ctx, jti)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.SessionInfo), args.Error(1)
}

func (m *MockRepository) ListUserSessions(ctx context.Context, userID string, limit, offset int) ([]types.SessionInfo, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]types.SessionInfo), args.Int(1), args.Error(2)
}

func (m *MockRepository) RevokeSession(ctx context.Context, jti, userID, revokedBy, reason, eventID string) error {
	args := m.Called(ctx, jti, userID, revokedBy, reason, eventID)
	return args.Error(0)
}

func (m *MockRepository) BulkRevokeSessions(ctx context.Context, jtis []string, userID, revokedBy, reason, eventID string) error {
	args := m.Called(ctx, jtis, userID, revokedBy, reason, eventID)
	return args.Error(0)
}

func (m *MockRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

func TestNewService(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	assert.NotNil(t, svc)
	assert.Equal(t, repo, svc.repo)
}

func TestCreateSession_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	session := &types.SessionInfo{
		JTI:       "test-jti",
		UserID:    "user-123",
		Username:  "testuser",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0.4472.124",
		IPAddress: "192.168.1.1",
	}

	repo.On("StoreSession", ctx, session).Return(nil)

	err := svc.CreateSession(ctx, session)

	assert.NoError(t, err)
	assert.Equal(t, "desktop", session.DeviceType)
	assert.Contains(t, session.DeviceName, "Chrome")
	assert.Contains(t, session.DeviceName, "Windows")
	repo.AssertExpectations(t)
}

func TestCreateSession_MissingJTI(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	session := &types.SessionInfo{
		UserID: "user-123",
	}

	err := svc.CreateSession(ctx, session)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "jti and user_id are required")
}

func TestCreateSession_MissingUserID(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	session := &types.SessionInfo{
		JTI: "test-jti",
	}

	err := svc.CreateSession(ctx, session)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "jti and user_id are required")
}

func TestGetSession_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	expectedSession := &types.SessionInfo{
		JTI:      "test-jti",
		UserID:   "user-123",
		Username: "testuser",
	}

	repo.On("GetSession", ctx, "test-jti").Return(expectedSession, nil)

	session, err := svc.GetSession(ctx, "test-jti")

	assert.NoError(t, err)
	assert.Equal(t, expectedSession, session)
	repo.AssertExpectations(t)
}

func TestGetUserSessions_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	sessions := []types.SessionInfo{
		{JTI: "jti-1", UserID: "user-123", Username: "testuser"},
		{JTI: "jti-2", UserID: "user-123", Username: "testuser"},
	}

	repo.On("ListUserSessions", ctx, "user-123", 50, 0).Return(sessions, 2, nil)

	resp, err := svc.GetUserSessions(ctx, "user-123", 50, 0)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "user-123", resp.UserID)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Sessions, 2)
	assert.Equal(t, 50, resp.Pagination.Limit)
	assert.False(t, resp.Pagination.HasMore)
	repo.AssertExpectations(t)
}

func TestGetUserSessions_DefaultLimit(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("ListUserSessions", ctx, "user-123", 50, 0).Return([]types.SessionInfo{}, 0, nil)

	resp, err := svc.GetUserSessions(ctx, "user-123", 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, 50, resp.Pagination.Limit) // Default limit applied
	repo.AssertExpectations(t)
}

func TestGetUserSessions_MaxLimit(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("ListUserSessions", ctx, "user-123", 100, 0).Return([]types.SessionInfo{}, 0, nil)

	resp, err := svc.GetUserSessions(ctx, "user-123", 200, 0)

	assert.NoError(t, err)
	assert.Equal(t, 100, resp.Pagination.Limit) // Max limit enforced
	repo.AssertExpectations(t)
}

func TestGetUserSessions_Pagination(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	sessions := []types.SessionInfo{
		{JTI: "jti-3", UserID: "user-123", Username: "testuser"},
		{JTI: "jti-4", UserID: "user-123", Username: "testuser"},
	}

	repo.On("ListUserSessions", ctx, "user-123", 2, 2).Return(sessions, 5, nil)

	resp, err := svc.GetUserSessions(ctx, "user-123", 2, 2)

	assert.NoError(t, err)
	assert.Equal(t, 5, resp.Total)
	assert.Len(t, resp.Sessions, 2)
	assert.Equal(t, 2, resp.Pagination.Offset)
	assert.True(t, resp.Pagination.HasMore) // offset(2) + len(2) < total(5)
	repo.AssertExpectations(t)
}

func TestValidateSession_Valid(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	session := &types.SessionInfo{JTI: "test-jti", UserID: "user-123"}

	repo.On("IsRevoked", ctx, "test-jti").Return(false, nil)
	repo.On("GetSession", ctx, "test-jti").Return(session, nil)

	valid, err := svc.ValidateSession(ctx, "test-jti")

	assert.NoError(t, err)
	assert.True(t, valid)
	repo.AssertExpectations(t)
}

func TestValidateSession_Revoked(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("IsRevoked", ctx, "test-jti").Return(true, nil)

	valid, err := svc.ValidateSession(ctx, "test-jti")

	assert.NoError(t, err)
	assert.False(t, valid)
	repo.AssertExpectations(t)
}

func TestValidateSession_NotFound(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("IsRevoked", ctx, "test-jti").Return(false, nil)
	repo.On("GetSession", ctx, "test-jti").Return(nil, errors.New("not found"))

	valid, err := svc.ValidateSession(ctx, "test-jti")

	assert.NoError(t, err)
	assert.False(t, valid)
	repo.AssertExpectations(t)
}

func TestTerminateSession_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	session := &types.SessionInfo{
		JTI:    "test-jti",
		UserID: "user-123",
	}

	repo.On("GetSession", ctx, "test-jti").Return(session, nil)
	repo.On("RevokeSession", ctx, "test-jti", "user-123", "admin-456", "security_breach", "event-789").Return(nil)

	err := svc.TerminateSession(ctx, "test-jti", "user-123", "admin-456", "security_breach", "event-789")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestTerminateSession_NotFound(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("GetSession", ctx, "test-jti").Return(nil, errors.New("not found"))

	err := svc.TerminateSession(ctx, "test-jti", "user-123", "admin-456", "security_breach", "event-789")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
	repo.AssertExpectations(t)
}

func TestTerminateSession_WrongUser(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	session := &types.SessionInfo{
		JTI:    "test-jti",
		UserID: "different-user",
	}

	repo.On("GetSession", ctx, "test-jti").Return(session, nil)

	err := svc.TerminateSession(ctx, "test-jti", "user-123", "admin-456", "security_breach", "event-789")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session belongs to different user")
	repo.AssertExpectations(t)
}

func TestTerminateUserSessions_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	sessions := []types.SessionInfo{
		{JTI: "jti-1", UserID: "user-123"},
		{JTI: "jti-2", UserID: "user-123"},
		{JTI: "jti-3", UserID: "user-123"},
	}

	repo.On("ListUserSessions", ctx, "user-123", 1000, 0).Return(sessions, 3, nil)
	repo.On("BulkRevokeSessions", ctx, []string{"jti-1", "jti-2", "jti-3"}, "user-123", "admin-456", "policy_violation", "event-789").Return(nil)

	count, err := svc.TerminateUserSessions(ctx, "user-123", "admin-456", "policy_violation", "event-789")

	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	repo.AssertExpectations(t)
}

func TestTerminateUserSessions_NoSessions(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("ListUserSessions", ctx, "user-123", 1000, 0).Return([]types.SessionInfo{}, 0, nil)

	count, err := svc.TerminateUserSessions(ctx, "user-123", "admin-456", "policy_violation", "event-789")

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	repo.AssertExpectations(t)
}

func TestDetectDeviceType(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  string
	}{
		{
			name:      "Desktop Chrome",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0",
			expected:  "desktop",
		},
		{
			name:      "Mobile iPhone",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
			expected:  "mobile",
		},
		{
			name:      "Mobile Android",
			userAgent: "Mozilla/5.0 (Linux; Android 10) Mobile Safari/537.36",
			expected:  "mobile",
		},
		{
			name:      "Tablet iPad",
			userAgent: "Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X)",
			expected:  "tablet",
		},
		{
			name:      "Desktop Firefox",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
			expected:  "desktop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectDeviceType(tt.userAgent)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseDeviceName(t *testing.T) {
	tests := []struct {
		name            string
		userAgent       string
		expectedBrowser string
		expectedOS      string
	}{
		{
			name:            "Windows Chrome",
			userAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expectedBrowser: "Chrome",
			expectedOS:      "Windows",
		},
		{
			name:            "macOS Safari",
			userAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Safari/605.1.15",
			expectedBrowser: "Safari",
			expectedOS:      "macOS",
		},
		{
			name:            "Linux Firefox",
			userAgent:       "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
			expectedBrowser: "Firefox",
			expectedOS:      "Linux",
		},
		{
			name:            "Android Mobile",
			userAgent:       "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36",
			expectedBrowser: "Chrome",
			expectedOS:      "Android",
		},
		{
			name:            "iOS iPhone",
			userAgent:       "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			expectedBrowser: "Safari",
			expectedOS:      "iOS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDeviceName(tt.userAgent)
			assert.Contains(t, result, tt.expectedBrowser)
			assert.Contains(t, result, tt.expectedOS)
		})
	}
}
