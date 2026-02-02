package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Cache provides caching functionality using Redis
type Cache struct {
	client *redis.Client
}

// NewCache creates a new cache instance
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// Key prefixes for different cache types
const (
	DashboardCachePrefix = "dashboard:"
	AnalyticsCachePrefix = "analytics:"
)

// Default TTLs
const (
	DashboardTTL = 5 * time.Minute
	AnalyticsTTL = 10 * time.Minute
)

// DashboardCacheKey generates a cache key for dashboard data
func DashboardCacheKey(orgID uuid.UUID, startDate, endDate time.Time) string {
	return fmt.Sprintf("%s%s:%s:%s",
		DashboardCachePrefix,
		orgID.String(),
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)
}

// Get retrieves a value from cache and unmarshals it into the target
func (c *Cache) Get(ctx context.Context, key string, target interface{}) error {
	if c.client == nil {
		return redis.Nil // Return not found if no client
	}

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// Set stores a value in cache with the specified TTL
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.client == nil {
		return nil // No-op if no client
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a key from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return nil
	}

	return c.client.Del(ctx, key).Err()
}

// DeleteByPattern removes all keys matching a pattern
func (c *Cache) DeleteByPattern(ctx context.Context, pattern string) error {
	if c.client == nil {
		return nil
	}

	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// InvalidateDashboardCache invalidates all dashboard cache for an organization
func (c *Cache) InvalidateDashboardCache(ctx context.Context, orgID uuid.UUID) error {
	pattern := fmt.Sprintf("%s%s:*", DashboardCachePrefix, orgID.String())
	return c.DeleteByPattern(ctx, pattern)
}

// InvalidateAllDashboardCache invalidates all dashboard cache
func (c *Cache) InvalidateAllDashboardCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", DashboardCachePrefix)
	return c.DeleteByPattern(ctx, pattern)
}

// IsNotFound checks if the error is a cache miss
func IsNotFound(err error) bool {
	return err == redis.Nil
}
