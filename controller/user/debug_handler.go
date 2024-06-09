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
	// StepIn 单步调试，会进入函数内部
	StepIn(ctx *gin.Context)
	// StepOut 单步调试，会跳出当前程序
	StepOut(ctx *gin.Context)
	// StepOver 单步调试，跳过不进入程序内部
	StepOver(ctx *gin.Context)
	// Continue 到达下一个断点
	Continue(ctx *gin.Context)
	// AddBreakpoints 添加断点
	AddBreakpoints(ctx *gin.Context)
	// RemoveBreakpoints 移除断点
	RemoveBreakpoints(ctx *gin.Context)
	// GetStackTrace 获取当前栈信息
	GetStackTrace(ctx *gin.Context)
	// GetFrameVariables 根据栈帧id获取变量列表
	GetFrameVariables(ctx *gin.Context)
	// GetVariables 根据引用获取变量信息，如果是指针，获取指针指向的内容，如果是结构体，获取结构体内容
	GetVariables(ctx *gin.Context)
	// CloseDebugSession 关闭调试session
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
	var req dto.RemoveBreakpointRequest
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

func (d *debugController) GetStackTrace(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	stackFrames, err := d.debugService.GetStackTrace(key)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(stackFrames)
}

func (d *debugController) GetFrameVariables(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	frameId := ctx.PostForm("frameId")
	variables, err := d.debugService.GetFrameVariables(key, frameId)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(variables)
}

// GetVariables 根据引用获取变量信息，如果是指针，获取指针指向的内容，如果是结构体，获取结构体内容
func (d *debugController) GetVariables(ctx *gin.Context) {
	result := r.NewResult(ctx)
	key := ctx.PostForm("key")
	reference := ctx.PostForm("reference")
	variables, err := d.debugService.GetVariables(key, reference)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(variables)
}
