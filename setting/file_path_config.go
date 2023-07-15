package setting

import "gopkg.in/ini.v1"

type FilePathConfig struct {
	QuestionFileDir string `ini:"questionFileDir"` //题目文件目录
	TempDir         string `ini:"tmpDir"`          //临时目录
}

func NewFilePathConfig(cfg *ini.File) *FilePathConfig {
	filePathConfig := &FilePathConfig{}
	cfg.Section("filePath").MapTo(filePathConfig)
	return filePathConfig
}
