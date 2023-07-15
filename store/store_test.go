package store

import (
	"FanCode/setting"
	"os"
	"strings"
	"testing"
)

func TestCOS_SaveFile(t *testing.T) {
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.ReplaceAll(path, "store", "conf/config.ini")
	setting.Init(path)
	path, _ = os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	f := strings.NewReader("文件内容")
	store := NewCOS()
	store.SaveFile("/question1/file2.text", f)
}
