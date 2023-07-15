package file_store

import (
	"strings"
)

type Store interface {
	// 上传文件到对象存储
	SaveFile(file *strings.Reader) error
	// 下载一个文件
	//storePaht:对象存储的路劲，loadPath:本地路径
	LoadFile(storePath, loadPath string) error
	// 上传文件夹到对象存储
	UploadFolder(storePath, loadPath string) error
	// 下载文件夹
	LoadFolder(storePath, loadPath string) error
	// 删除文件夹
	DeleteFolder(storePath string) error
}
