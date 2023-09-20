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
	CheckChunkSet(path string) ([]string, *e.Error)
	// CancelUpload 取消上传
	CancelUpload(path string) *e.Error
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

func (f *fileService) CheckChunkSet(path string) ([]string, *e.Error) {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return nil, e.ErrServer
	}
	answer := make([]string, len(dirs))
	for i, a := range dirs {
		answer[i] = a.Name()
	}
	return answer, nil
}

// CancelUpload 取消上传
func CancelUpload(path string) *e.Error {
	err := os.RemoveAll(path)
	if err != nil {
		return e.ErrServer
	}
	return nil
}
