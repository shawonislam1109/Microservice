package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"isp-management-system/internal/models"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	userCacheTTL = 5 * time.Minute
)

// Cache defines the interface for cache operations.
type Cache interface {
	GetUser(ctx context.Context, username string) (*models.User, error)
	SetUser(ctx context.Context, user *models.User) error
	Close() error
}

// redisCache implements the Cache interface for Redis.
type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache client.
func NewRedisCache(addr, password string, db int) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &redisCache{client: client}, nil
}

// userCacheKey generates a standardized key for storing a user in Redis.
func userCacheKey(username string) string {
	return fmt.Sprintf("user:%s", username)
}

// GetUser retrieves a user from the Redis cache.
func (c *redisCache) GetUser(ctx context.Context, username string) (*models.User, error) {
	key := userCacheKey(username)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("user not found in cache")
		}
		return nil, fmt.Errorf("failed to get user from redis: %w", err)
	}

	var user models.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}

// SetUser stores a user in the Redis cache with a TTL.
func (c *redisCache) SetUser(ctx context.Context, user *models.User) error {
	key := userCacheKey(user.Username)
	val, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	if err := c.client.Set(ctx, key, val, userCacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set user in redis: %w", err)
	}

	return nil
}

// Close closes the Redis client connection.
func (c *redisCache) Close() error {
	return c.client.Close()
}