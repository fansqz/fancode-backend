package judger

import (
	"FanCode/constants"
	"FanCode/utils"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type JudgeCore struct {
}

func NewJudgeCore() *JudgeCore {
	return &JudgeCore{}
}

// Compile 编译，编译时在容器外进行编译的
// compileFiles第个文件是main文件
func (j *JudgeCore) Compile(compileFiles []string, outFilePath string, options *CompileOptions) (*CompileResult, error) {
	result := &CompileResult{
		Compiled:         false,
		ErrorMessage:     "",
		CompiledFilePath: "",
	}

	var cmd *exec.Cmd
	var ctx context.Context
	language := j.getLanguage(options)
	switch language {
	case constants.ProgramC:
		compileFiles = append([]string{"-o", outFilePath}, compileFiles...)
		// 创建一个带有超时时间的上下文
		var cancel context.CancelFunc
		if options != nil && options.LimitTime != 0 {
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(options.LimitTime))
			defer cancel()
		} else {
			ctx = context.Background()
		}
		cmd = exec.CommandContext(ctx, "gcc", compileFiles...)
	case constants.ProgramJava:
		return j.compileJava(compileFiles, outFilePath, options)
	case constants.ProgramGo:
		compileFiles = append([]string{"build", "-o", outFilePath}, compileFiles...)
		var cancel context.CancelFunc
		if options != nil && options.LimitTime != 0 {
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(options.LimitTime))
			defer cancel()
		} else {
			ctx = context.Background()
		}
		cmd = exec.CommandContext(ctx, "go", compileFiles...)
	default:
		result.ErrorMessage = "不支持该语言\n"
		return result, nil
	}
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}

	var err error
	if err = cmd.Start(); err == nil {
		err = cmd.Wait()
	}
	if err != nil {
		return j.setErrMessageForCompileResult(ctx, cmd, err, result, options)
	}
	result.Compiled = true
	result.CompiledFilePath = outFilePath
	return result, nil
}

func (j *JudgeCore) getLanguage(options *CompileOptions) string {
	language := constants.ProgramC
	if options != nil && options.Language != "" {
		language = options.Language
	}
	return language
}

// 编译java语言会比较麻烦
func (j *JudgeCore) compileJava(compileFiles []string, outFilePath string, options *CompileOptions) (*CompileResult, error) {
	result := &CompileResult{
		Compiled:         false,
		ErrorMessage:     "",
		CompiledFilePath: outFilePath,
	}

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

	// 设置超时
	var cancel context.CancelFunc
	var ctx context.Context
	if options != nil && options.LimitTime != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(options.LimitTime))
		defer cancel()
	} else {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "javac", compileFiles...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var err error
	if err = cmd.Start(); err == nil {
		err = cmd.Wait()
	}
	if err != nil {
		return j.setErrMessageForCompileResult(ctx, cmd, err, result, options)
	}

	// 添加一些jar包必要的文件
	var file *os.File
	if file, err = os.Create(path.Join(classPath, "MANIFEST.MF")); err != nil {
		return nil, err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	//将第一个文件当作main文件
	if _, err = writer.WriteString("Manifest-Version: 1.0\nMain-Class: " + mainClass + "\nBuilt-By: fancode\n"); err != nil {
		return nil, err
	}
	if err = writer.Flush(); err != nil {
		return nil, err
	}

	// 打包成jar包
	cmd = exec.Command("jar", "cvfm", outFilePath, path.Join(classPath, "MANIFEST.MF"),
		"-C", classPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return nil, err
	}

	if err = cmd.Wait(); err != nil {
		// 如果是由于超时导致的错误，则返回自定义错误
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			result.ErrorMessage = "编译超时\n"
			return result, nil
		}
		return nil, err
	}
	result.Compiled = true
	result.CompiledFilePath = outFilePath
	return result, nil
}

func (j *JudgeCore) setErrMessageForCompileResult(ctx context.Context, cmd *exec.Cmd, err error, result *CompileResult, options *CompileOptions) (*CompileResult, error) {
	if err != nil {
		errBytes := cmd.Stderr.(*bytes.Buffer).Bytes()
		errMessage := string(errBytes)
		if len(options.ExcludedPaths) != 0 {
			errMessage = j.maskPath(string(errBytes), options.ExcludedPaths, options.ReplacementPath)
		}
		// 如果是由于超时导致的错误，则返回自定义错误
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			result.ErrorMessage = "编译超时\n" + errMessage
			return result, nil
		}
		if len(errBytes) != 0 {
			result.ErrorMessage = errMessage
			return result, nil
		}
	}
	return nil, err
}

