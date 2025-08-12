package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ekyc-backend/pkg/config"
	"github.com/ekyc-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	logger *logger.Logger
}

func NewRedis(cfg *config.Config, logger *logger.Logger) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis connection established")

	return &Redis{
		client: client,
		logger: logger,
	}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *Redis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *Redis) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Idempotency helpers
func (r *Redis) CheckIdempotency(ctx context.Context, idempotencyKey string) (bool, error) {
	key := fmt.Sprintf("idempotency:%s", idempotencyKey)
	exists, err := r.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *Redis) SetIdempotency(ctx context.Context, idempotencyKey string, result interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("idempotency:%s", idempotencyKey)
	return r.Set(ctx, key, result, expiration)
}

func (r *Redis) GetIdempotencyResult(ctx context.Context, idempotencyKey string) (string, error) {
	key := fmt.Sprintf("idempotency:%s", idempotencyKey)
	return r.Get(ctx, key)
}

// Rate limiting helpers
func (r *Redis) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	current, err := r.Incr(ctx, key)
	if err != nil {
		return false, err
	}

	if current == 1 {
		r.Expire(ctx, key, window)
	}

	return current <= int64(limit), nil
}

// Session management
func (r *Redis) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Set(ctx, key, data, expiration)
}

func (r *Redis) GetSession(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Get(ctx, key)
}

func (r *Redis) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Del(ctx, key)
}
