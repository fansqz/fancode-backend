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
	// DownloadFolder 下载文件夹, storePath：文件在对象存储中的位置，localPath：文件在本地中的位置
	DownloadFolder(storePath, localPath string) error
	// DownloadAndCompressFolder 下载并压缩文件夹，zipPath:压缩文件路劲，比如a/b/c.zip
	DownloadAndCompressFolder(storePath, localPath, zipPath string) error
	// DeleteFolder 删除文件夹
	DeleteFolder(storePath string) error
}
