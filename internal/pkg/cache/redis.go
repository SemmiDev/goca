package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sammidev/goca/internal/config"
)

type RedisClient struct {
	*redis.Client
}

var _ Cache = (*RedisClient)(nil)

func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisDSN(),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		MinIdleConns: cfg.RedisMinIdleConns,
		PoolSize:     cfg.RedisPoolSize,
		PoolTimeout:  cfg.RedisPoolTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RedisPingTimeout)
	defer cancel()

	if err := db.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		Client: db,
	}, nil
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (interface{}, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
