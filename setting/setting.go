// Package setting
// @Author: fzw
// @Create: 2023/7/14
// @Description: 初始化时读取配置文件相关工具
package setting

import (
	"gopkg.in/ini.v1"
	"strings"
)

var Conf = new(AppConfig)

// AppConfig
// @Description:应用配置
type AppConfig struct {
	Release          bool   `ini:"release"` //是否是上线模式
	Port             string `ini:"port"`    //端口
	ReleaseStartPath string `ini:"releaseStartPath"`
	ProUrl           string `ini:"proUrl"`
	*MySqlConfig
	*ReleasePathConfig
	*StoreConfig
}

type ReleasePathConfig struct {
	StartWith []string
}

// Init
//
//	@Description: 初始化配置
//	@param file 配置文件路径
//	@return error
func Init(file string) error {
	cfg, err := ini.Load(file)
	if err != nil {
		return err
	}

	storeConfig := &StoreConfig{}
	mysqlConfig := &MySqlConfig{}

	cfg.MapTo(Conf)
	cfg.Section("mysql").MapTo(mysqlConfig)
	cfg.Section("store").MapTo(storeConfig)
	//遍历releasePath
	startPaths := strings.Split(Conf.ReleaseStartPath, ",")
	releasePathConfig := &ReleasePathConfig{StartWith: startPaths}

	Conf.MySqlConfig = mysqlConfig
	Conf.ReleasePathConfig = releasePathConfig
	Conf.StoreConfig = storeConfig
	return nil
}
