package global

import (
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var (
	// Mode project mode: development/staging/production
	// RuntimeRoot runtime root path prefix
	Conf  *AppConfig
	Mysql *gorm.DB
	Redis *redis.Client
)
