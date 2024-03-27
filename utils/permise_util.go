package utils

import (
	"os"
	"path/filepath"
)

func CheckFilePermissions() bool {
	// 写入临时文件以检查权限
	tempDir := os.TempDir()
	testFile := filepath.Join(tempDir, "test_perm.txt")

	// 尝试创建文件
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()

	// 尝试写入数据
	err = os.WriteFile(testFile, []byte("test"), 0666)
	if err != nil {
		return false
	}

	// 尝试执行文件（在Windows上可能不相关）
	err = os.Chmod(testFile, 0777)
	if err != nil {
		return false
	}

	// 尝试删除文件
	err = os.Remove(testFile)
	if err != nil {
		return false
	}

	return true
}
