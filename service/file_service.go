package service

import (
	e "FanCode/error"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"os"
)

// FileService 文件上传相关service
type FileService interface {
	// StartUpload 启动上传命令
	StartUpload() (string, *e.Error)
	// Upload 上传分片
	Upload(path string, ctx *gin.Context, file *multipart.FileHeader) *e.Error
	// CheckChunkSet 检测分片的文件名称集合
	CheckChunkSet(ctx *gin.Context)
	// CancelUpload 取消上传
	CancelUpload(ctx *gin.Context)
	// CompleteUpload 完成大文件上传功能
	CompleteUpload(ctx *gin.Context)
}

type fileService struct {
}

func (f *fileService) StartUpload() (string, *e.Error) {
	tempPath := getTempDir()
	err := os.MkdirAll(tempPath, 0755)
	if err != nil {
		return "", e.ErrServer
	}
	return tempPath, nil
}

func (f *fileService) Upload(path string, ctx *gin.Context, file *multipart.FileHeader) *e.Error {
	err := ctx.SaveUploadedFile(file, path)
	if err != nil {
		return e.ErrServer
	}
	return nil
}
