package service

import (
	"FanCode/config"
	"FanCode/constants"
	"FanCode/utils"
	"os"
	"path"
)

/**
 * 放一些公用的方法
 */

const (
	AcmCCodeFilePath    = "./resources/acmTemplate/c"
	AcmGoCodeFilePath   = "./resources/acmTemplate/go"
	AcmJavaCodeFilePath = "./resources/acmTemplate/java"
)

// getExecutePath 给用户的此次运行生成一个临时目录
func getExecutePath(config *config.AppConfig) string {
	uuid := utils.GetUUID()
	executePath := path.Join(config.FilePathConfig.TempDir, uuid)
	return executePath
}

// getTempDir 获取一个随机的临时文件夹
func getTempDir(config *config.AppConfig) string {
	uuid := utils.GetUUID()
	executePath := config.FilePathConfig.TempDir + "/" + uuid
	return executePath
}

func getAcmCodeTemplate(language string) (string, error) {
	var filePath string
	switch language {
	case constants.ProgramC:
		filePath = AcmCCodeFilePath
	case constants.ProgramGo:
		filePath = AcmGoCodeFilePath
	case constants.ProgramJava:
		filePath = AcmJavaCodeFilePath
	}
	code, err := os.ReadFile(filePath)
	return string(code), err
}
