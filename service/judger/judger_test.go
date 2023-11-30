package judger

import (
	"FanCode/constants"
	"gotest.tools/v3/assert"
	"math/rand"
	"os"
	"strconv"
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
	_, err := judgeCore.Compile(compileFiles, "./test_file/test_execute", &CompileOptions{
		Language: language,
		Timeout:  1000 * time.Second,
	})
	if err != nil {
		assert.NilError(t, err)
		return
	}
	defer os.Remove("./test_file/test_execute")

	// 运行
	executeOption := &ExecuteOptions{
		Language:    language,
		LimitTime:   int64(10 * time.Second),
		MemoryLimit: 100 * 1024 * 1024,
		CPUQuota:    100000,
	}

	err = judgeCore.Execute("./test_file/test_execute", input, output, exitCh, executeOption)
	if err != nil {
		assert.NilError(t, err)
		return
	}

	// 输入数据
	defer func() {
		exitCh <- "exit"
	}()
	for i := 0; i < 10; i++ {
		a := rand.Int() % 100
		b := rand.Int() % 100
		input <- []byte(strconv.Itoa(a) + " " + strconv.Itoa(b))
		select {
		case result := <-output:
			assert.Equal(t, true, result.Executed)
			assert.Equal(t, strconv.Itoa(a+b)+"\n", string(result.Output))
		}
	}
}

func TestJudgeCore_Timeout(t *testing.T) {
	judgeCore := NewJudgeCore()
	// 程序进行编译
	input := make(chan []byte)
	output := make(chan ExecuteResult)
	exitCh := make(chan string)

	// 编译
	_, err := judgeCore.Compile([]string{"./test_file/test_timeout.c"}, "./test_file/test_timeout",
		&CompileOptions{Timeout: 2 * time.Second})
	assert.NilError(t, err)
	defer func() {
		// 删除文件
		err = os.Remove("./test_file/test_timeout")
		assert.NilError(t, err)
		exitCh <- "exit"
	}()

	executeOption := &ExecuteOptions{
		Language:    constants.ProgramC,
		LimitTime:   int64(1 * time.Second),
		MemoryLimit: 1 * 1024 * 1024, //限制1m
		CPUQuota:    10000,           //限制cpu
	}
	err = judgeCore.Execute("./test_file/test_timeout", input, output, exitCh, executeOption)
	assert.NilError(t, err)
	input <- []byte("1 2")
	select {
	case result := <-output:
		assert.Equal(t, false, result.Executed)
		assert.Equal(t, "运行超时\n", result.ErrorMessage)
	}
}

func TestJudgeCore_MemoryOut(t *testing.T) {
	judgeCore := NewJudgeCore()
	// 程序进行编译
	input := make(chan []byte)
	output := make(chan ExecuteResult)
	exitCh := make(chan string)

	// 编译
	_, err := judgeCore.Compile([]string{"./test_file/test_memory_limit.c"}, "./test_file/test_memory_limit",
		&CompileOptions{Timeout: 2 * time.Second})
	assert.NilError(t, err)
	defer func() {
		// 删除文件
		err = os.Remove("./test_file/test_memory_limit")
		assert.NilError(t, err)
		exitCh <- "exit"
	}()

	executeOption := &ExecuteOptions{
		Language:    constants.ProgramC,
		LimitTime:   int64(1 * time.Second),
		MemoryLimit: 1 * 1024 * 1024, //限制1m
		CPUQuota:    10000,           //限制cpu
	}
	err = judgeCore.Execute("./test_file/test_memory_limit", input, output, exitCh, executeOption)
	assert.NilError(t, err)
	input <- []byte("1 2")
	select {
	case result := <-output:
		assert.Equal(t, false, result.Executed)
		assert.Equal(t, "内存超出限制\n", result.ErrorMessage)
	}
}

func TestJudgeCore_Compile(t *testing.T) {
	judgeCore := NewJudgeCore()
	// 编译
	compileResult, err := judgeCore.Compile([]string{"./test_file/test_compile_err.c"}, "./test_file/test_compile_err",
		&CompileOptions{
			Timeout:       2 * time.Second,
			ExcludedPaths: []string{"./test_file"},
		})
	assert.NilError(t, err)
	assert.Equal(t, false, compileResult.Compiled)
	return
}
