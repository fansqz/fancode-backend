package service

import (
	"FanCode/config"
	"FanCode/constants"
	e "FanCode/error"
	"FanCode/utils"
	"os"
	"path"
)

/**
 * 放一些公用的方法
 */

const (
	/* 一道题目的结构如下：
	// problemFile:
	//	c     //保存c代码
	//	java // 保存java代码
	//	go    // 保存go代码
	//	io    //保存用例
	*/
	CCodePath    = "c"
	JavaCodePath = "java"
	GoCodePath   = "go"
	CaseFilePath = "io"
)

const (
	CMainFile    = "main.c"
	JavaMainFile = "Main.java"
	GoMainFile   = "main.go"
)

const (
	AcmCCodeFilePath    = "./resources/acmTemplate/c"
	AcmGoCodeFilePath   = "./resources/acmTemplate/go"
	AcmJavaCodeFilePath = "./resources/acmTemplate/java"
)

// 根据题目的路径获取题目中编程语言的路径
func getCodePathByProblemPath(problemPath string, language string) (string, *e.Error) {
	switch language {
	case constants.ProgramC:
		return path.Join(problemPath, CCodePath), nil
	case constants.ProgramJava:
		return path.Join(problemPath, JavaCodePath), nil
	case constants.ProgramGo:
		return path.Join(problemPath, GoCodePath), nil
	default:
		return "", e.ErrLanguageNotSupported
	}
}

// 根据编程语言获取该编程语言的Main文件名称
func getMainFileNameByLanguage(language string) (string, *e.Error) {
	switch language {
	case constants.ProgramC:
		return CMainFile, nil
	case constants.ProgramJava:
		return JavaMainFile, nil
	case constants.ProgramGo:
		return GoMainFile, nil
	default:
		return "", e.ErrLanguageNotSupported
	}
}

// 根据题目的路径获取题目中用例的路径
func getCasePathByLocalProblemPath(localProblemPath string) string {
	return path.Join(localProblemPath, CaseFilePath)
}

// 根据题目的相对路径，获取题目的本地路径
func getLocalProblemPath(config *config.AppConfig, p string) string {
	return path.Join(config.FilePathConfig.ProblemFileDir, p)
}

// 给用户的此次运行生成一个临时目录
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
