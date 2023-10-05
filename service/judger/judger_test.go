package judger

import (
	"FanCode/constants"
	"gotest.tools/v3/assert"
	"os"
	"testing"
	"time"
)

func TestJudgeCore_Execute(t *testing.T) {
	execute(constants.ProgramC, t)
	execute(constants.ProgramJava, t)
}

func execute(language string, t *testing.T) {
	judgeCore := NewJudgeCore()
	// 程序进行编译
	input := make(chan []byte)
	output := make(chan ExecuteResult)
	exitCh := make(chan string)

	// 编译
	var compileFiles []string
	switch language {
	case constants.ProgramC:
		compileFiles = []string{"./test_file/test_execute.c"}
	case constants.ProgramJava:
		compileFiles = []string{"./test_file/test_execute.java"}
	}
	err := judgeCore.Compile(language, compileFiles,
		"./test_file/test_execute", 1000*time.Second)
	if err != nil {
		assert.NilError(t, err)
		return
	}
	defer os.Remove("./test_file/test_execute")

	// 运行
	executeOption := &ExecuteOption{
		Language:    language,
		OutputCh:    output,
		InputCh:     input,
		ExitCh:      exitCh,
		ExecFile:    "./test_file/test_execute.jar",
		LimitTime:   10 * time.Second,
		LimitMemory: 100 * 1024,
	}

	err = judgeCore.Execute(executeOption)
	if err != nil {
		assert.NilError(t, err)
		return
	}

	// 输入数据
	input <- []byte("1 2")
	defer func() {
		exitCh <- "exit"
	}()

	// 校验
	select {
	case result := <-output:
		assert.Equal(t, true, result.Executed)
		assert.Equal(t, "3\n", string(result.Output))
	}
}

func TestJudgeCore_Timeout(t *testing.T) {
	judgeCore := NewJudgeCore()
	// 程序进行编译
	input := make(chan []byte)
	output := make(chan ExecuteResult)
	exitCh := make(chan string)

	// 编译
	err := judgeCore.Compile(constants.ProgramC, []string{"./test_file/test_timeout.c"}, "./test_file/test_timeout", 2*time.Second)
	if err != nil {
		assert.NilError(t, err)
		return
	}
	defer func() {
		// 删除文件
		err = os.Remove("./test_file/test_timeout")
		assert.NilError(t, err)
		exitCh <- "exit"
	}()

	executeOption := &ExecuteOption{
		Language:    constants.ProgramC,
		OutputCh:    output,
		InputCh:     input,
		ExitCh:      exitCh,
		ExecFile:    "./test_file/test_timeout",
		LimitTime:   1 * time.Second,
		LimitMemory: 1 * 1024 * 1024, //限制1m
	}
	err = judgeCore.Execute(executeOption)
	if err != nil {
		assert.NilError(t, err)
		return
	}
	input <- []byte("1 2")
	select {
	case result := <-output:
		assert.Equal(t, false, result.Executed)
	}
}
