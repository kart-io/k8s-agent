package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	redisstorage "github.com/kart-io/k8s-agent/common/storage/redis"
	"github.com/kart-io/logger/core"
)

// SessionManager Redis 会话管理器
// 提供会话的创建、读取、更新、删除和过期时间管理
type SessionManager struct {
	client *redisstorage.Client
	logger core.Logger
	prefix string
}

// NewSessionManager 创建会话管理器
//
// Parameters:
//   - client: Redis 客户端实例
//   - logger: 日志记录器
//   - prefix: 会话键前缀（用于命名空间隔离）
//
// Example usage:
//
//	sessionMgr := auth.NewSessionManager(client, logger, "myapp")
//	err := sessionMgr.Set(ctx, sessionID, sessionData, 1*time.Hour)
func NewSessionManager(client *redisstorage.Client, logger core.Logger, prefix string) *SessionManager {
	return &SessionManager{
		client: client,
		logger: logger.With("component", "session"),
		prefix: prefix,
	}
}

// sessionKey 生成会话键
func (sm *SessionManager) sessionKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s", sm.prefix, sessionID)
}

// userSessionsKey 生成用户会话列表键
func (sm *SessionManager) userSessionsKey(userID string) string {
	return fmt.Sprintf("%s:user:sessions:%s", sm.prefix, userID)
}

// Set 设置会话数据
// 会话数据会被序列化为 JSON 格式存储
//
// Parameters:
//   - ctx: Context for timeout control
//   - sessionID: 会话唯一标识符
//   - data: 会话数据（任意类型，会被 JSON 序列化）
//   - ttl: 会话过期时间
//
// Returns:
//   - error: 序列化失败或 Redis 操作失败时返回错误
func (sm *SessionManager) Set(ctx context.Context, sessionID string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	key := sm.sessionKey(sessionID)
	if err := sm.client.Client().Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session: %w", err)
	}

	sm.logger.Debugw("Session set",
		"session_id", sessionID,
		"ttl", ttl,
	)

	return nil
}

// Get 获取会话数据
// 将 JSON 格式的会话数据反序列化到 result 中
//
// Parameters:
//   - ctx: Context for timeout control
//   - sessionID: 会话唯一标识符
//   - result: 用于接收会话数据的指针（必须是指针类型）
//
// Returns:
//   - error: 会话不存在（redis.Nil）或反序列化失败时返回错误
func (sm *SessionManager) Get(ctx context.Context, sessionID string, result interface{}) error {
	key := sm.sessionKey(sessionID)
	data, err := sm.client.Client().Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("session not found: %s", sessionID)
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	if err := json.Unmarshal([]byte(data), result); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return nil
}

// Delete 删除会话
//
// Parameters:
//   - ctx: Context for timeout control
//   - sessionID: 会话唯一标识符
//
// Returns:
//   - error: Redis 操作失败时返回错误
func (sm *SessionManager) Delete(ctx context.Context, sessionID string) error {
	key := sm.sessionKey(sessionID)
	if err := sm.client.Client().Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	sm.logger.Debugw("Session deleted", "session_id", sessionID)
	return nil
}

// Exists 检查会话是否存在
//
// Parameters:
//   - ctx: Context for timeout control
//   - sessionID: 会话唯一标识符
//
// Returns:
//   - bool: 会话是否存在
//   - error: Redis 操作失败时返回错误
func (sm *SessionManager) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := sm.sessionKey(sessionID)
	count, err := sm.client.Client().Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}
	return count > 0, nil
}

// Refresh 刷新会话过期时间
// 重置会话的 TTL，延长会话有效期
//
// Parameters:
//   - ctx: Context for timeout control
//   - sessionID: 会话唯一标识符
//   - ttl: 新的过期时间
//
// Returns:
//   - error: Redis 操作失败时返回错误
func (sm *SessionManager) Refresh(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := sm.sessionKey(sessionID)
	if err := sm.client.Client().Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	sm.logger.Debugw("Session refreshed",
		"session_id", sessionID,
		"ttl", ttl,
	)

	return nil
}

// GetTTL 获取会话剩余时间
//
// Parameters:
//   - ctx: Context for timeout control
//   - sessionID: 会话唯一标识符
//
// Returns:
//   - time.Duration: 会话剩余有效时间，-2 表示键不存在，-1 表示没有设置过期时间
//   - error: Redis 操作失败时返回错误
func (sm *SessionManager) GetTTL(ctx context.Context, sessionID string) (time.Duration, error) {
	key := sm.sessionKey(sessionID)
	ttl, err := sm.client.Client().TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get session TTL: %w", err)
	}
	return ttl, nil
}

