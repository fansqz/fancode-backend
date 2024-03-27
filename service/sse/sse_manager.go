package sse

import (
	json2 "encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var connectMap = map[string]http.ResponseWriter{}

// CreateSssConnection 创建sse连接
func CreateSssConnection(key string, w http.ResponseWriter) {
	// 设置必要的HTTP头部，这些都是SSE所需要的
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	connectMap[key] = w
}

// SendData 向客户端发送数据
func SendData(key string, data interface{}) error {
	w, ok := connectMap[key]
	if !ok {
		return errors.New("连接不存在")
	}
	json, err := json2.Marshal(data)
	if err != nil {
		return err
	}
	// 写入事件数据
	fmt.Fprintf(w, "data: %s\n\n", string(json))
	// 刷新缓冲，确保立即发送到客户端
	flusher, _ := w.(http.Flusher)
	flusher.Flush()
	return nil
}

// Close 关闭连接
func Close(key string) error {
	w, ok := connectMap[key]
	if !ok {
		return errors.New("连接不存在")
	}
	// 发送一个特殊的消息，告诉客户端连接关闭
	fmt.Fprintf(w, "data: %s\n\n", "close")
	// 刷新缓冲，确保立即发送到客户端
	flusher, _ := w.(http.Flusher)
	flusher.Flush()
	delete(connectMap, key)
	return nil
}
