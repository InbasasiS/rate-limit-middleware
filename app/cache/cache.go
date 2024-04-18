package cache

import (
	"time"

	"github.com/rate-limit/app/cache/redis"
)

type Cache interface {
	Get(key string) (string, error)
	Set(key, value string, ex time.Duration) error
	Incr(key string) (int64, error)
	Expire(key string, ex time.Duration) error
}

func NewCacheClient() Cache {
	return redis.NewRedisCache()
}
