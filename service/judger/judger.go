package judger

import (
	"FanCode/constants"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	// 用于创建judge容器的镜像
	ImageName = "judge-docker-image"
)

type JudgeCore struct {
}

// ExecuteOption 判题需要传递参数
type ExecuteOption struct {
	ExecFile    string
	Language    int
	InputCh     <-chan []byte
	OutputCh    chan<- []byte
	ErrOutPutCh chan<- error
	ExitCh      <-chan string
	LimitTime   time.Duration
	LimitMemory int
}

func NewJudgeCore() *JudgeCore {
	return &JudgeCore{}
}

// Compile 编译，编译时在容器外进行编译的
// language: 语言类型
// compileFiles: 需要编译文件列表
// outFilePath: 输出位置
// timeout: 限制编译时间
func (j *JudgeCore) Compile(language int, compileFiles []string, outFilePath string, timeout time.Duration) error {
	if language == constants.ProgramC {
		compileFiles = append([]string{"-o", outFilePath}, compileFiles...)

		// 创建一个带有超时时间的上下文
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// 执行编译命令
		cmd := exec.CommandContext(ctx, "gcc", compileFiles...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			// 如果是由于超时导致的错误，则返回自定义错误
			if ctx.Err() == context.DeadlineExceeded {
				return errors.New("编译超时")
			}
			return err
		}

		return nil
	} else {
		return errors.New("不支持该语言")
	}
}

// Execute 运行
//
// input:
// language: 执行程序是何种语言
// execFile: 执行程序
// input: 输入管道
//
// output: 正常执行结果输出管道，异常执行结果输出管道，异常
func (j *JudgeCore) Execute(executeOption *ExecuteOption) error {

	// 根据扩展名设置执行命令
	cmd := ""
	switch executeOption.Language {
	case constants.ProgramC:
		cmd = fmt.Sprintf("%s", executeOption.ExecFile)
	case constants.ProgramJava:
		cmd = fmt.Sprintf("java -jar %s", executeOption.ExecFile)
	default:
		return fmt.Errorf("不支持该语言")
	}

	go func() {
		for {
			select {
			case inputItem := <-executeOption.InputCh:
				cmd2 := exec.Command(cmd)
				cmd2.Stdin = bytes.NewReader(inputItem)
				cmd2.Stdout = &bytes.Buffer{}
				cmd2.Stderr = &bytes.Buffer{}
				err := cmd2.Start()
				if err != nil {
					executeOption.ErrOutPutCh <- err
					break
				}
				err = cmd2.Wait()
				if err != nil {
					executeOption.ErrOutPutCh <- err
					break
				}
				if cmd2.Stderr.(*bytes.Buffer).Len() != 0 {
					executeOption.ErrOutPutCh <- errors.New(string(cmd2.Stderr.(*bytes.Buffer).Bytes()))
				}
				if cmd2.Stdout.(*bytes.Buffer).Len() != 0 {
					executeOption.OutputCh <- cmd2.Stdout.(*bytes.Buffer).Bytes()
				}
			case <-executeOption.ExitCh:
				return
			}
		}
	}()

	return nil
}
