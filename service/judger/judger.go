package judger

import (
	"FanCode/constants"
	"FanCode/utils"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"time"
)

const (
	// cgorup相关参数，用于限制系统资源
	memoryLimitFile  = "memory.limit_in_bytes"
	swapLimitFile    = "memory.swappiness"
	cgroupMemoryRoot = "/sys/fs/cgroup/memory"
)

type JudgeCore struct {
}

func NewJudgeCore() *JudgeCore {
	return &JudgeCore{}
}

// Compile 编译，编译时在容器外进行编译的
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

	// 创建cgroup限制资源
	uuid := utils.GetUUID()
	cgroupPath := path.Join(cgroupMemoryRoot, uuid)
	err := os.MkdirAll(cgroupPath, 0755)
	if err != nil {
		return fmt.Errorf("创建cgroup组出错")
	}
	// 设置内存限制
	limitMemory := fmt.Sprintf("%d", executeOption.LimitMemory)
	err = os.WriteFile(filepath.Join(cgroupPath, memoryLimitFile), []byte(limitMemory), 0644)
	if err != nil {
		return fmt.Errorf("cgroup限制内存出错")
	}
	// 限制交换空间
	err = os.WriteFile(filepath.Join(cgroupPath, swapLimitFile), []byte("0"), 0644)
	if err != nil {
		return fmt.Errorf("cgroup限制交换空间")
	}

	go func() {
		for {
			select {
			case inputItem := <-executeOption.InputCh:
				ctx, cancel := context.WithTimeout(context.Background(), executeOption.LimitTime)
				defer cancel()

				if err != nil {
					executeOption.OutputCh <- ExecuteResult{
						Executed: false,
						Error:    err,
					}
					break
				}

				// 创建子进程，并将其加入cgroup
				cmd2 := exec.CommandContext(ctx, "cgexec", "-g", fmt.Sprintf("memory:"+uuid), cmd)
				cmd2.Stdin = bytes.NewReader(inputItem)
				cmd2.Stdout = &bytes.Buffer{}
				cmd2.Stderr = &bytes.Buffer{}
				cmd2.SysProcAttr = &syscall.SysProcAttr{
					Setpgid: true,
				}
				result := ExecuteResult{}

				err = cmd2.Start()
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
						err = syscall.Kill(-cmd2.Process.Pid, syscall.SIGKILL)
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
				// 删除cgroup文件
				err = os.Remove(cgroupPath)
				if err != nil {
					log.Println(err)
				}
				return
			}
		}
	}()

	return nil
}
