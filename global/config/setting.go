// Package initialize
// @Author: fzw
// @Create: 2023/7/14
// @Description: 初始化时读取配置文件相关工具
package config

import (
	"FanCode/global"
	"gopkg.in/ini.v1"
	"strings"
)

// InitSetting
//
//	@Description: 初始化配置
//	@param file 配置文件路径
//	@return error
func InitSetting(file string) error {
	cfg, err := ini.Load(file)
	if err != nil {
		return err
	}
	global.Conf = new(AppConfig)
	cfg.MapTo(global.Conf)
	//遍历releasePath
	startPaths := strings.Split(global.Conf.ReleaseStartPath, ",")
	releasePathConfig := &ReleasePathConfig{StartWith: startPaths}
	global.Conf.ReleasePathConfig = releasePathConfig

	global.Conf.MySqlConfig = NewMySqlConfig(cfg)
	global.Conf.RedisConfig = NewRedisConfig(cfg)
	global.Conf.EmailConfig = NewEmailConfig(cfg)
	global.Conf.COSConfig = NewCOSConfig(cfg)
	global.Conf.FilePathConfig = NewFilePathConfig(cfg)
	return nil
}
