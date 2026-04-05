package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type Cache interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, bool, error)
	SetBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error
	DeleteBalance(ctx context.Context, userID uuid.UUID) error
	SetBalanceBatch(ctx context.Context, balances map[uuid.UUID]decimal.Decimal) error
}

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(addr string, ttl time.Duration) (*RedisCache, error) {
	opts, err := redis.ParseURL(addr)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &RedisCache{client: client, ttl: ttl}, nil
}

func balanceKey(userID uuid.UUID) string {
	return "balance:" + userID.String()
}

func (c *RedisCache) GetBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, bool, error) {
	val, err := c.client.Get(ctx, balanceKey(userID)).Result()
	if err == redis.Nil {
		return decimal.Zero, false, nil
	}
	if err != nil {
		return decimal.Zero, false, fmt.Errorf("redis get: %w", err)
	}

	amount, err := decimal.NewFromString(val)
	if err != nil {
		return decimal.Zero, false, fmt.Errorf("parse decimal: %w", err)
	}
	return amount, true, nil
}

func (c *RedisCache) SetBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error {
	return c.client.Set(ctx, balanceKey(userID), amount.String(), c.ttl).Err()
}

func (c *RedisCache) DeleteBalance(ctx context.Context, userID uuid.UUID) error {
	return c.client.Del(ctx, balanceKey(userID)).Err()
}

func (c *RedisCache) SetBalanceBatch(ctx context.Context, balances map[uuid.UUID]decimal.Decimal) error {
	pipe := c.client.Pipeline()
	for userID, amount := range balances {
		pipe.Set(ctx, balanceKey(userID), amount.String(), c.ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
