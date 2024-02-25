package service

import (
	"FanCode/config"
	"FanCode/constants"
	e "FanCode/error"
	"FanCode/models/dto"
	"FanCode/service/debug"
	"FanCode/service/debug/define"
	"FanCode/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

// DebugService
// 用户调试相关
type DebugService interface {
	HandleMessage(*gin.Context, interface{})
}

type debugService struct {
	config       *config.AppConfig
	judgeService *judgeService
	wsService    *wsService
}

func NewDebugService(cf *config.AppConfig, js *judgeService, ws *wsService) DebugService {
	return &debugService{
		config:       cf,
		judgeService: js,
		wsService:    ws,
	}
}

func (d *debugService) HandleMessage(ctx *gin.Context, data interface{}) {
	var result map[string]interface{}

	// 解析 JSON 到 map
	data2, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}
	if err := json.Unmarshal(data2, &result); err != nil {
		log.Println(err)
		return
	}
	option := result["option"].(constants.DebugOptionType)
	switch option {
	case constants.StartDebug:
		var startReq dto.StartDebugRequest
		if err = json.Unmarshal(data2, &startReq); err != nil {
			log.Println(err)
			return
		}
		d.Start(ctx, startReq)
	}
}

func (d *debugService) Start(ctx *gin.Context, startReq dto.StartDebugRequest) {
	user := ctx.Keys["user"].(*dto.UserInfo)
	wsConn, _ := d.wsService.GetConn(int32(user.ID))
	// 创建工作目录, 用户的临时文件
	executePath := getExecutePath(d.config)
	if err := os.MkdirAll(executePath, os.ModePerm); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		resp := dto.NewFailDebugResponseByRequest(&startReq.DebugRequestBase, "系统错误")
		wsConn.WriteJSON(resp)
		return
	}
	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err2 *e.Error
	if compileFiles, err2 = d.judgeService.saveUserCode(startReq.Language,
		startReq.Code, executePath); err2 != nil {
		// Add logging for error
		log.Printf("SaveUserCode error: %v\n", err2)
		resp := dto.NewFailDebugResponseByRequest(&startReq.DebugRequestBase, "系统错误")
		wsConn.WriteJSON(resp)
		return
	}

	// 启动debugging
	key := utils.GetUUID()
	if err := debug.StartDebugging(key, startReq.Language, compileFiles, executePath); err != nil {
		// Add logging for error
		log.Printf("StartDebugging error: %v\n", err)
		resp := dto.NewFailDebugResponseByRequest(&startReq.DebugRequestBase, "系统错误")
		wsConn.WriteJSON(resp)
		return
	}

	// 读取输入数据的管道并创建协程处理数据
	channel, _ := debug.GetDebuggerRespChan(key)
	go d.ListenAndHandleDebugEvents(ctx, channel)

	debugger, _ := debug.GetDebugger(key)
	if err := debugger.Start(); err != nil {
		log.Printf("StartDebugging error: %v\n", err)
		resp := dto.NewFailDebugResponseByRequest(&startReq.DebugRequestBase, "调试任务启动失败")
		wsConn.WriteJSON(resp)
		return
	}
	resp := dto.NewSuccessDebugResponseByRequest(&startReq.DebugRequestBase, "调试任务启动成功")
	wsConn.WriteJSON(resp)
}

// ListenAndHandleDebugEvents 循环监控调试事件，并生成event响应给用户
func (d *debugService) ListenAndHandleDebugEvents(ctx *gin.Context, channel chan interface{}) {
	for {
		data := <-channel
		if _, ok := data.(define.BreakpointEvent); ok {

		}
		if _, ok := data.(define.OutputEvent); ok {

		}
		if _, ok := data.(define.StoppedEvent); ok {

		}
		if _, ok := data.(define.ContinuedEvent); ok {

		}
		if _, ok := data.(define.ExitedEvent); ok {

		}
	}
}

func (d *debugService) SendToConsole(ctx *gin.Context) {

}

func (d *debugService) Next(ctx *gin.Context) {

}

func (d *debugService) Step(ctx *gin.Context) {

}

func (d *debugService) Continue(ctx *gin.Context) {

}

func (d *debugService) AddBreakpoints(ctx *gin.Context) {

}

func (d *debugService) RemoveBreakpoints(ctx *gin.Context) {

}

func (d *debugService) Terminate(ctx *gin.Context) {

}
