package debug

import (
	"FanCode/constants"
	"fmt"
	"testing"
)

func TestDebugHandler(t *testing.T) {
	testCode()
}

func testCode() {
	debugHandler := NewDebugHandler(constants.ProgramC)
	// 编译文件
	r, _ := debugHandler.Compile([]string{"./test_file/test_debug_c.c"}, "./test_file/test_debug_c", nil)
	var debugResult *DebugResult
	var err error

	// 启动调试
	debugResult, err = debugHandler.StartDebug(r.CompiledFilePath, "./test_file", &DebugOptions{
		Breakpoints: []Breakpoint{
			{"test_debug_c.c", 7},
			{"test_debug_c.c", 11},
		},
	})

	fmt.Println(debugResult)
	fmt.Println(err)
}
