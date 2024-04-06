package debugger

import (
	"FanCode/constants"
	"FanCode/service/judger"
	"context"
	"fmt"
	"github.com/cyrus-and/gdb"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	GdbInitFileName  = "init.gdb"
	CompileLimitTime = int64(10 * time.Second)
)

type gdbDebugger struct {
	workPath string // 工作目录
	judger   *judger.JudgeCore

	gdb *gdb.Gdb
	// 请求管道
	reqChan chan interface{}

	// 保证map的线程安全
	lock sync.Mutex
	// 断点映射
	// dap中没有断点编号，但是gdb却有，该映射是(file:line):number的映射
	breakpointMap map[string]string
	// userChan 通过该chan将数据返回给用户
	userChan chan interface{}
}

func NewGdbDebugger() Debugger {
	// 创建gdb对象
	gdb, err := gdb.New(funcNotificationCallback)
	if err != nil {
		log.Println(err)
	}
	return &gdbDebugger{
		gdb:           gdb,
		breakpointMap: make(map[string]string, 10),
		reqChan:       make(chan interface{}, 20),
		judger:        judger.NewJudgeCore(),
	}
}

func (g *gdbDebugger) Launch(compileFiles []string, workPath string) (chan interface{}, error) {
	// 设置gdb文件
	g.workPath = workPath
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

		// 创建命令
		g.gdb.Send("file-exec-file", execFile)
	}()
	return cha, nil
}

func funcNotificationCallback(notification map[string]interface{}) {
	log.Println(notification)
}

func (g *gdbDebugger) Start() error {
	log.Println("[gdb_debugger] start")
	_, err := g.gdb.Send("exec-run")
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
	//_ = g.gdbStdin.Close()
	//_ = g.gdbStdout.Close()
	//_ = g.gdbStderr.Close()
}

// processGdbData 循环处理gdb的输出命令
//func (g *gdbDebugger) processGdbData() {
//	b := make([]byte, 6*1024)
//	for {
//		n, err := g.gdbStdout.Read(b)
//		if err != nil {
//			if err != io.EOF {
//				// 如果不是EOF，打印出错误信息
//				log.Printf("读取数据时发生错误: %v", err)
//			}
//			// todo 这里需要什么策略
//			break // 无论是EOF错误还是其他错误，都退出循环
//		}
//		output := string(b[0:n])
//		g.handleOutput(output)
//	}
//}

