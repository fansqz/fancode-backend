package debug

import (
	"FanCode/constants"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDebugHandler(t *testing.T) {
	testCode(t)
}

func testCode(t *testing.T) {
	debugHandler := NewDebugHandler(constants.ProgramC)
	// 编译文件
	r, _ := debugHandler.Compile([]string{"./test_file/test_debug_c.c"}, "./test_file/test_debug_c", nil)
	var debugResult *DebugResult
	var err error

	// 启动调试
	debugResult, err = debugHandler.Start(r.CompiledFilePath, &StartDebugOptions{
		WorkPath: "./test_file",
		Breakpoints: []Breakpoint{
			{"test_debug_c.c", 7},
			{"test_debug_c.c", 11},
		},
	})

	assert.Equal(t, 7, debugResult.Line)

	fmt.Println(debugResult)

	debugResult, _ = debugHandler.Next(1)
	fmt.Println(err)
}
