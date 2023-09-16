package judger

import (
	"FanCode/constants"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type JudgeCore struct {
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
				ctx, cancel := context.WithTimeout(context.Background(), executeOption.LimitTime)
				defer cancel()

				cmd2 := exec.CommandContext(ctx, cmd)
				cmd2.Stdin = bytes.NewReader(inputItem)
				cmd2.Stdout = &bytes.Buffer{}
				cmd2.Stderr = &bytes.Buffer{}
				result := ExecuteResult{}

				err := cmd2.Start()
				if err != nil {
					result.Executed = false
					result.Error = err
					executeOption.OutputCh <- result
					break
				}

				// 等待程序执行
				err = cmd2.Wait()
				if err != nil {
					if ctx.Err() == context.DeadlineExceeded {
						result.Executed = false
						result.Error = ExecuteTimoutErr
						executeOption.OutputCh <- result
						err = cmd2.Process.Kill()
						if err != nil {
							log.Println(err)
						}
					}
					break
				}

				// 读取结果
				if cmd2.Stderr.(*bytes.Buffer).Len() != 0 {
					result.Executed = false
					result.Error = errors.New(string(cmd2.Stderr.(*bytes.Buffer).Bytes()))
					executeOption.OutputCh <- result
				} else if cmd2.Stdout.(*bytes.Buffer).Len() != 0 {
					result.Executed = true
					result.Output = cmd2.Stdout.(*bytes.Buffer).Bytes()
					executeOption.OutputCh <- result
				}
			case <-executeOption.ExitCh:
				return
			}
		}
	}()

	return nil
}
