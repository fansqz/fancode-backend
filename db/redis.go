package db

import "github.com/gomodule/redigo/redis"

var (
	RedisClient redis.Conn
)

//
// InitRedis
//  @Description: todo:初始化redis，留以后做吧
//
func InitRedis() {
	var err error
	RedisClient, err = redis.Dial("tcp", "local")
	if err != nil {
		panic(err)
	}
}

//
// Close
//  @Description: 关闭redis
//
func Close() {
	err := RedisClient.Close()
	if err != nil {
		return
	}
}
