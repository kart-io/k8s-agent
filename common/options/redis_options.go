package options

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/options/validation"
	"github.com/kart-io/logger/core"
)

// RedisOptions Redis配置
type RedisOptions struct {
	Addr         string        `mapstructure:"addr" yaml:"addr" json:"addr"`
	Password     string        `mapstructure:"password" yaml:"password" json:"password"`
	DB           int           `mapstructure:"db" yaml:"db" json:"db"`
	PoolSize     int           `mapstructure:"pool_size" yaml:"pool_size" json:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns" yaml:"min_idle_conns" json:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout" json:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" json:"write_timeout"`
}

// NewRedisOptions 创建默认的Redis配置
func NewRedisOptions() *RedisOptions {
	return &RedisOptions{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// Validate 验证配置
func (o *RedisOptions) Validate() error {
	// 使用通用验证器
	if err := validation.ValidateAddr(o.Addr, "redis"); err != nil {
		return err
	}

	if err := validation.ValidateRedisDB(o.DB); err != nil {
		return err
	}

	if err := validation.ValidatePositiveInt(o.PoolSize, "redis pool_size"); err != nil {
		return err
	}

	if err := validation.ValidateNonNegativeInt(o.MinIdleConns, "redis min_idle_conns"); err != nil {
		return err
	}

	if err := validation.ValidateTimeouts(o.DialTimeout, o.ReadTimeout, o.WriteTimeout, "redis"); err != nil {
		return err
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *RedisOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Addr, "redis.addr", o.Addr, "Redis server address")
	fs.StringVar(&o.Password, "redis.password", o.Password, "Redis password")
	fs.IntVar(&o.DB, "redis.db", o.DB, "Redis database index")
	fs.IntVar(&o.PoolSize, "redis.pool-size", o.PoolSize, "Redis connection pool size")
	fs.IntVar(&o.MinIdleConns, "redis.min-idle-conns", o.MinIdleConns, "Redis minimum idle connections")
	fs.DurationVar(&o.DialTimeout, "redis.dial-timeout", o.DialTimeout, "Redis dial timeout")
	fs.DurationVar(&o.ReadTimeout, "redis.read-timeout", o.ReadTimeout, "Redis read timeout")
	fs.DurationVar(&o.WriteTimeout, "redis.write-timeout", o.WriteTimeout, "Redis write timeout")
}

// ApplyTo 将配置应用到目标接口
// 接受一个函数切片指针，将配置转换为 db.RedisOption 函数选项
func (o *RedisOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// 类型断言，检查是否为函数选项切片指针
	switch v := target.(type) {
	case *[]interface{}:
		// 将配置转换为通用选项
		*v = append(*v,
			map[string]interface{}{
				"addr":         o.Addr,
				"password":     o.Password,
				"db":           o.DB,
				"poolSize":     o.PoolSize,
				"minIdleConns": o.MinIdleConns,
				"dialTimeout":  o.DialTimeout,
				"readTimeout":  o.ReadTimeout,
				"writeTimeout": o.WriteTimeout,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
// 设置默认值和计算派生值
func (o *RedisOptions) Complete() error {
	// 如果地址为空，设置默认值
	if o.Addr == "" {
		o.Addr = "localhost:6379"
	}

	// 确保数据库索引在有效范围内
	if o.DB < 0 || o.DB > 15 {
		o.DB = 0
	}

	// 确保连接池配置合理
	if o.PoolSize <= 0 {
		o.PoolSize = 10
	}

	if o.MinIdleConns < 0 {
		o.MinIdleConns = 5
	}

	// 确保 MinIdleConns 不大于 PoolSize
	if o.MinIdleConns > o.PoolSize {
		o.MinIdleConns = o.PoolSize
	}

	// 确保超时时间有合理的值
	if o.DialTimeout <= 0 {
		o.DialTimeout = 5 * time.Second
	}

	if o.ReadTimeout <= 0 {
		o.ReadTimeout = 3 * time.Second
	}

	if o.WriteTimeout <= 0 {
		o.WriteTimeout = 3 * time.Second
	}

	return nil
}

// WithRedisAddr 设置Redis地址
func WithRedisAddr(addr string) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.Addr = addr
	}
}

// WithRedisPassword 设置Redis密码
func WithRedisPassword(password string) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.Password = password
	}
}

// WithRedisDB 设置Redis数据库索引
func WithRedisDB(db int) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.DB = db
	}
}

// WithRedisPoolSize 设置连接池大小
func WithRedisPoolSize(size int) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.PoolSize = size
	}
}

// WithRedisMinIdleConns 设置最小空闲连接数
func WithRedisMinIdleConns(min int) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.MinIdleConns = min
	}
}

// WithRedisDialTimeout 设置连接超时时间
func WithRedisDialTimeout(timeout time.Duration) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.DialTimeout = timeout
	}
}

// WithRedisReadTimeout 设置读超时时间
func WithRedisReadTimeout(timeout time.Duration) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.ReadTimeout = timeout
	}
}

// WithRedisWriteTimeout 设置写超时时间
func WithRedisWriteTimeout(timeout time.Duration) func(*RedisOptions) {
	return func(o *RedisOptions) {
		o.WriteTimeout = timeout
	}
}

// 连接redis
func (o *RedisOptions) ConnectRedis(log core.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         o.Addr,
		Password:     o.Password,
		DB:           o.DB,
		PoolSize:     o.PoolSize,
		MinIdleConns: o.MinIdleConns,
		DialTimeout:  o.DialTimeout,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,
	})
	return client, nil
}

// Health 检查 Redis 健康状态
func (o *RedisOptions) Health(ctx context.Context, log core.Logger) error {
	client, err := o.ConnectRedis(log)
	if err != nil {
		return errors.New("failed to connect to Redis: " + err.Error())
	}
	defer client.Close()
	if err := client.Ping(ctx).Err(); err != nil {
		return errors.New("failed to ping Redis: " + err.Error())
	}
	return nil
}
