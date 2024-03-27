package user

import (
	e "FanCode/error"
	"FanCode/models/dto"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
)

type DebugController interface {
	// Start 启动调试，会创建一个sse链接
	Start(ctx *gin.Context)
	// SendToConsole 提交
	SendToConsole(ctx *gin.Context)
	// Next
	Next(ctx *gin.Context)
	// Step
	Step(ctx *gin.Context)
	// Continue
	Continue(ctx *gin.Context)
	// AddBreakpoints
	AddBreakpoints(ctx *gin.Context)
	// RemoveBreakpoints
	RemoveBreakpoints(ctx *gin.Context)
	// Terminate
	Terminate(ctx *gin.Context)
}

type debugController struct {
	debugService service.DebugService
}

func NewDebugController(ds service.DebugService) DebugController {
	return &debugController{
		debugService: ds,
	}
}

// Start 启动调试，会创建一个sse链接
func (d *debugController) Start(ctx *gin.Context) {
	result := r.NewResult(ctx)
	var startReq dto.StartDebugRequest
	if err := ctx.BindJSON(startReq); err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	d.debugService.Start(ctx, startReq)
}

// SendToConsole 提交
func (d *debugController) SendToConsole(ctx *gin.Context) {
	result := r.NewResult(ctx)
	input := ctx.PostForm("input")
	key := ctx.PostForm("key")
	if err := d.debugService.SendToConsole(key, input); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("请求成功")
}

// Next
func (d *debugController) Next(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	if err := d.debugService.Next(key); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("请求成功")
}

// Step
func (d *debugController) Step(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	if err := d.debugService.Step(key); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("请求成功")
}

// Continue
func (d *debugController) Continue(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	if err := d.debugService.Continue(key); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("请求成功")
}

// AddBreakpoints
func (d *debugController) AddBreakpoints(ctx *gin.Context) {
	result := r.NewResult(ctx)
	var req dto.AddBreakpointRequest
	if err := ctx.BindJSON(req); err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	if err := d.debugService.AddBreakpoints(req.Key, req.Breakpoints); err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	result.SuccessMessage("请求成功")
}

// RemoveBreakpoints
func (d *debugController) RemoveBreakpoints(ctx *gin.Context) {
	result := r.NewResult(ctx)
	var req dto.AddBreakpointRequest
	if err := ctx.BindJSON(req); err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	if err := d.debugService.RemoveBreakpoints(req.Key, req.Breakpoints); err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	result.SuccessMessage("请求成功")
}

// Terminate
func (d *debugController) Terminate(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	if err := d.debugService.Terminate(key); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("请求成功")
}
