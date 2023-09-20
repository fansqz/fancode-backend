package service

import (
	"FanCode/global"
	"FanCode/utils"
)

/**
 * 放一些公用的方法
 */

// getTempDir 获取一个随机的临时文件夹
func getTempDir() string {
	uuid := utils.GetUUID()
	executePath := global.Conf.FilePathConfig.TempDir + "/" + uuid
	return executePath
}
