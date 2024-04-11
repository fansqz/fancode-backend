package debug

import (
	"FanCode/constants"
	de "FanCode/service/debug/debugger"
	"log"
)

type debugSessionManage struct {
	debugContextMap map[string]*DebugSession
}

var DebugSessionManage = &debugSessionManage{
	debugContextMap: make(map[string]*DebugSession, 20),
}

// CreateDebugSession 创建调试上下文
func (d *debugSessionManage) CreateDebugSession(key string, language constants.LanguageType) error {
	_, ok := d.debugContextMap[key]
	if ok {
		d.DestroyDebugSession(key)
	}
	channel := make(chan interface{}, 10)
	notificationCallback := func(data interface{}) {
		channel <- data
	}
	var debugger de.Debugger
	switch language {
	case constants.LanguageC:
		debugger = de.NewGdbDebugger(notificationCallback)
	}
	d.debugContextMap[key] = &DebugSession{
		StopProcessEventChan: make(chan struct{}),
		Language:             language,
		Debugger:             debugger,
		DebuggerChan:         channel,
	}
	return nil
}

// GetDebugSession 根据key获取一个debugger
func (d *debugSessionManage) GetDebugSession(key string) (*DebugSession, bool) {
	debugContext, ok := d.debugContextMap[key]
	return debugContext, ok
}

// DestroyDebugSession 销毁一个debugger
func (d *debugSessionManage) DestroyDebugSession(key string) {
	debugContext, ok := d.debugContextMap[key]
	if !ok {
		return
	}
	for i := 0; i < 3; i++ {
		if err := debugContext.Debugger.Terminate(); err != nil {
			log.Println(err)
			continue
		} else {
			break
		}
	}
	close(d.debugContextMap[key].DebuggerChan)
	delete(d.debugContextMap, key)
}
