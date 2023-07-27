// Package setting
// @Author: fzw
// @Create: 2023/7/14
// @Description: 初始化时读取配置文件相关工具
package initialize

import (
	"FanCode/global"
	"FanCode/global/config"
	"gopkg.in/ini.v1"
	"strings"
)

// Init
//
//	@Description: 初始化配置
//	@param file 配置文件路径
//	@return error
func InitSetting(file string) error {
	cfg, err := ini.Load(file)
	if err != nil {
		return err
	}
	global.Conf = new(config.AppConfig)
	cfg.MapTo(global.Conf)
	//遍历releasePath
	startPaths := strings.Split(global.Conf.ReleaseStartPath, ",")
	releasePathConfig := &config.ReleasePathConfig{StartWith: startPaths}
	global.Conf.ReleasePathConfig = releasePathConfig

	global.Conf.MySqlConfig = config.NewMySqlConfig(cfg)
	global.Conf.COSConfig = config.NewCOSConfig(cfg)
	global.Conf.FilePathConfig = config.NewFilePathConfig(cfg)
	global.Conf.CasbinConfig = config.NewCasbinConfig(cfg)
	return nil
}
