// Package result
// @Author: fzw
// @Create: 2023/7/14
// @Description: 用做统一结果返回的
package result

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// result
// @Description: 统一result，并返回数据给前端
type result struct {
	ctx *gin.Context
}

func NewResult(ctx *gin.Context) *result {
	return &result{ctx: ctx}
}

// Success1
//
//	@Description: 返回成功结果
//	@receiver r
//	@param data
func (r *result) SuccessData(data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	res := Success()
	res.Data = data
	r.ctx.JSON(http.StatusOK, res)
}

func (r *result) SuccessMessage(message string) {
	res := &ResultCont{
		Code:    OK.GetCode(),
		Message: message,
	}
	r.ctx.JSON(http.StatusOK, res)
}

func (r *result) Success(message string, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	res := &ResultCont{
		Code:    OK.GetCode(),
		Message: message,
		Data:    data,
	}
	r.ctx.JSON(http.StatusOK, res)
}

// Error1
//
//	@Description: 返回错误的消息
//	@receiver r
func (r *result) ErrorMessage(code int, message string) {
	res := &ResultCont{
		Code:    code,
		Message: message,
	}
	r.ctx.JSON(http.StatusOK, res)
}

func (r *result) Error(code int, message string, data interface{}) {
	res := &ResultCont{
		Code:    code,
		Message: message,
		Data:    data,
	}
	r.ctx.JSON(http.StatusOK, res)
}

func (r *result) SimpleErrorMessage(message string) {
	res := &ResultCont{
		Code:    CUSTOM_SIMPLE_ERROR_MESSAGE.code,
		Message: message,
	}
	r.ctx.JSON(http.StatusOK, res)
}