// handleOutput 解析GDB输出行
func (g *gdbDebugger) handleOutput(output string) {
	log.Printf("[gdb]%s", output)
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
		case strings.HasPrefix(line, "*stopped"):
			g.handleStoppedResp(line)
		case strings.HasPrefix(line, "^done"):
			g.handleDoneResp(line)
		case strings.HasPrefix(line, "*running"):
			g.handleRunningResp(line)
		case strings.HasPrefix(line, "(gdb)\n"):
			continue
		default:
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
// `^done` 结果记录主要用于查询命令或那些会立即完成并返回结果的命令，
// 例如设置断点、检查变量值、改变栈帧等。当这些命令成功完成处理时，
// GDB 发送 `^done`，可能还会跟随附加信息提供命令的输出结果。
func (g *gdbDebugger) handleDoneResp(line string) {
	// 获取一个request，判断有啥结果
	req := <-g.reqChan
	if _, ok := req.(*addBreakpointReq); ok {
		// 添加断点的响应
		m, okk := g.parseAddBpOutput(line)
		if !okk {
			log.Println("断点信息解析失败")
		}
		g.breakpointMap[m["file"]+":"+m["line"]] = m["number"]
		// 断点事件
		line, _ := strconv.Atoi(m["line"])
		breakpointEvent := &BreakpointEvent{
			Reason: constants.NewType,
			Breakpoints: []Breakpoint{{
				File: m["file"],
				Line: line,
			}},
		}
		g.userChan <- breakpointEvent
		log.Printf("[gdb_debugger] add breakpoint req %v\n", breakpointEvent)
	} else if rbReq, ok := req.(*removeBreakpointReq); ok {
		delete(g.breakpointMap, rbReq.BP.File+":"+strconv.Itoa(rbReq.BP.Line))
		// 断点事件
		breakpointEvent := &BreakpointEvent{
			Reason: constants.RemovedType,
			Breakpoints: []Breakpoint{{
				File: rbReq.BP.File,
				Line: rbReq.BP.Line,
			}},
		}
		g.userChan <- breakpointEvent
		log.Printf("[gdb_debugger] remove breakpoint req %v\n", breakpointEvent)
	}
}

// *stopped开头的gdb命令，说明gdb调试到了某个地方停止了
func (g *gdbDebugger) handleStoppedResp(input string) {

	input = strings.Trim(input, "*stopped")
	fields := strings.Split(input, ",")
	reasonPat := regexp.MustCompile(`reason="([^"]+)"`)
	framePat := regexp.MustCompile(`frame=\{([^}]+)}`)

	reason := ""
	file := ""
	line := 0
	for _, field := range fields {
		if reasonPat.MatchString(field) {
			reason = reasonPat.FindStringSubmatch(field)[1]
		}
		if framePat.MatchString(field) {
			frameData := framePat.FindStringSubmatch(field)[1]
			frameFields := strings.Split(frameData, ",")
			for _, f := range frameFields {
				parts := strings.Split(f, "=")
				key := strings.Trim(parts[0], "\" ")
				value := strings.Trim(parts[1], "\" ")
				switch key {
				case "file":
					file = value
				case "line":
					fmt.Sscanf(value, "%d", &line)
				}
			}
		}
	}
	var stoppedReason constants.StoppedReasonType
	if reason == "breakpoint-hit" {
		stoppedReason = constants.BreakpointStopped
	} else if reason == "end-stepping-range" {
		stoppedReason = constants.StepStopped
	} else if reason == "exited-normally" {
		// 程序退出
		exitedEvent := &ExitedEvent{
			ExitCode: 0,
		}
		g.userChan <- exitedEvent
		return
	}
	stoppedEvent := &StoppedEvent{
		Reason: stoppedReason,
		File:   g.maskPath(file),
		Line:   line,
	}
	g.userChan <- stoppedEvent
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
	result["file"] = g.maskPath(result["file"])
	return result, true
}

func (g *gdbDebugger) handleRunningResp(line string) {
	continueEvent := &ContinuedEvent{}
	g.userChan <- continueEvent
}

// processUserData 循环处理用户输出
//func (g *gdbDebugger) processUserData() {
//	// 读取伪终端的输出流来获取被调试程序的输出
//	scanner := bufio.NewScanner(g.ptyMaster)
//	for scanner.Scan() {
//		output := scanner.Text()
//		// 处理被调试程序的输出
//		g.userChan <- &OutputEvent{
//			Output:   output,
//			Category: constants.Stdout,
//		}
//	}
//	if err := scanner.Err(); err != nil {
//		fmt.Fprintln(os.Stderr, "Error reading from pty:", err)
//	}
//
//}

func (g *gdbDebugger) SendToConsole(input string) error {
	// 向伪终端主设备发送数据
	_, err := g.gdb.Write([]byte(input))
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

func (g *gdbDebugger) Next() error {
	resp, err := g.gdb.Send("exec-next")
	log.Println(resp)
	return err
}

func (g *gdbDebugger) Step() error {
	resp, err := g.gdb.Send("-exec-step")
	log.Println(resp)
	return err
}

func (g *gdbDebugger) Continue() error {
	_, err := g.gdb.Send("-exec-continue")
	return err
}

// todo: 添加断点和删除断点串行执行
func (g *gdbDebugger) AddBreakpoints(breakpoints []Breakpoint) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	for _, bp := range breakpoints {
		g.reqChan <- &addBreakpointReq{
			BP: bp,
		}
		_, err := g.gdb.Send("break-insert", path.Join(g.workPath, bp.File)+":"+strconv.Itoa(bp.Line))
		if err != nil {
			continue
		}
	}
	return nil
}

func (g *gdbDebugger) RemoveBreakpoints(breakpoints []Breakpoint) error {
	//g.lock.Lock()
	//defer g.lock.Unlock()
	//isError := false
	//for _, bp := range breakpoints {
	//	if bp.File[0] != '/' {
	//		bp.File = "/" + bp.File
	//	}
	//	number, ok := g.breakpointMap[bp.File+":"+strconv.Itoa(bp.Line)]
	//	if !ok {
	//		continue
	//	}
	//	bs := utils.Slice("-break-delete " + number + "\n")
	//	g.reqChan <- &removeBreakpointReq{
	//		BP: bp,
	//	}
	//	n, err := g.gdbStdin.Write(bs)
	//	if err != nil || n != len(bs) {
	//		isError = true
	//		continue
	//	}
	//}
	//if isError {
	//	return errors.New("断点删除失败")
	//}
	return nil
}

// 判断消息类型
func (g *gdbDebugger) getMessageType(message []byte) string {
	return ""
}

// Terminate todo: 如何正确处理异常
func (g *gdbDebugger) Terminate() error {
	//g.closePipe()
	//_ = g.gdbProcess.Process.Kill()
	//g.gdbProcess = nil
	//g.gdbStdin = nil
	//g.gdbStdout = nil
	//g.gdbStderr = nil
	//// 清除WorkPath
	//os.RemoveAll(g.workPath)
	//g.workPath = ""
	//g.ptyMaster.Close()
	//g.ptySlave.Close()
	return nil
}

func (g *gdbDebugger) maskPath(message string) string {
	if message == "" {
		return ""
	}
	if filepath.IsAbs(g.workPath) && filepath.IsAbs("./") {
		relativePath := "." + string(filepath.Separator)
		absolutePath := filepath.Join(g.workPath, relativePath)
		message = strings.Replace(message, relativePath, absolutePath, -1)
	}
	repl := ""
	if g.workPath[len(g.workPath)-1] == '/' {
		repl = "/"
	}
	pattern := regexp.QuoteMeta(g.workPath)
	re := regexp.MustCompile(pattern)
	message = re.ReplaceAllString(message, repl)
	return message
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

func (d *gdbDebugger) configGDB(workPath string) error {
	fileStr := "set print elements 0\n" +
		"set print null-stop on\n" +
		"set print repeats 0\n" +
		"set print union on\n" +
		"set width 0\n"
	return os.WriteFile(path.Join(workPath, GdbInitFileName), []byte(fileStr), 0644)
}

type addBreakpointReq struct {
	BP Breakpoint
}

type removeBreakpointReq struct {
	BP Breakpoint
}
