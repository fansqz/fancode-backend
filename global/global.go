package global

import (
	"FanCode/global/config"
	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
)

var (
	// Mode project mode: development/staging/production
	// RuntimeRoot runtime root path prefix
	Conf  *config.AppConfig
	Mysql *gorm.DB
	Redis redis.Conn
)
