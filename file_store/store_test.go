package file_store

import (
	"FanCode/config"
	"os"
	"strings"
	"testing"
)

func InitConfig() *config.AppConfig {
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.ReplaceAll(path, "file_store", "conf/config.ini")
	conf, _ := config.InitSetting(path)
	return conf
}

func TestCOS_SaveFile(t *testing.T) {
	conf := InitConfig()
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	f := strings.NewReader("文件内容")
	store := NewProblemCOS(conf.COSConfig)
	store.SaveFile("/question1/file.text", f)
}

func TestCOS_LoadFile(t *testing.T) {
	conf := InitConfig()
	store := NewProblemCOS(conf.COSConfig)
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	store.DownloadFile("question1/file.text", path+"/file.text")
}

func TestCOS_LoadFolder(t *testing.T) {
	conf := InitConfig()
	InitConfig()
	store := NewProblemCOS(conf.COSConfig)
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	store.DownloadFolder("question1", path+"/question1")
}

func TestCOS_DeleteFolder(t *testing.T) {
	conf := InitConfig()
	store := NewProblemCOS(conf.COSConfig)
	store.DeleteFolder("question1")
}

func TestCOS_UploadFolder(t *testing.T) {
	conf := InitConfig()
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	store := NewProblemCOS(conf.COSConfig)
	store.UploadFolder("question", path+"/question")
}
