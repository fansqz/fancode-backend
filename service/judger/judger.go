package judger

import (
	"FanCode/constants"
	"FanCode/utils"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	// cgorup相关参数，用于限制系统资源
	procsFile        = "cgroup.procs"
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
// compileFiles第个文件是main文件
func (j *JudgeCore) Compile(language string, compileFiles []string, outFilePath string, timeout time.Duration) error {
	var cmd *exec.Cmd
	var ctx context.Context
	switch language {
	case constants.ProgramC:
		compileFiles = append([]string{"-o", outFilePath}, compileFiles...)
		// 创建一个带有超时时间的上下文
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, "gcc", compileFiles...)
	case constants.ProgramJava:
		return j.compileJava(compileFiles, outFilePath, timeout)
	case constants.ProgramGo:
		compileFiles = append([]string{"build", "-o", outFilePath}, compileFiles...)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, "go", compileFiles...)
	default:
		return errors.New("不支持该语言")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		// 如果是由于超时导致的错误，则返回自定义错误
		if ctx.Err() == context.DeadlineExceeded {
			return errors.New("编译超时")
		}
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}

	// 如果是java还需要将编译文件打包成jar包
	if language == constants.ProgramJava {

	}

	return err
}

// 编译java语言会比较麻烦
func (j *JudgeCore) compileJava(compileFiles []string, outFilePath string, timeout time.Duration) error {
	lastIndex := strings.LastIndex(outFilePath, "/")
	// 读取main文件，规定第一个文件时main文件
	mainClass := compileFiles[0][strings.LastIndex(compileFiles[0], "/")+1:]
	mainClass = strings.Split(mainClass, ".")[0]

	// 创建存放class文件的目录
	classPath := path.Join(outFilePath[0:lastIndex], "classPath")
	os.MkdirAll(classPath, os.ModePerm)
	defer os.RemoveAll(classPath)

	// 编译为class文件
	compileFiles = append([]string{"-d", classPath}, compileFiles...)
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "javac", compileFiles...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		// 如果是由于超时导致的错误，则返回自定义错误
		if ctx.Err() == context.DeadlineExceeded {
			return errors.New("编译超时")
		}
		return err
	}

	// 添加一些jar包必要的文件
	file, err := os.Create(path.Join(classPath, "MANIFEST.MF"))
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	//将第一个文件当作main文件
	_, err = writer.WriteString("Manifest-Version: 1.0\nMain-Class: " + mainClass + "\nBuilt-By: fancode\n")
	if err != nil {
		return err
	}
	err = writer.Flush()

	// 打包成jar包
	cmd = exec.Command("jar", "cvfm", outFilePath, path.Join(classPath, "MANIFEST.MF"),
		"-C", classPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	// 如果是由于超时导致的错误，则返回自定义错误
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New("编译超时")
	}
	return err
}

// Execute 运行
// todo: 计算空间的使用
func (j *JudgeCore) Execute(executeOption *ExecuteOption) error {
	// 根据扩展名设置执行命令
	cmdName := ""
	cmdArg := []string{}
	switch executeOption.Language {
	case constants.ProgramC:
		cmdName = executeOption.ExecFile
	case constants.ProgramJava:
		cmdName = "java"
		cmdArg = []string{"-jar", executeOption.ExecFile}
	case constants.ProgramGo:
		cmdName = executeOption.ExecFile
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
	limitMemoryPath := filepath.Join(cgroupPath, memoryLimitFile)
	err = os.WriteFile(limitMemoryPath, []byte(limitMemory), 0755)
	if err != nil {
		return fmt.Errorf("cgroup限制内存出错")
	}
	// 限制交换空
	limitSwapPath := filepath.Join(cgroupPath, swapLimitFile)
	err = os.WriteFile(limitSwapPath, []byte("0"), 0755)
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
				cmd2 := exec.CommandContext(ctx, cmdName, cmdArg...)
				result := ExecuteResult{}

				cmd2.Stdin = bytes.NewReader(inputItem)
				cmd2.Stdout = &bytes.Buffer{}
				cmd2.Stderr = &bytes.Buffer{}
				cmd2.SysProcAttr = &syscall.SysProcAttr{
					Setpgid: true,
				}

				beginTime := time.Now()
				err = cmd2.Start()

				if err != nil {
					result.Executed = false
					result.Error = errors.New(err.Error() + "\n" +
						string(cmd2.Stderr.(*bytes.Buffer).Bytes()))
					executeOption.OutputCh <- result
					break
				}

				// 将进程写入cgroup组
				pPath := path.Join(cgroupPath, procsFile)
				err = os.WriteFile(pPath, []byte(fmt.Sprintf("%d", cmd2.Process.Pid)), 0755)
				if err != nil {
					result.Executed = false
					result.Error = err
					executeOption.OutputCh <- result
					break
				}

				// 等待程序执行
				err = cmd2.Wait()
				endTime := time.Now()
				if err != nil {
					result.Executed = false
					if ctx.Err() == context.DeadlineExceeded {
						result.Error = ExecuteTimoutErr
					} else {
						result.Error = err
					}
					executeOption.OutputCh <- result
					break
				}

				// 读取结果
				if cmd2.Stderr.(*bytes.Buffer).Len() != 0 {
					result.Executed = false
					result.Error = errors.New(string(cmd2.Stderr.(*bytes.Buffer).Bytes()))
					executeOption.OutputCh <- result
				} else if cmd2.Stdout.(*bytes.Buffer).Len() != 0 {
					result.Executed = true
					result.TimeUsed = endTime.Sub(beginTime)
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
