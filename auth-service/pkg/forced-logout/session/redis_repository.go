package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kart/k8s-agent/auth-service/pkg/types"
	"github.com/redis/go-redis/v9"
)

// RedisRepository implements Repository using Redis
type RedisRepository struct {
	client *redis.Client
	ttl    time.Duration // JWT expiration time
}

// NewRedisRepository creates a new Redis-based session repository
func NewRedisRepository(client *redis.Client, jwtExpiresHours int) *RedisRepository {
	return &RedisRepository{
		client: client,
		ttl:    time.Duration(jwtExpiresHours) * time.Hour,
	}
}

// StoreSession stores a new session in Redis
func (r *RedisRepository) StoreSession(ctx context.Context, session *types.SessionInfo) error {
	pipe := r.client.Pipeline()

	// Store session metadata as hash
	sessionKey := fmt.Sprintf("session:%s", session.JTI)
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	pipe.Set(ctx, sessionKey, sessionData, r.ttl)

	// Add to user's session sorted set (score = login timestamp)
	userSessionsKey := fmt.Sprintf("user:sessions:%s", session.UserID)
	pipe.ZAdd(ctx, userSessionsKey, redis.Z{
		Score:  float64(session.LoginAt.Unix()),
		Member: session.JTI,
	})
	pipe.Expire(ctx, userSessionsKey, r.ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("store session: %w", err)
	}

	return nil
}

// GetSession retrieves session metadata by JTI
func (r *RedisRepository) GetSession(ctx context.Context, jti string) (*types.SessionInfo, error) {
	sessionKey := fmt.Sprintf("session:%s", jti)
	data, err := r.client.Get(ctx, sessionKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	var session types.SessionInfo
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &session, nil
}

// ListUserSessions retrieves all active sessions for a user with pagination
func (r *RedisRepository) ListUserSessions(ctx context.Context, userID string, limit, offset int) ([]types.SessionInfo, int, error) {
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)

	// Get total count
	total, err := r.client.ZCard(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("count sessions: %w", err)
	}

	// Get session JTIs with pagination (most recent first)
	jtis, err := r.client.ZRevRange(ctx, userSessionsKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("list session jtis: %w", err)
	}

	// Retrieve session metadata for each JTI
	sessions := make([]types.SessionInfo, 0, len(jtis))
	for _, jti := range jtis {
		session, err := r.GetSession(ctx, jti)
		if err != nil {
			// Session may have expired, skip it
			continue
		}
		sessions = append(sessions, *session)
	}

	return sessions, int(total), nil
}

// RevokeSession adds session to blacklist and removes from active sets
func (r *RedisRepository) RevokeSession(ctx context.Context, jti, userID, revokedBy, reason, eventID string) error {
	pipe := r.client.Pipeline()

	// Add to blacklist
	revokedKey := fmt.Sprintf("revoked:%s", jti)
	revokedData := types.RevokedSession{
		JTI:       jti,
		UserID:    userID,
		RevokedAt: time.Now(),
		RevokedBy: revokedBy,
		Reason:    reason,
		EventID:   eventID,
	}
	data, err := json.Marshal(revokedData)
	if err != nil {
		return fmt.Errorf("marshal revoked session: %w", err)
	}
	pipe.Set(ctx, revokedKey, data, r.ttl)

	// Remove from user's active sessions
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	pipe.ZRem(ctx, userSessionsKey, jti)

	// Delete session metadata
	sessionKey := fmt.Sprintf("session:%s", jti)
	pipe.Del(ctx, sessionKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}

// IsRevoked checks if a session JTI is blacklisted
func (r *RedisRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	revokedKey := fmt.Sprintf("revoked:%s", jti)
	exists, err := r.client.Exists(ctx, revokedKey).Result()
	if err != nil {
		return false, fmt.Errorf("check revoked: %w", err)
	}
	return exists == 1, nil
}

// BulkRevokeSessions revokes multiple sessions in one operation using pipelining
func (r *RedisRepository) BulkRevokeSessions(ctx context.Context, jtis []string, userID, revokedBy, reason, eventID string) error {
	pipe := r.client.Pipeline()

	revokedData := types.RevokedSession{
		RevokedAt: time.Now(),
		RevokedBy: revokedBy,
		Reason:    reason,
		EventID:   eventID,
	}

	for _, jti := range jtis {
		// Add to blacklist
		revokedKey := fmt.Sprintf("revoked:%s", jti)
		revokedData.JTI = jti
		revokedData.UserID = userID
		data, err := json.Marshal(revokedData)
		if err != nil {
			return fmt.Errorf("marshal revoked session %s: %w", jti, err)
		}
		pipe.Set(ctx, revokedKey, data, r.ttl)

		// Remove from user's active sessions
		userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
		pipe.ZRem(ctx, userSessionsKey, jti)

		// Delete session metadata
		sessionKey := fmt.Sprintf("session:%s", jti)
		pipe.Del(ctx, sessionKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("bulk revoke sessions: %w", err)
	}

	return nil
}
