package config

import "gopkg.in/ini.v1"

type RedisConfig struct {
	host     string `ini:"host"`
	port     string `ini:"port"`
	password string `ini:"password"`
}

func NewRedisConfig(cfg *ini.File) *RedisConfig {
	redisConfig := &RedisConfig{}
	cfg.Section("redis").MapTo(redisConfig)
	return redisConfig
}
