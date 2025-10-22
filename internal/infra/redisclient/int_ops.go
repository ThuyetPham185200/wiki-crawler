package redisclient

import "time"

// SetInt - set integer value
func (r *RedisClient) SetInt(key string, value int64, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// GetInt - get integer value
func (r *RedisClient) GetInt(key string) (int64, error) {
	return r.client.Get(ctx, key).Int64()
}

// IncrBy - tăng key lên một giá trị
func (r *RedisClient) IncrBy(key string, increment int64) (int64, error) {
	return r.client.IncrBy(ctx, key, increment).Result()
}

// DecrBy - giảm key đi một giá trị
func (r *RedisClient) DecrBy(key string, decrement int64) (int64, error) {
	return r.client.DecrBy(ctx, key, decrement).Result()
}
