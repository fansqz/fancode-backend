package debug

import (
	"FanCode/constants"
	de "FanCode/service/debug/define"
	"FanCode/service/judger"
	"FanCode/utils"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	GdbInitFileName  = "init.gdb"
	CompileLimitTime = 1000 * 10
)

type gdbCore struct {
	gdbProcess  *exec.Cmd       // 调试进程
	userProcess *exec.Cmd       // 用户进程
	breakpoints []de.Breakpoint // 断点
	workPath    string          // 工作目录
	debugFile   string          // 调试关注的文件
	judger      *judger.JudgeCore
	gdbStdin    io.WriteCloser // gdb相关
	gdbStdout   io.ReadCloser
	gdbStderr   io.ReadCloser
	userStdin   io.WriteCloser // user相关
	userStdout  io.ReadCloser
	userStderr  io.ReadCloser

	// 消息channl
	gdbMessageChan  chan []byte
	userMessageChan chan []byte
}

func NewGdbCore() de.DebugCore {
	return &gdbCore{
		judger: judger.NewJudgeCore(),
	}
}

func (g *gdbCore) Launch(compileFiles []string, workPath string) (chan interface{}, error) {
	// 设置gdb文件
	if err := g.configGDB(g.workPath); err != nil {
		return nil, err
	}
	cha := make(chan interface{}, 10)
	// 启动协程编译运行
	go func() {
		execFile := path.Join(workPath, "main")
		// 进行编译
		r, err := g.judger.Compile(compileFiles, execFile, &judger.CompileOptions{
			Language:        constants.ProgramC,
			LimitTime:       CompileLimitTime,
			ExcludedPaths:   []string{workPath},
			ReplacementPath: "/",
		})
		if err != nil {
			cha <- &de.CompileEvent{
				Success: false,
				Message: err.Error(),
			}
		}
		if !r.Compiled {
			cha <- &de.CompileEvent{
				Success: false,
				Message: r.ErrorMessage,
			}
		} else {
			cha <- &de.CompileEvent{
				Success: true,
				Message: "编译成功",
			}
		}

		// 启动用户程序窗口
		g.userProcess = exec.Command("bash")
		g.userStdin, _ = g.userProcess.StdinPipe()
		g.userStdout, _ = g.userProcess.StdoutPipe()
		g.userStderr, _ = g.userProcess.StderrPipe()
		// tty
		_, err = g.userStdin.Write(utils.Slice("tty\n"))
		if err != nil {
			cha <- &de.CompileEvent{
				Success: false,
				Message: "启动用户程序失败",
			}
			return
		}
		tty, err := g.deadlineRead(g.userStdout, 200*time.Millisecond)
		if err != nil {
			cha <- &de.CompileEvent{
				Success: false,
				Message: "启动用户程序失败",
			}
			return
		}
		if len(tty) > 0 && tty[len(tty)-1] == '\n' {
			tty = tty[:len(tty)-1]
		}
		// gdb调试窗口
		g.gdbProcess = exec.Command("gdb", "-tty", utils.String(tty), execFile)
		g.gdbStdin, _ = g.gdbProcess.StdinPipe()
		g.gdbStdout, _ = g.gdbProcess.StdoutPipe()
		g.gdbStderr, _ = g.gdbProcess.StderrPipe()
	}()
	return cha, nil
}

func (g *gdbCore) Start() error {
	return g.sendCommandToGDB("run")
}

// DeadlineRead 设置超时的读取，超出超时事件就会关闭Read
func (g *gdbCore) deadlineRead(read io.ReadCloser, timeout time.Duration) ([]byte, error) {
	// 创建一个缓冲区用于存储读取的数据
	buf := new(bytes.Buffer)

	// 创建一个 channel 用于通知读取完成
	done := make(chan error)

	// 在一个新的 goroutine 中进行读取操作
	go func() {
		_, err := buf.ReadFrom(read)
		done <- err
	}()

	// 创建一个定时器，在超时后关闭读取管道
	timer := time.AfterFunc(timeout, func() {
		read.Close()
	})

	select {
	case err := <-done:
		// 如果读取完成，停止定时器，并返回结果
		timer.Stop()
		return buf.Bytes(), err
	case <-time.After(timeout):
		// 如果超时，返回空结果和一个错误
		return nil, fmt.Errorf("read operation timed out")
	}
}

