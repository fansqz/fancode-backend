package controllers

import "github.com/gin-gonic/gin"

// JudgeController
// @Description: 判题模块
type JudgeController interface {
	Execute(ctx *gin.Context)
	Submit(ctx *gin.Context)
}
