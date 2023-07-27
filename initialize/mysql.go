// Package db
// @Author: fzw
// @Create: 2023/7/4
// @Description: 数据库开启关闭等
package initialize

import (
	"FanCode/global"
	"FanCode/global/config"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// InitMysql
//
//	@Description: 初始化mysql
//	@param cfg
//	@return error
func InitMysql(cfg *config.MySqlConfig) error {
	dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	var err error
	global.Mysql, err = gorm.Open("mysql", dsn)
	if err != nil {
		return err
	}
	//尝试ping通
	return global.Mysql.DB().Ping()
}

// CloseMysql
//
//	@Description: 关闭mysql
func CloseMysql() {
	err := global.Mysql.DB().Close()
	if err != nil {
		return
	}
}
