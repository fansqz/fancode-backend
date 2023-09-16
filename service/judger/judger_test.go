package judger

import (
	"FanCode/constants"
	"gotest.tools/v3/assert"
	"log"
	"testing"
	"time"
)

// 需要提前创建镜像
func Test_Judger(t *testing.T) {
	judgeCore := NewJudgeCore()
	// 程序进行编译
	input := make(chan []byte)
	output := make(chan []byte)
	errOutput := make(chan error)
	exitCh := make(chan string)

	// 编译
	err := judgeCore.Compile(constants.ProgramC, []string{"./test_file/main.c"}, "./test_file/test", 2*time.Second)
	if err != nil {
		log.Println(err)
		return
	}
	executeOption := &ExecuteOption{
		Language:    constants.ProgramC,
		OutputCh:    output,
		InputCh:     input,
		ErrOutPutCh: errOutput,
		ExitCh:      exitCh,
		ExecFile:    "./test_file/test",
		LimitTime:   10 * time.Second,
		LimitMemory: 10 * 1024,
	}
	err = judgeCore.Execute(executeOption)
	if err != nil {
		log.Println(err)
		return
	}
	input <- []byte("1 2")
	select {
	case out := <-output:
		assert.Equal(t, "3\n", string(out))
	case err2 := <-errOutput:
		assert.Equal(t, nil, err2)
		return
	}
	exitCh <- "exit"
}
