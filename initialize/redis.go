package initialize

import (
	"FanCode/global"
	"github.com/go-redis/redis"
)

var (
	RedisClient *redis.Client
)

// InitRedis
//
//	@Description: todo:初始化redis，留以后做吧
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     global.Conf.Host + ":" + global.Conf.Port,
		Password: global.Conf.Password,
		DB:       0, // 数据库
	})
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
