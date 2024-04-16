package user

import (
	"FanCode/constants"
	e "FanCode/error"
	"FanCode/models/dto"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
)

type DebugController interface {
	CreateDebugSession(ctx *gin.Context)
	// Start 启动调试
	Start(ctx *gin.Context)
	// CreateSseConnect 会创建一个sse链接，用于接受服务器响应
	CreateSseConnect(ctx *gin.Context)
	// SendToConsole 提交
	SendToConsole(ctx *gin.Context)
	// Step
	StepIn(ctx *gin.Context)
	// Step
	StepOut(ctx *gin.Context)
	// Step
	StepOver(ctx *gin.Context)
	// Continue
	Continue(ctx *gin.Context)
	// AddBreakpoints
	AddBreakpoints(ctx *gin.Context)
	// RemoveBreakpoints
	RemoveBreakpoints(ctx *gin.Context)
	// CloseDebugSession
	CloseDebugSession(ctx *gin.Context)
}

type debugController struct {
	debugService service.DebugService
}

func NewDebugController(ds service.DebugService) DebugController {
	return &debugController{
		debugService: ds,
	}
}

func (d *debugController) CreateDebugSession(ctx *gin.Context) {
	result := r.NewResult(ctx)
	language := ctx.PostForm("language")
	languageType := constants.LanguageType(language)
	key, err := d.debugService.CreateDebugSession(ctx, languageType)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(key)
}

// Start 开始调试
func (d *debugController) Start(ctx *gin.Context) {
	result := r.NewResult(ctx)
	var startReq dto.StartDebugRequest
	if err := ctx.BindJSON(&startReq); err != nil {
		return
	}
	err := d.debugService.Start(ctx, startReq)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("启动成功")
}

// CreateSseConnect
func (d *debugController) CreateSseConnect(ctx *gin.Context) {
	key := ctx.Param("key")
	d.debugService.CreateSseConnect(ctx, key)
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

// Step
func (d *debugController) StepIn(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	if err := d.debugService.StepIn(key); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("请求成功")
}

// Step
func (d *debugController) StepOut(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	if err := d.debugService.StepOut(key); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("请求成功")
}

// Step
func (d *debugController) StepOver(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	if err := d.debugService.StepOver(key); err != nil {
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
	if err := ctx.BindJSON(&req); err != nil {
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
	if err := ctx.BindJSON(&req); err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	if err := d.debugService.RemoveBreakpoints(req.Key, req.Breakpoints); err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	result.SuccessMessage("请求成功")
}

// CloseDebugSession
func (d *debugController) CloseDebugSession(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	if err := d.debugService.CloseDebugSession(key); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("请求成功")
}