// closePipe 关闭所有stdout stdin stderr
// todo: 异常如何处理
func (g *gdbCore) closePipe() {
	_ = g.userStdin.Close()
	_ = g.userStdout.Close()
	_ = g.userStderr.Close()
	_ = g.gdbStdin.Close()
	_ = g.gdbStdout.Close()
	_ = g.gdbStderr.Close()
}

// sendCommandToGDB 向gdb输入命令
func (g *gdbCore) sendCommandToGDB(cmd string) error {
	// 确保命令以换行符结束，这是 GDB 命令的格式
	if !strings.HasSuffix(cmd, "\n") {
		cmd += "\n"
	}

	// Write方法返回的写入字节数和可能出现的错误
	nbytes, err := g.gdbStdin.Write([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to send command to GDB: %w", err)
	}

	// 为了确保所有的命令都被写入，检查写入的字节数是否和你期望的一致
	if nbytes != len(cmd) {
		return fmt.Errorf("failed to send complete command to GDB: expected to write %d bytes but wrote %d", len(cmd), nbytes)
	}

	return nil
}

func (g *gdbCore) SendToConsole(input string) error {
	n, err := g.userStdin.Write(utils.Slice(input))
	if err != nil {
		return fmt.Errorf("failed to send input to user process: %w", err)
	}
	// 为了确保所有的命令都被写入，检查写入的字节数是否和你期望的一致
	if n != len(input) {
		return fmt.Errorf("failed to send complete input to user process: expected to write %d bytes but wrote %d", len(input), n)
	}
	return nil
}

func (g *gdbCore) Next() error {
	return g.sendCommandToGDB("n 1\n")
}

func (g *gdbCore) Step() error {
	return g.sendCommandToGDB("s 1\n")
}

func (g *gdbCore) Continue() error {
	return g.sendCommandToGDB("c 1\n")
}

func (g *gdbCore) AddBreakpoints(source string, breakpoints []de.Breakpoint) ([]de.Breakpoint, error) {
	for _, bp := range breakpoints {
		bs := utils.Slice("b " + source + ":" + strconv.Itoa(bp.Line) + "\n")
		n, err := g.gdbStdin.Write(bs)
		if err != nil || n != len(bs) {
			continue
		}
		// 读取断点事件
		message := <-g.gdbMessageChan
		if g.getMessageType(message) != "addBreakpoint" {
			go func() {
				g.handlerGdbMessage(message)
			}()
		} else {

		}
	}
	// todo
	return nil, nil
}

func (g *gdbCore) RemoveBreakpoints(source string, breakpoints []de.Breakpoint) ([]de.Breakpoint, error) {
	// todo
	return nil, nil
}

func (g *gdbCore) handlerGdbMessage(message []byte) {

}

// 判断消息类型
func (g *gdbCore) getMessageType(message []byte) string {
	return ""
}

// Terminate todo: 如何正确处理异常
func (g *gdbCore) Terminate() error {
	g.closePipe()
	_ = g.gdbProcess.Process.Kill()
	_ = g.userProcess.Process.Kill()
	g.gdbProcess = nil
	g.userProcess = nil
	g.gdbStdin = nil
	g.gdbStdout = nil
	g.gdbStderr = nil
	g.userStdin = nil
	g.userStdout = nil
	g.userStderr = nil
	// 清除WorkPath
	os.RemoveAll(g.workPath)
	g.workPath = ""
	return nil
}

func (d *gdbCore) configGDB(workPath string) error {
	fileStr := "set print elements 0\n" +
		"set print null-stop on\n" +
		"set print repeats 0\n" +
		"set print union on\n" +
		"set width 0\n"
	return os.WriteFile(path.Join(workPath, GdbInitFileName), []byte(fileStr), 0644)
}

func (d *gdbCore) getTimeoutContext(limitTime int64) (context.Context, context.CancelFunc) {
	var ctx context.Context
	var cancel context.CancelFunc
	if limitTime != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(limitTime))
	} else {
		ctx = context.Background()
	}
	cancel = func() {}
	return ctx, cancel
}
