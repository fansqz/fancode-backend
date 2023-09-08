// Package db
// @Author: fzw
// @Create: 2023/7/4
// @Description: 数据库开启关闭等
package global

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitMysql
//
//	@Description: 初始化mysql
//	@param cfg
//	@return error
func InitMysql(cfg *MySqlConfig) error {
	dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	var err error
	Mysql, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return err
}
