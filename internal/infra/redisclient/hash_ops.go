package redisclient

// HSet - lưu field vào hash
func (r *RedisClient) HSet(key string, field string, value interface{}) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

// HGet - lấy field từ hash
func (r *RedisClient) HGet(key string, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll - lấy toàn bộ hash
func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}