// maskPath 函数用于屏蔽错误消息中的路径信息
func (j *JudgeCore) maskPath(errorMessage string, excludedPaths []string, replacementPath string) string {
	if errorMessage == "" {
		return ""
	}
	// 遍历需要屏蔽的敏感路径
	for _, excludedPath := range excludedPaths {
		// 如果 excludedPath 是绝对路径，但 errorMessage 中含有相对路径 "./"，则将 "./" 替换为绝对路径
		if filepath.IsAbs(excludedPath) && filepath.IsAbs("./") {
			relativePath := "." + string(filepath.Separator)
			absolutePath := filepath.Join(excludedPath, relativePath)
			errorMessage = strings.Replace(errorMessage, relativePath, absolutePath, -1)
		}

		// 构建正则表达式，匹配包含敏感路径的错误消息
		pattern := regexp.QuoteMeta(excludedPath)
		re := regexp.MustCompile(pattern)
		errorMessage = re.ReplaceAllString(errorMessage, replacementPath)
	}

	return errorMessage
}

// Execute 运行
func (j *JudgeCore) Execute(execFile string, inputCh <-chan []byte, outputCh chan<- ExecuteResult, exitCh <-chan string, options *ExecuteOptions) error {
	language := constants.ProgramC
	if options != nil && options.Language != "" {
		language = options.Language
	}
	// 根据扩展名设置执行命令
	cmdName := ""
	cmdArg := []string{}
	switch language {
	case constants.ProgramC:
		cmdName = execFile
	case constants.ProgramJava:
		cmdName = "java"
		cmdArg = []string{"-jar", execFile}
	case constants.ProgramGo:
		cmdName = execFile
	default:
		return fmt.Errorf("不支持该语言")
	}

	// 创建cgroup限制资源
	cgroup, err := NewCGroup(utils.GetUUID())
	if err != nil {
		return err
	}
	if options != nil && options.MemoryLimit != 0 {
		err = cgroup.SetMemoryLimit(options.MemoryLimit)
		if err != nil {
			return err
		}
	}
	if options != nil && options.CPUQuota != 0 {
		err = cgroup.SetCPUQuota(options.CPUQuota)
		if err != nil {
			return err
		}
	}

	go func() {
		for {
			select {
			case inputItem := <-inputCh:

				// 设置超时上下文
				var ctx context.Context
				var cancel context.CancelFunc
				if options != nil && options.LimitTime != 0 {
					ctx, cancel = context.WithTimeout(context.Background(), time.Duration(options.LimitTime)+100)
					defer cancel()
				} else {
					ctx = context.Background()
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
				if err = cmd2.Start(); err != nil {
					result.Executed = false
					result.ErrorMessage = err.Error() + "\n"
					outputCh <- result
					break
				}

				// 将进程写入cgroup组
				if err = cgroup.AddPID(cmd2.Process.Pid); err != nil {
					result.Executed = false
					result.ErrorMessage = err.Error() + "\n"
					outputCh <- result
					break
				}

				// 等待程序执行
				cmd2.Wait()
				// 读取使用cpu和内存，以及执行时间
				rusage := cmd2.ProcessState.SysUsage().(*syscall.Rusage)
				result.UsedCpuTime = rusage.Utime.Sec*1000 + rusage.Utime.Usec/1000
				result.UsedMemory = rusage.Maxrss * 1024
				result.UsedTime = int64(time.Now().Sub(beginTime))

				// 输出的错误信息
				errMessage := string(cmd2.Stderr.(*bytes.Buffer).Bytes())
				if options != nil && len(options.ExcludedPaths) != 0 {
					errMessage = j.maskPath(errMessage, options.ExcludedPaths, options.ReplacementPath)
				}
				outMessage := cmd2.Stdout.(*bytes.Buffer).Bytes()
				// 检测内存占用，cpu占用，以及执行时间
				if options != nil && options.LimitTime < result.UsedTime {
					result.Executed = false
					result.ErrorMessage = "运行超时\n"
				} else if options != nil && options.MemoryLimit < result.UsedMemory {
					result.Executed = false
					result.ErrorMessage = "内存超出限制\n"
				} else if options != nil && options.CPUQuota < result.UsedCpuTime {
					result.Executed = false
					result.ErrorMessage = "cpu超出限制\n"
				} else if len(errMessage) != 0 {
					result.Executed = false
					result.ErrorMessage = errMessage
				} else {
					result.Executed = true
					result.Output = outMessage
				}
				outputCh <- result
			case <-exitCh:
				cgroup.Release()
				return
			}
		}
	}()

	return nil
}
