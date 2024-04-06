package debugger

import (
	"FanCode/constants"
	e "FanCode/error"
	"FanCode/service/judger"
	"FanCode/utils"
	gdb2 "github.com/cyrus-and/gdb"
	"log"
	"os"
	"path"
	"testing"
)

func TestAccountService_GetAccountInfo(t *testing.T) {
	// 创建工作目录, 用户的临时文件
	executePath := getExecutePath("./")
	if err := os.MkdirAll(executePath, os.ModePerm); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		return
	}
	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err2 *e.Error
	code := `
#include <stdio.h>

int main() {
    int a, b;
    a = 1;
    b = 2;
    scanf("%d %d", &a, &b);
    // 64 位输出请用 printf("%lld") to 
    printf("%d\n", a + b);
    
    return 0;
}
`
	if compileFiles, err2 = saveUserCode(constants.LanguageC,
		code, executePath); err2 != nil {
		log.Println(err2)
		return
	}
	J := judger.NewJudgeCore()
	execFile := path.Join(executePath, "main")
	J.Compile(compileFiles, execFile, &judger.CompileOptions{
		Language:        constants.LanguageC,
		LimitTime:       CompileLimitTime,
		ExcludedPaths:   []string{},
		ReplacementPath: "/",
	})
	gdb, _ := gdb2.New(funcNotificationCallback2)
	r, _ := gdb.Send("file-exec-file", execFile)
	log.Println(r)

	// 添加断点
	r, _ = gdb.Send("break-insert", execFile+":4")
	println(r)
	r, _ = gdb.Send("break-insert", execFile+":6")
	println(r)
	r, _ = gdb.Send("break-insert", execFile+":7")
	println(r)
	r, _ = gdb.Send("break-insert", execFile+":9")
	println(r)

	r, _ = gdb.Send("exec-run")
	println(r)

	r, _ = gdb.Send("exec-next")
	println(r)

}

func funcNotificationCallback2(notification map[string]interface{}) {
	log.Println(notification)
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
