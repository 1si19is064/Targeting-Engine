package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"Targeting-Engine/internal/models"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Test connection
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// SetCampaigns caches campaigns
func (rc *RedisCache) SetCampaigns(campaigns map[string]*models.Campaign) error {
	data, err := json.Marshal(campaigns)
	if err != nil {
		return fmt.Errorf("failed to marshal campaigns: %w", err)
	}

	return rc.client.Set(rc.ctx, "campaigns", data, 5*time.Minute).Err()
}

// GetCampaigns retrieves cached campaigns
func (rc *RedisCache) GetCampaigns() (map[string]*models.Campaign, error) {
	data, err := rc.client.Get(rc.ctx, "campaigns").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Key not found
		}
		return nil, fmt.Errorf("failed to get campaigns from cache: %w", err)
	}

	var campaigns map[string]*models.Campaign
	if err := json.Unmarshal([]byte(data), &campaigns); err != nil {
		return nil, fmt.Errorf("failed to unmarshal campaigns: %w", err)
	}

	return campaigns, nil
}

// SetTargetingRules caches targeting rules
func (rc *RedisCache) SetTargetingRules(rules map[string][]*models.TargetingRule) error {
	data, err := json.Marshal(rules)
	if err != nil {
		return fmt.Errorf("failed to marshal targeting rules: %w", err)
	}

	return rc.client.Set(rc.ctx, "targeting_rules", data, 5*time.Minute).Err()
}

// GetTargetingRules retrieves cached targeting rules
func (rc *RedisCache) GetTargetingRules() (map[string][]*models.TargetingRule, error) {
	data, err := rc.client.Get(rc.ctx, "targeting_rules").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Key not found
		}
		return nil, fmt.Errorf("failed to get targeting rules from cache: %w", err)
	}

	var rules map[string][]*models.TargetingRule
	if err := json.Unmarshal([]byte(data), &rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal targeting rules: %w", err)
	}

	return rules, nil
}

// InvalidateCache removes all cached data
func (rc *RedisCache) InvalidateCache() error {
	return rc.client.FlushDB(rc.ctx).Err()
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}
