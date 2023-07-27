package file_store

import (
	"FanCode/initialize"
	"os"
	"strings"
	"testing"
)

func InitConfig() {
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.ReplaceAll(path, "file_store", "conf/config.ini")
	initialize.Init(path)
}

func TestCOS_SaveFile(t *testing.T) {
	InitConfig()
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	f := strings.NewReader("文件内容")
	store := NewCOS()
	store.SaveFile("/question1/file.text", f)
}

func TestCOS_LoadFile(t *testing.T) {
	InitConfig()
	store := NewCOS()
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	store.DownloadFile("question1/file.text", path+"/file.text")
}

func TestCOS_LoadFolder(t *testing.T) {
	InitConfig()
	store := NewCOS()
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	store.DownloadFolder("question1", path+"/question1")
}

func TestCOS_DeleteFolder(t *testing.T) {
	InitConfig()
	store := NewCOS()
	store.DeleteFolder("question1")
}

func TestCOS_UploadFolder(t *testing.T) {
	InitConfig()
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	store := NewCOS()
	store.UploadFolder("question", path+"/question")
}
