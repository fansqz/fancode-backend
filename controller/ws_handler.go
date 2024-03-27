package controller

import (
	"FanCode/constants"
	"FanCode/models/dto"
	"FanCode/service"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsController interface {
	// HandlerWebSocket 启动ws链接
	HandlerWebSocket(ctx *gin.Context)
}

type wsController struct {
	debugService service.DebugService
	wsService    service.WsService
}

func NewWsController(ds service.DebugService, ws service.WsService) WsController {
	return &wsController{
		debugService: ds,
		wsService:    ws,
	}
}

func (w *wsController) HandlerWebSocket(ctx *gin.Context) {
	w.handleWebSocket(ctx, ctx.Writer, ctx.Request)
}

func (ws *wsController) handleWebSocket(ctx *gin.Context, w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	user := ctx.Keys["user"].(*dto.UserInfo)
	ws.wsService.AddWsConn(int32(user.ID), conn)
	defer ws.wsService.RemoveConn(int32(user.ID))
	// 处理WebSocket连接
	for {
		// 读取消息
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
			// 解析json，并根据type交给不同的handler处理
			var wsReq dto.WsRequest
			if err = json.Unmarshal(p, &wsReq); err != nil {
				log.Println(err)
				continue
			}
			switch wsReq.Type {
			case constants.DebugWs:
				ws.debugService.HandleMessage(ctx, wsReq.Data)
			}
		} else if messageType == websocket.PingMessage {
			if err = conn.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
				log.Println(err)
			}
		} else if messageType == websocket.PongMessage {

		} else if messageType == websocket.CloseMessage {
			break
		}

		// 发送消息
		err = conn.WriteMessage(messageType, []byte("Hello, world!"))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
