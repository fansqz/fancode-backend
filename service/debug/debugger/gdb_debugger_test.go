package debugger

import (
	"FanCode/constants"
	e "FanCode/error"
	"FanCode/utils"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"path"
	"testing"
)

var cha = make(chan interface{}, 10)
var userCha = make(chan []byte, 10)

func TestGdbDebugger(t *testing.T) {
	// 创建工作目录, 用户的临时文件
	executePath := getExecutePath("/var/fanCode/tempDir")
	defer os.RemoveAll(executePath)
	if err := os.MkdirAll(executePath, os.ModePerm); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		return
	}
	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err2 *e.Error
	code := `#include <stdio.h>           //1
int main() {                              //2
    int a, b;                             //3
    a = 1;                                //4
	printf("aaa");                        //5
	scanf("%d", &a);                      //6
    printf("a * a = %d\n", a * a);        //7
	b = 3;                                //8
	scanf("%d", &a);                      //9
    printf("a + a = %d\n", a + a);        //10
    return 0;                             //11
}
`
	if compileFiles, err2 = saveUserCode(constants.LanguageC,
		code, executePath); err2 != nil {
		log.Println(err2)
		return
	}
	debug := NewGdbDebugger(debugNotificationCallback)
	err := debug.Launch(compileFiles, executePath, constants.LanguageC)
	assert.Nil(t, err)
	// 接受调试编译成功信息
	data := <-cha
	assert.Equal(t, &CompileEvent{
		Success: true,
		Message: "用户代码编译成功",
	}, data)
	data = <-cha
	assert.Equal(t, &LaunchEvent{
		Success: true,
		Message: "目标代码加载成功",
	}, data)

	// 添加断点
	err = debug.AddBreakpoints([]*Breakpoint{{"/main.c", 4}, {"/main.c", 8}, {"/main.c", 11}})
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &BreakpointEvent{
		Reason:      constants.NewType,
		Breakpoints: []*Breakpoint{{"/main.c", 4}},
	}, data)
	data = <-cha
	assert.Equal(t, &BreakpointEvent{
		Reason:      constants.NewType,
		Breakpoints: []*Breakpoint{{"/main.c", 8}},
	}, data)
	data = <-cha
	assert.Equal(t, &BreakpointEvent{
		Reason:      constants.NewType,
		Breakpoints: []*Breakpoint{{"/main.c", 11}},
	}, data)

	// 启动用户程序
	err = debug.Start()
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &ContinuedEvent{}, data)

	// 程序到达第一个断点
	data = <-cha
	assert.Equal(t, &StoppedEvent{
		Reason: constants.BreakpointStopped,
		File:   "/main.c",
		Line:   4,
	}, data)

	// continue
	err = debug.Continue()
	assert.Nil(t, err)
	j := 0
	for i := 0; i < 2; i++ {
		data = <-cha
		switch data.(type) {
		case *ContinuedEvent:
			j++
		case *OutputEvent:
			j += 2
			assert.Equal(t, &OutputEvent{"aaa"}, data)
		}
	}
	assert.Equal(t, j, 3)

	// 输入用户输入
	err = debug.SendToConsole("9\n")
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &OutputEvent{
		Output: "a * a = 81\n",
	}, data)
	data = <-cha
	assert.Equal(t, &StoppedEvent{
		Reason: constants.BreakpointStopped,
		File:   "/main.c",
		Line:   8,
	}, data)

	// 测试step
	err = debug.StepOver()
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &ContinuedEvent{}, data)
	data = <-cha
	assert.Equal(t, &StoppedEvent{
		Reason: constants.StepStopped,
		File:   "/main.c",
		Line:   9,
	}, data)

	//  测试stepIn是否会进入系统依赖的函数内部
	err = debug.StepIn()
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &ContinuedEvent{}, data)
	assert.Equal(t, len(cha), 0)

	// 输入scanf
	err = debug.SendToConsole("10\n")
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &StoppedEvent{
		Reason: constants.BreakpointStopped,
		File:   "/main.c",
		Line:   10,
	}, data)

	// 测试结束
	err = debug.Continue()
	assert.Nil(t, err)
	j = 0
	for i := 0; i < 2; i++ {
		data = <-cha
		switch data.(type) {
		case *ContinuedEvent:
			j++
		case *OutputEvent:
			j += 2
			assert.Equal(t, &OutputEvent{"a + a = 20\n"}, data)
		}
	}
	data = <-cha
	assert.Equal(t, &ExitedEvent{
		ExitCode: 0,
	}, data)
}

func debugNotificationCallback(data interface{}) {
	cha <- data
}

// getExecutePath 给用户的此次运行生成一个临时目录
func getExecutePath(tempDir string) string {
	uuid := utils.GetUUID()
	executePath := path.Join(tempDir, uuid)
	return executePath
}

func saveUserCode(language constants.LanguageType, codeStr string, executePath string) ([]string, *e.Error) {
	var compileFiles []string
	var mainFile string
	var err2 *e.Error

	if mainFile, err2 = getMainFileNameByLanguage(language); err2 != nil {
		log.Println(err2)
		return nil, err2
	}
	if err := os.WriteFile(path.Join(executePath, mainFile), []byte(codeStr), 0644); err != nil {
		log.Println(err)
		return nil, e.ErrServer
	}
	// 将main文件进行编译即可
	compileFiles = []string{path.Join(executePath, mainFile)}

	return compileFiles, nil
}

func getMainFileNameByLanguage(language constants.LanguageType) (string, *e.Error) {
	switch language {
	case constants.LanguageC:
		return "main.c", nil
	case constants.LanguageJava:
		return "Main.java", nil
	case constants.LanguageGo:
		return "main.go", nil
	default:
		return "", e.ErrLanguageNotSupported
	}
}
