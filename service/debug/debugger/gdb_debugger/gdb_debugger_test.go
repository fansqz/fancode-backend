package gdb_debugger

import (
	"FanCode/constants"
	e "FanCode/error"
	"FanCode/service/debug/debugger"
	"FanCode/utils"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"path"
	"testing"
)

func TestGdbDebugger(t *testing.T) {
	var cha = make(chan interface{}, 10)
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
	debugNotificationCallback := func(data interface{}) {
		cha <- data
	}
	debug := NewGdbDebugger(debugNotificationCallback)
	err := debug.Launch(compileFiles, executePath, constants.LanguageC)
	assert.Nil(t, err)
	// 接受调试编译成功信息
	data := <-cha
	assert.Equal(t, &debugger.CompileEvent{
		Success: true,
		Message: "用户代码编译成功",
	}, data)
	data = <-cha
	assert.Equal(t, &debugger.LaunchEvent{
		Success: true,
		Message: "目标代码加载成功",
	}, data)

	// 添加断点
	err = debug.AddBreakpoints([]*debugger.Breakpoint{{"/main.c", 4}, {"/main.c", 8}, {"/main.c", 11}})
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &debugger.BreakpointEvent{
		Reason:      constants.NewType,
		Breakpoints: []*debugger.Breakpoint{{"/main.c", 4}},
	}, data)
	data = <-cha
	assert.Equal(t, &debugger.BreakpointEvent{
		Reason:      constants.NewType,
		Breakpoints: []*debugger.Breakpoint{{"/main.c", 8}},
	}, data)
	data = <-cha
	assert.Equal(t, &debugger.BreakpointEvent{
		Reason:      constants.NewType,
		Breakpoints: []*debugger.Breakpoint{{"/main.c", 11}},
	}, data)

	// 启动用户程序
	err = debug.Start()
	err = debug.SendToConsole("1 2\n")
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &debugger.ContinuedEvent{}, data)

	// 程序到达第一个断点
	data = <-cha
	assert.Equal(t, &debugger.StoppedEvent{
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
		case *debugger.ContinuedEvent:
			j++
		case *debugger.OutputEvent:
			j += 2
			assert.Equal(t, &debugger.OutputEvent{"aaa"}, data)
		}
	}
	assert.Equal(t, j, 3)

	data = <-cha
	assert.Equal(t, &debugger.StoppedEvent{
		Reason: constants.BreakpointStopped,
		File:   "/main.c",
		Line:   8,
	}, data)

	// 测试step
	err = debug.StepOver()
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &debugger.ContinuedEvent{}, data)
	data = <-cha
	assert.Equal(t, &debugger.StoppedEvent{
		Reason: constants.StepStopped,
		File:   "/main.c",
		Line:   9,
	}, data)

	//  测试stepIn是否会进入系统依赖的函数内部
	err = debug.StepIn()
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &debugger.ContinuedEvent{}, data)
	assert.Equal(t, len(cha), 0)

	// 输入scanf
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &debugger.StoppedEvent{
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
		case *debugger.ContinuedEvent:
			j++
		case *debugger.OutputEvent:
			j += 2
			assert.Equal(t, &debugger.OutputEvent{"a + a = 20\n"}, data)
		}
	}
	data = <-cha
	assert.Equal(t, &debugger.ExitedEvent{
		ExitCode: 0,
	}, data)
}

