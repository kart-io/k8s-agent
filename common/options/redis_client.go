package options

import (
	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/logger/core"
)

// NewRedisClient 根据配置创建 Redis 客户端
// 这个方法直接在 Options 中实现 Redis 连接创建，简化调用链
//
// 之前的调用方式：
//
//	client, err := db.NewRedisFromOptions(logger, opts.Redis)
//
// 现在的调用方式：
//
//	client, err := opts.Redis.NewRedisClient(logger)
//
// 使用示例：
//
//	opts := cmd/service/app/options.NewServerOptions()
//	client, err := opts.Redis.NewRedisClient(logger)
//	if err != nil {
//	    return err
//	}
func (o *RedisOptions) NewRedisClient(log core.Logger) (*db.RedisClient, error) {
	return db.NewRedis(log,
		db.WithAddr(o.Addr),
		db.WithRedisPassword(o.Password),
		db.WithRedisDB(o.DB),
		db.WithRedisPoolSize(o.PoolSize),
		db.WithRedisMinIdleConns(o.MinIdleConns),
		db.WithRedisDialTimeout(o.DialTimeout),
		db.WithRedisReadTimeout(o.ReadTimeout),
		db.WithRedisWriteTimeout(o.WriteTimeout),
	)
}
