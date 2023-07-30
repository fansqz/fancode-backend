package global

import (
	"FanCode/global/config"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
)

var (
	// Mode project mode: development/staging/production
	// RuntimeRoot runtime root path prefix
	Conf  *config.AppConfig
	Mysql *gorm.DB
	Redis redis.Conn
)
