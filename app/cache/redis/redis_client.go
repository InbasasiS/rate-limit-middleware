package redis

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/rate-limit/config"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache() *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     config.GetConfig().Cache.Host + ":" + config.GetConfig().Cache.Port,
		Password: "",
		DB:       0,
	})
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(key string) (string, error) {
	val, err := r.client.Get(key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *RedisCache) Set(key, value string, ex time.Duration) error {
	return r.client.Set(key, value, ex).Err()
}

func (r *RedisCache) Incr(key string) (int64, error) {
	val, err := r.client.Incr(key).Result()
	if err != nil {
		return -1, err
	}
	return val, nil
}

func (r *RedisCache) Del(key string) error {
	err := r.client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisCache) Expire(key string, ex time.Duration) error {
	err := r.client.Expire(key, ex).Err()
	if err != nil {
		return err
	}
	return nil
}
