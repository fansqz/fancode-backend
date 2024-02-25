package debug

import (
	"FanCode/constants"
	"FanCode/service/debug/define"
	"log"
)

/*
*
debugger_util
提供的是一个调试工具
使用StartDebugging启动一次调试任务，指定workPath作为调试任务的工作目录以及调试任务的唯一标识key
结束调试任务时使用DestroyDebugging结束。
*/
var debuggerMap = make(map[string]define.Debugger, 20)
var debuggerChanMap = make(map[string]chan interface{}, 20)

// StartDebugging 启动一次调试任务，用key来标识
func StartDebugging(key string, language constants.LanguageType, compileFiles []string, workPath string) error {
	_, ok := debuggerMap[key]
	if ok {
		DestroyDebugging(key)
	}
	var debugger define.Debugger
	switch language {
	case constants.LanguageC:
		debugger = NewGdbDebugger()
	}
	chn, err := debugger.Launch(compileFiles, workPath)
	if err != nil {
		return err
	}
	debuggerMap[key] = debugger
	debuggerChanMap[key] = chn
	return nil
}

// GetDebugger 根据key获取一个debugger
func GetDebugger(key string) (define.Debugger, bool) {
	debugger, ok := debuggerMap[key]
	return debugger, ok
}

// GetDebuggerRespChan 根据key获取一个debugger的响应消息管道
func GetDebuggerRespChan(key string) (chan interface{}, bool) {
	chn, ok := debuggerChanMap[key]
	return chn, ok
}

// DestroyDebugging 销毁一个debugger
func DestroyDebugging(key string) {
	debugger, ok := debuggerMap[key]
	if !ok {
		return
	}
	for i := 0; i < 3; i++ {
		if err := debugger.Terminate(); err != nil {
			log.Println(err)
			continue
		} else {
			break
		}
	}
	delete(debuggerMap, key)
	delete(debuggerChanMap, key)
}
