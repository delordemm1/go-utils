package cache

import (
	"context"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	rCache *cache.Cache
}
type rediser interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd
	SetXX(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.BoolCmd
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.BoolCmd

	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

func NewCacheService(redis rediser, tinyLFUSize int, tinyLFUTime time.Duration) *CacheService {
	return &CacheService{
		rCache: cache.New(&cache.Options{
			Redis:      redis,
			LocalCache: cache.NewTinyLFU(tinyLFUSize, tinyLFUTime),
		}),
	}
}

func (s *CacheService) Get(ctx context.Context, key string, value interface{}) error {
	return s.rCache.Get(ctx, key, value)
}

func (s *CacheService) Set(ctx context.Context, key string, ttl time.Duration, value interface{}) error {
	return s.rCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
		TTL:   ttl,
	})
}

func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.rCache.Delete(ctx, key)
}

func (s *CacheService) Exists(ctx context.Context, key string) bool {
	return s.rCache.Exists(ctx, key)
}

func (s *CacheService) Remember(ctx context.Context, key string, ttl time.Duration, val interface{}, do func(*cache.Item) (interface{}, error)) error {
	if !s.Exists(ctx, key) {
		err := s.rCache.Set(&cache.Item{
			Ctx:   ctx,
			Key:   key,
			Value: val,
			TTL:   ttl,
			Do:    do,
		})
		if err != nil {
			return err
		}
	}
	return s.Get(ctx, key, val)
}
