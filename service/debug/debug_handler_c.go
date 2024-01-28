package debug

import (
	"FanCode/constants"
	"FanCode/service/judger"
	"context"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	GdbInitFileName = "init.gdb"
)

type debugHandlerC struct {
	process     *exec.Cmd    // 调试进程
	breakpoints []Breakpoint // 断点
	workPath    string       // 工作目录
	debugFile   string       // 调试关注的文件
	judger      *judger.JudgeCore
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	stdoutCh    chan []byte
}

func NewDebugHandlerC() DebugHandler {
	return &debugHandlerC{
		judger: judger.NewJudgeCore(),
	}
}

func (d *debugHandlerC) Compile(compileFiles []string, outFilePath string, options *CompileOptions) (*CompileResult, error) {
	if options == nil {
		options = &CompileOptions{}
	}
	r, err := d.judger.Compile(compileFiles, outFilePath, &judger.CompileOptions{
		Language:        constants.ProgramC,
		LimitTime:       options.LimitTime,
		ExcludedPaths:   options.ExcludedPaths,
		ReplacementPath: options.ReplacementPath,
	})
	return &CompileResult{
		Compiled:         r.Compiled,
		ErrorMessage:     r.ErrorMessage,
		CompiledFilePath: r.CompiledFilePath,
	}, err
}

func (d *debugHandlerC) Reset() error {
	d.breakpoints = nil
	if d.process != nil {
		return d.process.Process.Kill()
	}
	return nil
}

func (d *debugHandlerC) Start(execFile string, options *StartDebugOptions) (*DebugResult, error) {
	if options != nil && options.WorkPath != "" {
		if options.WorkPath[len(options.WorkPath)-1:] != "/" {
			options.WorkPath = options.WorkPath + "/"
		}
		d.workPath = options.WorkPath
	}
	if options != nil && options.DebugFile != "" {
		d.debugFile = options.DebugFile
	}
	if options != nil && len(options.Breakpoints) != 0 {
		d.breakpoints = options.Breakpoints
	}

	if err := d.configGDB(d.workPath); err != nil {
		return nil, err
	}
	d.process = exec.Command("gdb", "-x", path.Join(d.workPath, "init.gdb"), execFile)
	d.stdin, _ = d.process.StdinPipe()
	d.stdout, _ = d.process.StdoutPipe()
	d.stderr, _ = d.process.StderrPipe()
	d.stdoutCh = make(chan []byte, 10)
	// 执行启动
	if err := d.process.Start(); err != nil {
		return nil, err
	}

	// 启动协程读取输出数据
	go d.readStdoutFunc()

	debugResult := &DebugResult{}
	// 读取输出
	if out, _ := d.readStdout(); out == "" {
		// 无数据杀死进程
		if err := d.process.Process.Kill(); err != nil {
			return nil, err
		}
		debugResult.IsEnd = true
		return debugResult, nil
	}

	// 取消gdb缓存
	d.stdin.Write([]byte("call setbuf(stdout, NULL)\n"))
	d.readStdout()

	// 设置断点
	for _, bp := range d.breakpoints {
		_, _ = d.stdin.Write([]byte("b " + bp.File + ":" + strconv.Itoa(bp.Line) + "\n"))
		d.readStdout()
	}

	// 开始运行
	d.stdin.Write([]byte("r\n"))
	d.flushAndReadStdoutForDebugResult(debugResult)

	// 读取堆栈信息
	debugResult.BackTrace = d.getBackTrace()

	// 读取行号等信息
	d.setRunPosition(debugResult, debugResult.BackTrace)

	return debugResult, nil
}

func (d *debugHandlerC) readStdoutFunc() {
	for {
		b := make([]byte, 10*1024)
		n, err := d.stdout.Read(b)
		if err != nil {
			break
		}
		d.stdoutCh <- b[0:n]
	}
}
func (d *debugHandlerC) Restart() (*DebugResult, error) {

	// 开始运行
	d.stdin.Write([]byte("r\n"))
	d.readStdout()
	d.stdin.Write([]byte("y\n"))

	debugResult := &DebugResult{}
	d.flushAndReadStdoutForDebugResult(debugResult)
	if debugResult.IsEnd {
		return debugResult, nil
	}

	// 读取堆栈信息
	debugResult.BackTrace = d.getBackTrace()

	// 读取行号等信息
	d.setRunPosition(debugResult, debugResult.BackTrace)

	return debugResult, nil
}

func (d *debugHandlerC) configGDB(workPath string) error {
	fileStr := "set print elements 0\n" +
		"set print null-stop on\n" +
		"set print repeats 0\n" +
		"set print union on\n" +
		"set width 0\n"
	return os.WriteFile(path.Join(workPath, GdbInitFileName), []byte(fileStr), 0644)
}

// getBackTrace 读取函数调用栈信息
func (d *debugHandlerC) getBackTrace() BackTrace {

	d.stdin.Write([]byte("bt\n"))
	// 读取调用栈信息
	staceData, _ := d.readStdout()
	stace := strings.Split(staceData, "#")
	stace = stace[1:]
	re := regexp.MustCompile(`\d+\s+(0x[0-9a-f]+)?\s?(\w+)?\s?\((.*?)\) at (\S+):(\d+)`)

	frameList := make([]StackFrame, len(stace))
	// 解析调用栈信息
	for i, frameStr := range stace {
		match := re.FindStringSubmatch(frameStr)
		if match != nil {
			function := match[2]
			args := match[3]
			file := match[4]
			line := match[5]

			// 屏蔽路径
			file = strings.Replace(function, d.workPath, "/", 1)

			stackFrame := StackFrame{
				Function: function,
				Args:     args,
				File:     file,
				Line:     parseInt(line),
			}

			frameList[i] = stackFrame
		}
	}
	return frameList
}

