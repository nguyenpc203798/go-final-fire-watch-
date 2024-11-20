// dbs/redis.go
package dbs

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var ctx = context.Background()

// Khởi tạo kết nối Redis
func InitializeRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     GetEnv("REDIS_ADDR"), // Địa chỉ Redis server
		Password: GetEnv("REDIS_PASS"), // Không có mật khẩu
		DB:       0,                    // DB mặc định
	})

	// Kiểm tra kết nối
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Không thể kết nối tới Redis: ", err)
	}

	log.Println("Kết nối thành công tới Redis")
}

func DeleteCacheByKeyword(ctx context.Context, keyword string) error {
	// Tìm tất cả các key phù hợp với từ khóa
	pattern := fmt.Sprintf("*%s*", keyword) // Tìm các key chứa từ khóa
	keys, err := RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("Error fetching keys: %v", err)
	}

	if len(keys) == 0 {
		log.Printf("No keys found matching pattern: %s", pattern)
		return nil
	}

	// Xóa tất cả các key phù hợp
	_, err = RedisClient.Del(ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("Error deleting keys: %v", err)
	}

	log.Printf("Deleted %d keys matching pattern: %s", len(keys), pattern)
	return nil
}
