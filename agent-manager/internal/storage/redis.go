package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)

// RedisStore implements caching using Redis
type RedisStore struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisStore creates a new Redis store
func NewRedisStore(config types.RedisConfig, log *zap.Logger) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	store := &RedisStore{
		client: client,
		logger: log.With(zap.String("component", "redis")),
	}

	store.logger.Info("Redis store initialized", zap.String("addr", config.Addr))

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

	return s.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedAgent retrieves cached agent information
func (s *RedisStore) GetCachedAgent(ctx context.Context, id string) (*types.Agent, error) {
	key := s.agentKey(id)
	data, err := s.client.Get(ctx, key).Bytes()
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
	return s.client.Del(ctx, key).Err()
}

// Agent status tracking

// SetAgentOnline marks agent as online
func (s *RedisStore) SetAgentOnline(ctx context.Context, agentID string, ttl time.Duration) error {
	key := s.agentStatusKey(agentID)
	return s.client.Set(ctx, key, "online", ttl).Err()
}

// IsAgentOnline checks if agent is online
func (s *RedisStore) IsAgentOnline(ctx context.Context, agentID string) (bool, error) {
	key := s.agentStatusKey(agentID)
	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetOnlineAgents returns list of online agent IDs
func (s *RedisStore) GetOnlineAgents(ctx context.Context) ([]string, error) {
	pattern := "agent:status:*"
	var agentIDs []string

	iter := s.client.Scan(ctx, 0, pattern, 100).Iterator()
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

	return s.client.LPush(ctx, key, data).Err()
}

// DequeueCommand removes and returns a command from the queue
func (s *RedisStore) DequeueCommand(ctx context.Context, clusterID string, timeout time.Duration) (*types.Command, error) {
	key := s.commandQueueKey(clusterID)
	result, err := s.client.BRPop(ctx, timeout, key).Result()
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
	return s.client.LLen(ctx, key).Result()
}

// Metrics aggregation

// IncrementEventCounter increments event counter
func (s *RedisStore) IncrementEventCounter(ctx context.Context, clusterID, severity string) error {
	key := s.eventCounterKey(clusterID, severity)
	return s.client.Incr(ctx, key).Err()
}

// GetEventCount returns event count
func (s *RedisStore) GetEventCount(ctx context.Context, clusterID, severity string) (int64, error) {
	key := s.eventCounterKey(clusterID, severity)
	return s.client.Get(ctx, key).Int64()
}

// ResetEventCounters resets event counters
func (s *RedisStore) ResetEventCounters(ctx context.Context) error {
	pattern := "event:count:*"
	iter := s.client.Scan(ctx, 0, pattern, 100).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Session management

// CreateSession creates a new session
func (s *RedisStore) CreateSession(ctx context.Context, sessionID, userID string, ttl time.Duration) error {
	key := s.sessionKey(sessionID)
	return s.client.Set(ctx, key, userID, ttl).Err()
}

// ValidateSession validates a session
func (s *RedisStore) ValidateSession(ctx context.Context, sessionID string) (string, error) {
	key := s.sessionKey(sessionID)
	userID, err := s.client.Get(ctx, key).Result()
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
	return s.client.Del(ctx, key).Err()
}

// Rate limiting

// CheckRateLimit checks if request is within rate limit
func (s *RedisStore) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	rateLimitKey := s.rateLimitKey(key)

	// Increment counter
	count, err := s.client.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		return false, err
	}

	// Set expiration on first request
	if count == 1 {
		s.client.Expire(ctx, rateLimitKey, window)
	}

	return count <= limit, nil
}

// Distributed lock

// AcquireLock acquires a distributed lock
func (s *RedisStore) AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, error) {
	key := s.lockKey(lockKey)
	return s.client.SetNX(ctx, key, "locked", ttl).Result()
}

// ReleaseLock releases a distributed lock
func (s *RedisStore) ReleaseLock(ctx context.Context, lockKey string) error {
	key := s.lockKey(lockKey)
	return s.client.Del(ctx, key).Err()
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

// Close closes the Redis connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// Health checks Redis health
func (s *RedisStore) Health(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}