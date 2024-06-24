package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Redis      *redis.Client
	Prefix     string
	MaxRetries int
}

func NewRedisRepository(url string, prefix string, maxRetries int) *RedisRepository {
	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(opt)

	return &RedisRepository{
		Redis:      rdb,
		Prefix:     prefix,
		MaxRetries: maxRetries,
	}
}

func (r *RedisRepository) FlushDB() error {
	ctx := context.Background()
	err := r.Redis.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("error in flusdb: %w", err)
	}
	err = r.Redis.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("error in flusdb: %w", err)
	}
	return nil
}

func (r *RedisRepository) Close() {
	r.Redis.Close()
}
