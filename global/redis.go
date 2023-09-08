package global

import (
	"github.com/go-redis/redis"
)

// InitRedis
func InitRedis(cfg *RedisConfig) error {
	Redis = redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       0, // 数据库
	})
	return nil
}

// Close
//
//	@Description: 关闭redis
func Close() {
	err := Redis.Close()
	if err != nil {
		return
	}
}