// AddUserSession 添加用户会话关联
// 将会话 ID 添加到用户的会话集合中，用于追踪用户的所有活跃会话
//
// Parameters:
//   - ctx: Context for timeout control
//   - userID: 用户唯一标识符
//   - sessionID: 会话唯一标识符
//   - ttl: 用户会话集合的过期时间（应该比单个会话的 TTL 长）
//
// Returns:
//   - error: Redis 操作失败时返回错误
func (sm *SessionManager) AddUserSession(ctx context.Context, userID, sessionID string, ttl time.Duration) error {
	key := sm.userSessionsKey(userID)

	// 添加到集合
	if err := sm.client.Client().SAdd(ctx, key, sessionID).Err(); err != nil {
		return fmt.Errorf("failed to add user session: %w", err)
	}

	// 设置过期时间
	if err := sm.client.Client().Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set user sessions expiry: %w", err)
	}

	sm.logger.Debugw("User session added",
		"user_id", userID,
		"session_id", sessionID,
	)

	return nil
}

// GetUserSessions 获取用户所有会话
// 返回用户所有活跃会话的 ID 列表
//
// Parameters:
//   - ctx: Context for timeout control
//   - userID: 用户唯一标识符
//
// Returns:
//   - []string: 会话 ID 列表
//   - error: Redis 操作失败时返回错误
func (sm *SessionManager) GetUserSessions(ctx context.Context, userID string) ([]string, error) {
	key := sm.userSessionsKey(userID)
	sessions, err := sm.client.Client().SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	return sessions, nil
}

// DeleteUserSession 删除用户会话关联
// 从用户的会话集合中移除指定的会话 ID
//
// Parameters:
//   - ctx: Context for timeout control
//   - userID: 用户唯一标识符
//   - sessionID: 会话唯一标识符
//
// Returns:
//   - error: Redis 操作失败时返回错误
func (sm *SessionManager) DeleteUserSession(ctx context.Context, userID, sessionID string) error {
	key := sm.userSessionsKey(userID)
	if err := sm.client.Client().SRem(ctx, key, sessionID).Err(); err != nil {
		return fmt.Errorf("failed to delete user session: %w", err)
	}

	sm.logger.Debugw("User session deleted",
		"user_id", userID,
		"session_id", sessionID,
	)

	return nil
}

// DeleteAllUserSessions 删除用户所有会话
// 删除用户的所有会话数据以及会话集合（用于强制登出）
//
// Parameters:
//   - ctx: Context for timeout control
//   - userID: 用户唯一标识符
//
// Returns:
//   - error: Redis 操作失败时返回错误
func (sm *SessionManager) DeleteAllUserSessions(ctx context.Context, userID string) error {
	// 获取所有会话 ID
	sessions, err := sm.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	// 使用 Pipeline 批量删除所有会话数据
	if len(sessions) > 0 {
		pipe := sm.client.Client().Pipeline()
		for _, sessionID := range sessions {
			key := sm.sessionKey(sessionID)
			pipe.Del(ctx, key)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete session data: %w", err)
		}
	}

	// 删除用户会话集合
	key := sm.userSessionsKey(userID)
	if err := sm.client.Client().Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete user sessions set: %w", err)
	}

	sm.logger.Infow("All user sessions deleted",
		"user_id", userID,
		"count", len(sessions),
	)

	return nil
}

// CountUserSessions 统计用户活跃会话数量
//
// Parameters:
//   - ctx: Context for timeout control
//   - userID: 用户唯一标识符
//
// Returns:
//   - int64: 活跃会话数量
//   - error: Redis 操作失败时返回错误
func (sm *SessionManager) CountUserSessions(ctx context.Context, userID string) (int64, error) {
	key := sm.userSessionsKey(userID)
	count, err := sm.client.Client().SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count user sessions: %w", err)
	}
	return count, nil
}

// Update 更新会话数据（保持原有的 TTL）
// 更新会话数据但不改变过期时间
//
// Parameters:
//   - ctx: Context for timeout control
//   - sessionID: 会话唯一标识符
//   - data: 新的会话数据
//
// Returns:
//   - error: 会话不存在或 Redis 操作失败时返回错误
func (sm *SessionManager) Update(ctx context.Context, sessionID string, data interface{}) error {
	key := sm.sessionKey(sessionID)

	// 先获取当前的 TTL
	ttl, err := sm.client.Client().TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get session TTL: %w", err)
	}

	// 检查会话是否存在
	if ttl < 0 {
		return fmt.Errorf("session not found or has no expiry: %s", sessionID)
	}

	// 序列化新数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// 使用原有的 TTL 设置新数据
	if err := sm.client.Client().Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	sm.logger.Debugw("Session updated",
		"session_id", sessionID,
		"remaining_ttl", ttl,
	)

	return nil
}
