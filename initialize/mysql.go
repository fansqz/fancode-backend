// Package db
// @Author: fzw
// @Create: 2023/7/4
// @Description: 数据库开启关闭等
package initialize

import (
	"FanCode/initialize/setting"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	DB *gorm.DB
)

//
// InitMysql
//  @Description: 初始化mysql
//  @param cfg
//  @return error
//
func InitMysql(cfg *setting.MySqlConfig) error {
	dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	var err error
	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		return err
	}
	//尝试ping通
	return DB.DB().Ping()
}

//
// CloseMysql
//  @Description: 关闭mysql
//
func CloseMysql() {
	err := DB.Close()
	if err != nil {
		return
	}
}
