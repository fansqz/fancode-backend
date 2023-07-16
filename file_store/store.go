package file_store

import (
	"strings"
)

type Store interface {
	// 上传文件到对象存储
	SaveFile(storePath string, file *strings.Reader) error
	// 下载一个文件
	//storePaht:对象存储的路劲，loadPath:本地路径
	DownloadFile(storePath, localPath string) error
	// 上传文件夹到对象存储
	UploadFolder(storePath string, localPath string)
	// 下载文件夹
	DownloadFolder(storePath, localPath string) error
	// 删除文件夹
	DeleteFolder(storePath string) error
}
