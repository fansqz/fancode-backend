package user

import (
	"github.com/gin-gonic/gin"
)

// DebugController
// 用户调试相关
type DebugController interface {
	// Launch 启动调试程序
	Launch(ctx *gin.Context)
	// Start 运行用户代码，正在开始调试
	Start(ctx *gin.Context) error
	// SendToConsole 输入数据到控制台
	SendToConsole(ctx *gin.Context)
	// Next 下一步，不会进入函数内部
	Next(ctx *gin.Context)
	// Step 下n一步，会进入函数内部
	Step(ctx *gin.Context)
	// Continue 忽略继续执行
	Continue(ctx *gin.Context)
	// AddBreakpoints 添加断点，同步
	AddBreakpoints(ctx *gin.Context)
	// RemoveBreakpoints 移除断点，同步
	RemoveBreakpoints(ctx *gin.Context)
	// Terminate 终止调试
	Terminate(ctx *gin.Context)
}

type debugController struct {
}

func (d *debugController) Launch(ctx *gin.Context) {

}

func (d *debugController) Start(ctx *gin.Context) {

}

func (d *debugController) SendToConsole(ctx *gin.Context) {

}

func (d *debugController) Next(ctx *gin.Context) {

}

func (d *debugController) Step(ctx *gin.Context) {

}

func (d *debugController) Continue(ctx *gin.Context) {

}

func (d *debugController) AddBreakpoints(ctx *gin.Context) {

}

func (d *debugController) RemoveBreakpoints(ctx *gin.Context) {

}

func (d *debugController) Terminate(ctx *gin.Context) {

}
