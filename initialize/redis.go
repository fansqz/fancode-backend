package initialize

import (
	"FanCode/global/config"
	"github.com/go-redis/redis"
)

var (
	RedisClient *redis.Client
)

// InitRedis
func InitRedis(cfg *config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
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
	err := RedisClient.Close()
	if err != nil {
		return
	}
}