func parseInt(s string) int {
	result, _ := strconv.Atoi(s)
	return result
}

func (d *debugHandlerC) Next(num int) (*DebugResult, error) {
	return d.next("n", num)
}

func (d *debugHandlerC) Step(num int) (*DebugResult, error) {
	return d.next("s", num)
}

func (d *debugHandlerC) Continue(num int) (*DebugResult, error) {
	return d.next("c", num)
}

func (d *debugHandlerC) next(cmd string, num int) (*DebugResult, error) {
	d.stdin.Write([]byte(cmd + " " + strconv.Itoa(num) + "\n"))

	debugResult := &DebugResult{}

	// 解析输出
	d.flushAndReadStdoutForDebugResult(debugResult)
	if debugResult.IsEnd {
		return debugResult, nil
	}

	// 读取堆栈信息
	debugResult.BackTrace = d.getBackTrace()

	// 读取行号等信息
	d.setRunPosition(debugResult, debugResult.BackTrace)

	return debugResult, nil
}

func (d *debugHandlerC) AddBreakpoints(breakpoints []Breakpoint) error {
	d.breakpoints = d.removeFromBreakpointSlice(d.breakpoints, breakpoints)
	for _, bp := range breakpoints {
		d.stdin.Write([]byte("b " + bp.File + ":" + strconv.Itoa(bp.Line) + "\n"))
		d.readStdout()
	}
	return nil
}

// removeFromBreakpointSlice source中删除toRemove断点
func (d *debugHandlerC) removeFromBreakpointSlice(source, toRemove []Breakpoint) []Breakpoint {
	removeSet := make(map[string]struct{})

	// 将待删除的元素添加到Set中
	for _, v := range toRemove {
		removeSet[v.toString()] = struct{}{}
	}

	result := make([]Breakpoint, 0, len(source))

	// 将不在Set中的元素添加到结果切片中
	for _, s := range source {
		if _, exists := removeSet[s.toString()]; !exists {
			result = append(result, s)
		}
	}

	return result
}

func (d *debugHandlerC) RemoveBreakpoints(breakpoints []Breakpoint) error {
	for _, bp := range breakpoints {
		d.stdin.Write([]byte("clear " + bp.File + ":" + strconv.Itoa(bp.Line) + "\n"))
		d.readStdout()
	}
	return nil
}

// redirect 判断是否执行到非关注文件下，如果执行到非关注文件下，返回到关注文件下
func (d *debugHandlerC) redirect(ctx context.Context, debugResult *DebugResult) {
	backTrace := d.getBackTrace()
	if len(backTrace) == 0 {
		return
	}
	file := backTrace[0].File
	if file == d.debugFile {
		return
	}

	index := -1
	for i, frame := range backTrace {
		if frame.File == d.debugFile {
			index = i
		}
	}

	// 如果调用栈中没有目标文件，程序直接执行到下一个断点，并退出
	if index == -1 {
		d.stdin.Write([]byte("c\n"))
		d.flushAndReadStdoutForDebugResult(debugResult)
		return
	}

	// 跳出函数，直到回到关注文件
	for i := 0; i < index; i++ {
		d.stdin.Write([]byte("f\n"))
		d.flushAndReadStdoutForDebugResult(debugResult)
	}
}

// flushAndReadStdoutForDebugResult
// 刷新控制台的用户输出，并读取数据添加到debugResult中
// 会读取两次控制台，第一次读取gdb输出，第二次读取用户输出
func (d *debugHandlerC) flushAndReadStdoutForDebugResult(debugResult *DebugResult) {
	gdbout, _ := d.readStdout()
	userOutput, gdbEnd := d.flushAndReadStdout()
	debugResult.IsEnd = d.parseGdbOutput(gdbout)
	debugResult.UserOutput += userOutput
	debugResult.IsEnd = gdbEnd
}

// flushAndReadStdout
// 刷新控制台的用户输出，并读取
// 返回1.gdb输出
func (d *debugHandlerC) flushAndReadStdout() (string, bool) {
	d.stdin.Write([]byte("call (void)fflush(0)\n"))
	return d.readStdout()
}

// readStdout 读取gdb输出
// answer: 输出信息
// gdbEnd: 是否含有(gdb)标识
func (d *debugHandlerC) readStdout() (answer string, gdbEnd bool) {
	// 超时时间设置为0.02秒
	timeout := time.Microsecond * 200
	select {
	case <-time.After(timeout):
		return "", false
	case output := <-d.stdoutCh:
		// 成功获取输出
		answer = string(output)
		answer = strings.TrimRight(answer, " ")
		if strings.HasSuffix(answer, "(gdb)") {
			answer = answer[:len(answer)-5]
			gdbEnd = true
		}
		return
	}
}

// parseGdbOutput 解析gdb输出，返回用户程序是否结束
func (d *debugHandlerC) parseGdbOutput(gdbOutput string) bool {
	re := regexp.MustCompile(`(.*)(\[Inferior \d+ \(process \d+\) exited normally\])`)
	matches := re.FindStringSubmatch(gdbOutput)

	return len(matches) != 0
}

func (d *debugHandlerC) setRunPosition(debugResult *DebugResult, backTrace BackTrace) {
	if len(backTrace) == 0 {
		return
	}
	debugResult.File = backTrace[0].File
	debugResult.Function = backTrace[0].Function
	debugResult.Line = backTrace[0].Line
}

func (d *debugHandlerC) getTimeoutContext(limitTime int64) (context.Context, context.CancelFunc) {
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
