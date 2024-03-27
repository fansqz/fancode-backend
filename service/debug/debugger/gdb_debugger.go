package debugger

import (
	"FanCode/constants"
	"FanCode/service/judger"
	"FanCode/utils"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	GdbInitFileName  = "init.gdb"
	CompileLimitTime = 1000 * 10
)

type gdbDebugger struct {
	gdbProcess  *exec.Cmd // 调试进程
	userProcess *exec.Cmd // 用户进程
	workPath    string    // 工作目录
	judger      *judger.JudgeCore
	gdbStdin    io.WriteCloser // gdb相关
	gdbStdout   io.ReadCloser
	gdbStderr   io.ReadCloser

	// 请求管道
	reqChan  chan interface{}
	execChan chan interface{}

	// 保证map的线程安全
	lock sync.Mutex
	// 断点映射
	// dap中没有断点编号，但是gdb却有，该映射是(file:line):number的映射
	breakpointMap map[string]string
	// userChan 通过该chan将数据返回给用户
	userChan chan interface{}
}

func NewGdbDebugger() Debugger {
	return &gdbDebugger{
		judger: judger.NewJudgeCore(),
	}
}

func (g *gdbDebugger) Launch(compileFiles []string, workPath string) (chan interface{}, error) {
	// 设置gdb文件
	if err := g.configGDB(g.workPath); err != nil {
		return nil, err
	}
	cha := make(chan interface{}, 10)
	g.userChan = cha
	// 启动协程编译运行
	go func() {
		execFile := path.Join(workPath, "main")
		// 进行编译
		r, err := g.judger.Compile(compileFiles, execFile, &judger.CompileOptions{
			Language:        constants.LanguageC,
			LimitTime:       CompileLimitTime,
			ExcludedPaths:   []string{workPath},
			ReplacementPath: "/",
		})
		if err != nil {
			cha <- &LaunchEvent{
				Success: false,
				Message: err.Error(),
			}
		}
		if !r.Compiled {
			cha <- &LaunchEvent{
				Success: false,
				Message: r.ErrorMessage,
			}
		}

		// gdb调试
		g.gdbProcess = exec.Command("gdb", "--interpreter=mi", execFile)
		g.gdbStdin, _ = g.gdbProcess.StdinPipe()
		g.gdbStdout, _ = g.gdbProcess.StdoutPipe()
		g.gdbStderr, _ = g.gdbProcess.StderrPipe()
		g.gdbProcess.Start()
		// 启动协程处理用户输出和gdb输出
		go g.processGdbData()
	}()
	return cha, nil
}

func (g *gdbDebugger) Start() error {
	_, err := g.gdbStdin.Write([]byte("-exec-run\n"))
	// 记录每个请求
	g.reqChan <- &startReq{}
	return err
}

// DeadlineRead 设置超时的读取，超出超时事件就会关闭Read
func (g *gdbDebugger) readWithTimeoutAndDelimiter(read io.ReadCloser, timeout time.Duration, end string) ([]byte, error) {
	// 创建一个缓冲区用于存储读取的数据
	buf := make([]byte, 1024)

	// 创建一个 channel 用于通知读取完成
	doneData := make(chan []byte)
	doneErr := make(chan error)

	// 在一个新的 goroutine 中进行读取操作
	go func() {
		var data []byte
		for {
			// 继续读
			n, err := read.Read(buf)
			if err != nil {
				doneErr <- err
				break
			} else {
				if data == nil {
					data = buf[0:n]
				} else {
					data = append(data, buf[0:n]...)
				}
				if strings.HasSuffix(string(data), end) {
					doneData <- data
					break
				}
			}
		}
	}()

	for {
		select {
		case data := <-doneData:
			return data, nil
		case err := <-doneErr:
			return nil, err
		}
	}

}

// closePipe 关闭所有stdout stdin stderr
// todo: 异常如何处理
func (g *gdbDebugger) closePipe() {
	_ = g.gdbStdin.Close()
	_ = g.gdbStdout.Close()
	_ = g.gdbStderr.Close()
}

func (g *gdbDebugger) processGdbData() {
	b := make([]byte, 6*1024)
	for {
		n, err := g.gdbStdout.Read(b)
		if err != nil {
			if err != io.EOF {
				// 如果不是EOF，打印出错误信息
				log.Printf("读取数据时发生错误: %v", err)
			}
			// todo 这里需要什么策略
			break // 无论是EOF错误还是其他错误，都退出循环
		}
		output := string(b[0:n])
		g.handleOutput(output)
	}
}

// 解析GDB输出行
func (g *gdbDebugger) handleOutput(output string) {
	output = strings.Trim(strings.Trim(output, "gdb"), "\n")
	lines := strings.Split(output, "\n")
	userOutput := ""
	i := 0
	for i < len(lines) {
		line := lines[i]
		switch {
		case strings.HasPrefix(line, "=thread-group-added"):
			// 开启一个线程
			i++
			nextLine := lines[i]

			// 启动gdb的命令
			if strings.HasPrefix(nextLine, "~\"GNU gdb") {
				g.userChan <- &LaunchEvent{
					Success: true,
				}
			}
		case strings.HasPrefix(line, "@"):
			// Target output (程序输出)
			userOutput = userOutput + line[1:]
		case strings.HasPrefix(line, "&"):
			// Log stream output (日志)
		case strings.HasPrefix(line, "*"), strings.HasPrefix(line, "+"), strings.HasPrefix(line, "="):
			// Async record (异步消息)
		case strings.HasPrefix(line, "^done"):
			g.handleDoneResp(line)
		case strings.HasPrefix(line, "^running"):

		default:
			fmt.Printf("Unknown line: %s\n", line)
		}
		i++
	}
	if userOutput != "" {
		g.userChan <- OutputEvent{
			Category: constants.Stdout,
			Output:   userOutput,
		}
	}
}

