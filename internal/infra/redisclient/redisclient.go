package redisclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

var (
	instance *RedisClient
	once     sync.Once
	ctx      = context.Background()
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// InitSingleton - khởi tạo 1 lần duy nhất
func InitSingleton(cfg RedisConfig) *RedisClient {
	once.Do(func() {
		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password, // "" nếu không có password
			DB:       cfg.DB,
		})

		// Test kết nối
		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			panic(fmt.Sprintf("❌ Không kết nối được Redis: %v", err))
		}

		fmt.Println("✅ Redis connected:", cfg.Addr)

		instance = &RedisClient{
			client: rdb,
		}
	})
	return instance
}

// GetInstance - lấy instance Redis
func GetInstance() *RedisClient {
	if instance == nil {
		panic("⚠ Redis chưa được init! Gọi InitSingleton trước.")
	}
	return instance
}

// Close - đóng kết nối Redis
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetClient - lấy raw *redis.Client nếu cần
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// SetKey - set key với TTL
func (r *RedisClient) SetKey(key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// GetKey - lấy value
func (r *RedisClient) GetKey(key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// IncrKey - tăng giá trị integer
func (r *RedisClient) IncrKey(key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}
