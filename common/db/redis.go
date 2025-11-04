package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kart-io/logger/core"
)

// RedisOptions Redis 配置选项
type RedisOptions struct {
	addr         string
	password     string
	db           int
	poolSize     int
	minIdleConns int
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// RedisOption Redis 配置选项函数
type RedisOption func(*RedisOptions)

// WithAddr 设置 Redis 地址
func WithAddr(addr string) RedisOption {
	return func(o *RedisOptions) {
		o.addr = addr
	}
}

// WithRedisPassword 设置 Redis 密码
func WithRedisPassword(password string) RedisOption {
	return func(o *RedisOptions) {
		o.password = password
	}
}

// WithRedisDB 设置 Redis 数据库索引
func WithRedisDB(db int) RedisOption {
	return func(o *RedisOptions) {
		o.db = db
	}
}

// WithRedisPoolSize 设置连接池大小
func WithRedisPoolSize(size int) RedisOption {
	return func(o *RedisOptions) {
		o.poolSize = size
	}
}

// WithRedisMinIdleConns 设置最小空闲连接数
func WithRedisMinIdleConns(n int) RedisOption {
	return func(o *RedisOptions) {
		o.minIdleConns = n
	}
}

// WithRedisDialTimeout 设置连接超时
func WithRedisDialTimeout(d time.Duration) RedisOption {
	return func(o *RedisOptions) {
		o.dialTimeout = d
	}
}

// WithRedisReadTimeout 设置读超时
func WithRedisReadTimeout(d time.Duration) RedisOption {
	return func(o *RedisOptions) {
		o.readTimeout = d
	}
}

// WithRedisWriteTimeout 设置写超时
func WithRedisWriteTimeout(d time.Duration) RedisOption {
	return func(o *RedisOptions) {
		o.writeTimeout = d
	}
}

// defaultRedisOptions 返回默认 Redis 配置
func defaultRedisOptions() *RedisOptions {
	return &RedisOptions{
		addr:         "localhost:6379",
		password:     "",
		db:           0,
		poolSize:     10,
		minIdleConns: 5,
		dialTimeout:  5 * time.Second,
		readTimeout:  3 * time.Second,
		writeTimeout: 3 * time.Second,
	}
}

// RedisClient Redis 客户端
type RedisClient struct {
	Client *redis.Client
	logger core.Logger
}

// NewRedis 创建 Redis 连接
func NewRedis(log core.Logger, opts ...RedisOption) (*RedisClient, error) {
	// 应用默认配置
	options := defaultRedisOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(options)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         options.addr,
		Password:     options.password,
		DB:           options.db,
		PoolSize:     options.poolSize,
		MinIdleConns: options.minIdleConns,
		DialTimeout:  options.dialTimeout,
		ReadTimeout:  options.readTimeout,
		WriteTimeout: options.writeTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Infow("Redis connected", "addr", options.addr, "db", options.db)

	return &RedisClient{
		Client: client,
		logger: log.With("component", "redis"),
	}, nil
}

// Get 获取缓存 (支持 JSON 反序列化)
func (c *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	return nil
}

// GetString 获取字符串值
func (c *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Set 设置缓存 (支持 JSON 序列化)
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.Client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// SetString 设置字符串值
func (c *RedisClient) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// Delete 删除缓存
func (c *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if err := c.Client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}

// Exists 检查 key 是否存在
func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// Expire 设置 key 过期时间
func (c *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.Client.Expire(ctx, key, expiration).Err()
}

// TTL 获取 key 的剩余生存时间
func (c *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}

// Incr 递增计数器
func (c *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

// Decr 递减计数器
func (c *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return c.Client.Decr(ctx, key).Result()
}

// Health 检查 Redis 健康
func (c *RedisClient) Health(ctx context.Context) error {
	if err := c.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// Close 关闭连接
func (c *RedisClient) Close() error {
	c.logger.Info("Closing Redis connection")
	return c.Client.Close()
}

// Shutdown 实现 ShutdownHandler 接口
func (c *RedisClient) Shutdown(ctx context.Context) error {
	return c.Close()
}
