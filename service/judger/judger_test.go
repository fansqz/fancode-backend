package judger

import (
	"FanCode/constants"
	"log"
	"testing"
)

// 需要提前创建镜像
func Test_Judger(t *testing.T) {
	judgeCore, err := NewJudgeCore(10)
	defer judgeCore.Release()
	if err != nil {
		log.Println(err)
		return
	}
	// 程序进行编译

	input := make(chan string)
	output, erroutput, err3 := judgeCore.RunCode("./test", input, constants.ProgramC)
	if err3 != nil {
		log.Println(err3)
		return
	}
	input <- "1 2"
	select {
	case out := <-output:
		println(out)
	case err4 := <-erroutput:
		log.Println(err4)
		return
	}
}
