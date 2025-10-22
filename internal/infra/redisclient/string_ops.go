package redisclient

import (
	"time"
)

// SetString - lưu string
func (r *RedisClient) SetString(key, value string, ttl time.Duration) error {
	return r.SetKey(key, value, ttl)
}

// GetString - lấy string
func (r *RedisClient) GetString(key string) (string, error) {
	return r.GetKey(key)
}
