// Package initialize
// @Author: fzw
// @Create: 2023/7/14
// @Description: 初始化时读取配置文件相关工具
package global

import (
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
	Conf = new(AppConfig)
	cfg.MapTo(Conf)
	//遍历releasePath
	startPaths := strings.Split(Conf.ReleaseStartPath, ",")
	releasePathConfig := &ReleasePathConfig{StartWith: startPaths}
	Conf.ReleasePathConfig = releasePathConfig

	Conf.MySqlConfig = NewMySqlConfig(cfg)
	Conf.RedisConfig = NewRedisConfig(cfg)
	Conf.EmailConfig = NewEmailConfig(cfg)
	Conf.COSConfig = NewCOSConfig(cfg)
	Conf.FilePathConfig = NewFilePathConfig(cfg)
	return nil
}
