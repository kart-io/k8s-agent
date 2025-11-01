package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kart-io/logger/core"
	"github.com/redis/go-redis/v9"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/pkg/types"
)

// Lua脚本用于原子性地执行rate limit操作
var rateLimitScript = redis.NewScript(`
	local current = redis.call("INCR", KEYS[1])
	if current == 1 then
		redis.call("EXPIRE", KEYS[1], ARGV[1])
	end
	return current
`)

// RedisStore implements caching using Redis
type RedisStore struct {
	*db.RedisClient // Embed common Redis client
	logger          core.Logger
}

// NewRedisStore creates a new Redis store
func NewRedisStore(config types.RedisConfig, log core.Logger) (*RedisStore, error) {
	// Create Redis client using common package with Options pattern
	redisClient, err := db.NewRedis(log,
		db.WithAddr(config.Addr),
		db.WithRedisPassword(config.Password),
		db.WithRedisDB(config.DB),
		db.WithRedisPoolSize(config.PoolSize),
		db.WithRedisMinIdleConns(config.MinIdleConns),
		db.WithRedisDialTimeout(config.DialTimeout),
		db.WithRedisReadTimeout(config.ReadTimeout),
		db.WithRedisWriteTimeout(config.WriteTimeout),
	)
	if err != nil {
		return nil, err
	}

	store := &RedisStore{
		RedisClient: redisClient,
		logger:      log.With("component", "redis"),
	}

	store.logger.Infow("Redis store initialized", "addr", config.Addr)

	return store, nil
}

// Agent cache operations

// CacheAgent caches agent information
func (s *RedisStore) CacheAgent(ctx context.Context, agent *types.Agent, ttl time.Duration) error {
	key := s.agentKey(agent.ID)
	data, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent: %w", err)
	}

	return s.Client.Set(ctx, key, data, ttl).Err()
}

// GetCachedAgent retrieves cached agent information
func (s *RedisStore) GetCachedAgent(ctx context.Context, id string) (*types.Agent, error) {
	key := s.agentKey(id)
	data, err := s.Client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var agent types.Agent
	if err := json.Unmarshal(data, &agent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent: %w", err)
	}

	return &agent, nil
}

// DeleteCachedAgent removes cached agent information
func (s *RedisStore) DeleteCachedAgent(ctx context.Context, id string) error {
	key := s.agentKey(id)
	return s.Client.Del(ctx, key).Err()
}

// Agent status tracking

// SetAgentOnline marks agent as online
func (s *RedisStore) SetAgentOnline(ctx context.Context, agentID string, ttl time.Duration) error {
	key := s.agentStatusKey(agentID)
	return s.Client.Set(ctx, key, "online", ttl).Err()
}

// IsAgentOnline checks if agent is online
func (s *RedisStore) IsAgentOnline(ctx context.Context, agentID string) (bool, error) {
	key := s.agentStatusKey(agentID)
	result, err := s.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetOnlineAgents returns list of online agent IDs
func (s *RedisStore) GetOnlineAgents(ctx context.Context) ([]string, error) {
	pattern := "agent:status:*"
	var agentIDs []string

	iter := s.Client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Extract agent ID from key "agent:status:{id}"
		if len(key) > 13 {
			agentID := key[13:]
			agentIDs = append(agentIDs, agentID)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return agentIDs, nil
}

// Command queue operations

// EnqueueCommand adds a command to the cluster's command queue
func (s *RedisStore) EnqueueCommand(ctx context.Context, clusterID string, cmd *types.Command) error {
	key := s.commandQueueKey(clusterID)
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	return s.Client.LPush(ctx, key, data).Err()
}

// DequeueCommand removes and returns a command from the queue
func (s *RedisStore) DequeueCommand(ctx context.Context, clusterID string, timeout time.Duration) (*types.Command, error) {
	key := s.commandQueueKey(clusterID)
	result, err := s.Client.BRPop(ctx, timeout, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Queue empty
		}
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected result format")
	}

	var cmd types.Command
	if err := json.Unmarshal([]byte(result[1]), &cmd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command: %w", err)
	}

	return &cmd, nil
}

// GetCommandQueueLength returns the length of command queue
func (s *RedisStore) GetCommandQueueLength(ctx context.Context, clusterID string) (int64, error) {
	key := s.commandQueueKey(clusterID)
	return s.Client.LLen(ctx, key).Result()
}

// Metrics aggregation

// IncrementEventCounter increments event counter
func (s *RedisStore) IncrementEventCounter(ctx context.Context, clusterID, severity string) error {
	key := s.eventCounterKey(clusterID, severity)
	return s.Client.Incr(ctx, key).Err()
}

// GetEventCount returns event count
func (s *RedisStore) GetEventCount(ctx context.Context, clusterID, severity string) (int64, error) {
	key := s.eventCounterKey(clusterID, severity)
	return s.Client.Get(ctx, key).Int64()
}

// ResetEventCounters resets event counters
func (s *RedisStore) ResetEventCounters(ctx context.Context) error {
	pattern := "event:count:*"
	iter := s.Client.Scan(ctx, 0, pattern, 100).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.Client.Del(ctx, keys...).Err()
	}

	return nil
}

// Session management

// CreateSession creates a new session
func (s *RedisStore) CreateSession(ctx context.Context, sessionID, userID string, ttl time.Duration) error {
	key := s.sessionKey(sessionID)
	return s.Client.Set(ctx, key, userID, ttl).Err()
}

// ValidateSession validates a session
func (s *RedisStore) ValidateSession(ctx context.Context, sessionID string) (string, error) {
	key := s.sessionKey(sessionID)
	userID, err := s.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("session not found")
		}
		return "", err
	}
	return userID, nil
}

// DeleteSession deletes a session
func (s *RedisStore) DeleteSession(ctx context.Context, sessionID string) error {
	key := s.sessionKey(sessionID)
	return s.Client.Del(ctx, key).Err()
}

// Rate limiting

// CheckRateLimit checks if request is within rate limit
// 使用Lua脚本保证Incr和Expire操作的原子性，避免竞态条件
func (s *RedisStore) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	rateLimitKey := s.rateLimitKey(key)

	// 使用Lua脚本原子性地执行Incr和Expire
	count, err := rateLimitScript.Run(ctx, s.Client, []string{rateLimitKey}, int(window.Seconds())).Int64()
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit script: %w", err)
	}

	return count <= limit, nil
}

// Distributed lock

// AcquireLock acquires a distributed lock
func (s *RedisStore) AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, error) {
	key := s.lockKey(lockKey)
	return s.Client.SetNX(ctx, key, "locked", ttl).Result()
}

// ReleaseLock releases a distributed lock
func (s *RedisStore) ReleaseLock(ctx context.Context, lockKey string) error {
	key := s.lockKey(lockKey)
	return s.Client.Del(ctx, key).Err()
}

// Key generation helpers

func (s *RedisStore) agentKey(id string) string {
	return fmt.Sprintf("agent:%s", id)
}

func (s *RedisStore) agentStatusKey(id string) string {
	return fmt.Sprintf("agent:status:%s", id)
}

func (s *RedisStore) commandQueueKey(clusterID string) string {
	return fmt.Sprintf("command:queue:%s", clusterID)
}

func (s *RedisStore) eventCounterKey(clusterID, severity string) string {
	return fmt.Sprintf("event:count:%s:%s", clusterID, severity)
}

func (s *RedisStore) sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func (s *RedisStore) rateLimitKey(key string) string {
	return fmt.Sprintf("ratelimit:%s", key)
}

func (s *RedisStore) lockKey(lockKey string) string {
	return fmt.Sprintf("lock:%s", lockKey)
}
