package service

import "github.com/gorilla/websocket"

// WsService 管理ws链接的
type WsService interface {
	AddWsConn(key int32, conn *websocket.Conn)
	RemoveConn(key int32)
}

// wsService
type wsService struct {
	// todo：支持并发操作！！！
	wsMap map[int32]*websocket.Conn
}

func NewWsService() WsService {
	return &wsService{
		wsMap: make(map[int32]*websocket.Conn, 10),
	}
}

func (ws *wsService) AddWsConn(key int32, conn *websocket.Conn) {
	conn, ok := ws.wsMap[key]
	if ok {
		conn.Close()
		ws.wsMap[key] = conn
	}
}

func (ws *wsService) RemoveConn(key int32) {
	conn, ok := ws.wsMap[key]
	if ok {
		conn.Close()
		delete(ws.wsMap, key)
	}
}

func (ws *wsService) GetConn(key int32) (*websocket.Conn, bool) {
	conn, ok := ws.wsMap[key]
	return conn, ok
}