func TestGdbDebugger2(t *testing.T) {
	var cha = make(chan interface{}, 10)
	// A创建工作目录, 用户的临时文件
	executePath := getExecutePath("/var/fanCode/tempDir")
	defer os.RemoveAll(executePath)
	if err := os.MkdirAll(executePath, os.ModePerm); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		return
	}
	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err2 *e.Error
	code := `#include <stdio.h>        // 1
#include <stdlib.h>       // 2
// 定义枚举类型           // 3
typedef enum {            // 4
    RED,                  // 5
    GREEN,                // 6
    BLUE                  // 7
} Color;                  // 8
// 定义结构体类型         // 9
typedef struct {          // 10
    int id;               // 11
    float weight;         // 12
    Color color;          // 13
} Item;                   // 14
// 定义联合体类型         // 15
typedef union {           // 16
    int ival;             // 17
    float fval;           // 18
    char cval;            // 19
} Value;                  // 20
// 全局变量               // 21
int globalInt = 10;       // 22
float globalFloat = 3.14; // 23
char globalChar = 'A';    // 24
Item globalItem = {1, 65.5, RED}; // 25
// 静态全局变量           // 26
static int staticGlobalInt = 20; // 27
// 函数声明               // 28
void manipulateLocals();  // 29
void manipulatePointers();// 30
int main() {              // 31
    manipulateLocals(2);   // 32
    manipulatePointers(); // 33
    return 0;             // 34
}                         // 35
void manipulateLocals(int argint) { // 36
    // 局部变量           // 37
    int localInt = 5;     // 38
    char localChar = 'G'; // 39
    // 静态局部变量       // 40
    static float staticLocalFloat = 6.78; // 41
    // 结构体局部变量     // 42
    Item localItem;       // 43
    localItem.id = 2;     // 44
    localItem.weight = 42.0;   // 45
    localItem.color = GREEN;   // 46
    // 局部枚举变量       // 47
    Color localColor = BLUE;   // 48
    // 局部联合体变量     // 49
    Value localValue;     // 50
    localValue.ival = 123;// 51
    // 输出局部变量的值   // 52
    // printf("localInt: %d, localChar: %c, staticLocalFloat: %.2f\n", localInt, localChar, staticLocalFloat); // 53
    // printf("localItem: id=%d, weight=%.1f, color=%d\n", localItem.id, localItem.weight, localItem.color);     // 54
    // printf("localColor: %d, localValue: %d\n", localColor, localValue.ival);   // 55
}                         // 56
void manipulatePointers() { // 57
    // 动态分配的变量   // 58
    int* dynamicInt = (int*) malloc(sizeof(int)); // 59
    *dynamicInt = 30;           // 60
    // 指针变量         // 61
    int* ptrToInt = &globalInt; // 62
    Item* ptrToItem = &globalItem; // 63
    Color* ptrToColor = (Color*) malloc(sizeof(Color)); // 64
    *ptrToColor = BLUE;         // 65
    // 数组变量         // 66
    int intArray[3] = { 1, 2, 3 }; // 67
    float floatArray[] = { 1.1f, 2.2f, 3.3f }; // 68
    Color colorArray[3] = { RED, GREEN, BLUE }; // 69
    // 字符串           // 70
    char* string = "Hello, World!"; // 71
    // 清理动态分配的内存 // 72
    free(dynamicInt);     // 73
    free(ptrToColor);     // 74
}                         // 75
`
	if compileFiles, err2 = saveUserCode(constants.LanguageC,
		code, executePath); err2 != nil {
		log.Println(err2)
		return
	}
	debugNotificationCallback := func(data interface{}) {
		cha <- data
	}
	debug := NewGdbDebugger(debugNotificationCallback)
	err := debug.Launch(compileFiles, executePath, constants.LanguageC)
	assert.Nil(t, err)
	// 接受调试编译成功信息
	data := <-cha
	assert.Equal(t, &debugger.CompileEvent{
		Success: true,
		Message: "用户代码编译成功",
	}, data)
	data = <-cha
	assert.Equal(t, &debugger.LaunchEvent{
		Success: true,
		Message: "目标代码加载成功",
	}, data)

	// 添加断点
	err = debug.AddBreakpoints([]*debugger.Breakpoint{{"/main.c", 53}, {"/main.c", 73}})
	assert.Nil(t, err)
	data = <-cha
	assert.Equal(t, &debugger.BreakpointEvent{
		Reason:      constants.NewType,
		Breakpoints: []*debugger.Breakpoint{{"/main.c", 53}},
	}, data)
	data = <-cha
	assert.Equal(t, &debugger.BreakpointEvent{
		Reason:      constants.NewType,
		Breakpoints: []*debugger.Breakpoint{{"/main.c", 73}},
	}, data)

	// 启动用户程序
	err = debug.Start()
	<-cha
	<-cha
	// 测试第一个断点中的变量信息
	stacks, err := debug.GetStackTrace()
	assert.Nil(t, err)
	// 获取作用域
	variables, err := debug.GetFrameVariables(stacks[0].Id)
	assert.Nil(t, err)
	for _, variable := range variables {
		if variable.VariablesReference != "" {
			v, err := debug.GetVariables(variable.VariablesReference)
			fmt.Println(v)
			assert.Nil(t, err)
		}
	}

	// 测试第二个断点中的变量信息
	err = debug.Continue()
	data = <-cha
	data = <-cha
	stacks, err = debug.GetStackTrace()
	assert.Nil(t, err)
	// 获取作用域
	variables, err = debug.GetFrameVariables(stacks[0].Id)
	assert.Nil(t, err)
	// 测试打印三层
	for _, variable := range variables {
		fmt.Println(variable)
		if variable.VariablesReference != "" {
			vs, _ := debug.GetVariables(variable.VariablesReference)
			fmt.Println(vs)
			for _, v := range vs {
				if v.VariablesReference != "" {
					cd, _ := debug.GetVariables(v.VariablesReference)
					fmt.Println(cd)
				}
			}
		}
	}
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
