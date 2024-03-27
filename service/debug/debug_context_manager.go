package debug

import (
	"FanCode/constants"
	de "FanCode/service/debug/debugger"
	"log"
)

/**
 * 管理调试上下文的
 */
var debugContextMap = make(map[string]*DebugContext, 20)

// StartDebugging 启动一次调试任务，用key来标识
func StartDebugging(key string, language constants.LanguageType, compileFiles []string, workPath string) error {
	_, ok := debugContextMap[key]
	if ok {
		DestroyDebugContext(key)
	}
	var debugger de.Debugger
	switch language {
	case constants.LanguageC:
		debugger = de.NewGdbDebugger()
	}
	chn, err := debugger.Launch(compileFiles, workPath)
	if err != nil {
		return err
	}
	debugContextMap[key] = &DebugContext{
		Language:     language,
		Debugger:     debugger,
		DebuggerChan: chn,
	}
	return nil
}

// GetDebugContext 根据key获取一个debugger
func GetDebugContext(key string) (*DebugContext, bool) {
	debugContext, ok := debugContextMap[key]
	return debugContext, ok
}

// DestroyDebugContext 销毁一个debugger
func DestroyDebugContext(key string) {
	debugContext, ok := debugContextMap[key]
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
	delete(debugContextMap, key)
}