// 判断该done请求是响应的是什么
func (g *gdbDebugger) handleDoneResp(line string) {
	// 获取一个request，判断有啥结果
	for {
		req := <-g.reqChan
		if _, ok := req.(*addBreakpointReq); ok {
			// 添加断点的响应
			m, okk := g.parseAddBpOutput(line)
			if !okk {
				log.Println("断点信息解析失败")
				break
			}
			g.breakpointMap[m["file"]+":"+m["line"]] = m["number"]
			// 断点事件
			line, _ := strconv.Atoi(m["line"])
			g.userChan <- &BreakpointEvent{
				Reason: constants.NewType,
				Breakpoint: Breakpoint{
					File: m["file"],
					Line: line,
				},
			}
		} else if rbReq, ok := req.(*removeBreakpointReq); ok {
			delete(g.breakpointMap, rbReq.BP.File+":"+strconv.Itoa(rbReq.BP.Line))
			// 断点事件
			g.userChan <- &BreakpointEvent{
				Reason: constants.RemovedType,
				Breakpoint: Breakpoint{
					File: rbReq.BP.File,
					Line: rbReq.BP.Line,
				},
			}
		}

	}
}

func (g *gdbDebugger) parseAddBpOutput(gdbOutput string) (map[string]string, bool) {
	if !regexp.MustCompile(`^\^done,bkpt=`).MatchString(gdbOutput) {
		return nil, false
	}

	trimmedOutput := regexp.MustCompile(`^\^done,bkpt=\{(.*)}$`).FindStringSubmatch(gdbOutput)
	if len(trimmedOutput) < 2 {
		fmt.Println("No content within braces found.")
		return nil, false
	}

	matches := regexp.MustCompile(`(\w+)="([^"]+)"`).FindAllStringSubmatch(trimmedOutput[1], -1)
	result := make(map[string]string)
	for _, match := range matches {
		if len(match) == 3 {
			result[match[1]] = match[2]
		}
	}
	return result, true
}

func (g *gdbDebugger) SendToConsole(input string) error {
	_, err := g.gdbStdin.Write([]byte("-interpreter-exec console \"" + input + "\"\n"))
	return err
}

func (g *gdbDebugger) Next() error {
	_, err := g.gdbStdin.Write([]byte("-exec-next\n"))
	return err
}

func (g *gdbDebugger) Step() error {
	_, err := g.gdbStdin.Write([]byte("-exec-step\n"))
	return err
}

func (g *gdbDebugger) Continue() error {
	_, err := g.gdbStdin.Write([]byte("-exec-continue\n"))
	return err
}

// todo: 添加断点和删除断点串行执行
func (g *gdbDebugger) AddBreakpoints(breakpoints []Breakpoint) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	isError := false
	for _, bp := range breakpoints {
		bs := utils.Slice("-break-insert " + bp.File + ":" + strconv.Itoa(bp.Line) + "\n")
		g.reqChan <- &addBreakpointReq{
			BP: bp,
		}
		n, err := g.gdbStdin.Write(bs)
		if err != nil || n != len(bs) {
			isError = true
			continue
		}
	}
	if isError {
		return errors.New("断点添加失败")
	}
	return nil
}

func (g *gdbDebugger) RemoveBreakpoints(breakpoints []Breakpoint) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	isError := false
	for _, bp := range breakpoints {
		number, ok := g.breakpointMap[bp.File+":"+strconv.Itoa(bp.Line)]
		if !ok {
			continue
		}
		bs := utils.Slice("-break-delete " + number + "\n")
		g.reqChan <- &removeBreakpointReq{
			BP: bp,
		}
		n, err := g.gdbStdin.Write(bs)
		if err != nil || n != len(bs) {
			isError = true
			continue
		}
	}
	if isError {
		return errors.New("断点删除失败")
	}
	return nil
}

// 判断消息类型
func (g *gdbDebugger) getMessageType(message []byte) string {
	return ""
}

// Terminate todo: 如何正确处理异常
func (g *gdbDebugger) Terminate() error {
	g.closePipe()
	_ = g.gdbProcess.Process.Kill()
	_ = g.userProcess.Process.Kill()
	g.gdbProcess = nil
	g.userProcess = nil
	g.gdbStdin = nil
	g.gdbStdout = nil
	g.gdbStderr = nil
	// 清除WorkPath
	os.RemoveAll(g.workPath)
	g.workPath = ""
	return nil
}

func (d *gdbDebugger) configGDB(workPath string) error {
	fileStr := "set print elements 0\n" +
		"set print null-stop on\n" +
		"set print repeats 0\n" +
		"set print union on\n" +
		"set width 0\n"
	return os.WriteFile(path.Join(workPath, GdbInitFileName), []byte(fileStr), 0644)
}

func (d *gdbDebugger) getTimeoutContext(limitTime int64) (context.Context, context.CancelFunc) {
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

type startReq struct {
}

type addBreakpointReq struct {
	BP Breakpoint
}

type removeBreakpointReq struct {
	BP Breakpoint
}
