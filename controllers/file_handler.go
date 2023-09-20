package controllers

import (
	e "FanCode/error"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
)

type FileController interface {
	// StartUpload 启动上传命令
	StartUpload(ctx *gin.Context)
	// Upload 上传分片
	Upload(ctx *gin.Context)
	// CheckChunkSet 检测分片的文件名称集合
	CheckChunkSet(ctx *gin.Context)
	// CancelUpload 取消上传
	CancelUpload(ctx *gin.Context)
	// CompleteUpload 完成大文件上传功能
	CompleteUpload(ctx *gin.Context)
}

type fileController struct {
	fileService service.FileService
}

func (f *fileController) StartUpload(ctx *gin.Context) {
	result := r.NewResult(ctx)
	temp, err := f.fileService.StartUpload()
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(temp)
}

// Upload 上传分片
func (f *fileController) Upload(ctx *gin.Context) {
	result := r.NewResult(ctx)
	path := ctx.PostForm("path")
	fileHead, err2 := ctx.FormFile("chunk")
	if err2 != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err := f.fileService.Upload(path, ctx, fileHead)
	if err != nil {
		result.Error(err)
	}
	result.SuccessMessage("success upload")
}

func (f *fileController) CheckChunkSet(ctx *gin.Context) {
	result := r.NewResult(ctx)
	path := ctx.PostForm("path")
	set, err := f.fileService.CheckChunkSet(path)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(set)
}

func (f *fileController) CancelUpload(ctx *gin.Context) {
	result := r.NewResult(ctx)
	path := ctx.PostForm("path")
	err := f.fileService.CancelUpload(path)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("取消成功")
}
