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

	var limitTime int64 = 0
	if options != nil && options.LimitTime > 0 {
		limitTime = options.LimitTime
	}
	if err := d.configGDB(d.workPath); err != nil {
		return nil, err
	}
	d.process = exec.Command("gdb", "-x", path.Join(d.workPath, "init.gdb"), execFile)
	d.stdin, _ = d.process.StdinPipe()
	d.stdout, _ = d.process.StdoutPipe()
	d.stderr, _ = d.process.StderrPipe()

	// 执行启动
	if err := d.process.Start(); err != nil {
		return nil, err
	}

	debugResult := &DebugResult{}

	// 读取输出
	ctx, cancel := d.getTimeoutContext(limitTime)
	defer cancel()
	_, timeout := d.readStdout1(ctx)
	if timeout {
		// 超时，杀死进程
		if err := d.process.Process.Kill(); err != nil {
			return nil, err
		}
		debugResult.IsEnd = true
		return debugResult, nil
	}

	// 取消gdb缓存
	d.stdin.Write([]byte("call setbuf(stdout, NULL)\n"))
	d.readStdout1(ctx)
	// 设置一些断点
	d.stdin.Write([]byte("b scanf\n"))
	d.readStdout1(ctx)
	d.stdin.Write([]byte("b getchar\n"))
	d.readStdout1(ctx)

	for _, bp := range d.breakpoints {
		_, _ = d.stdin.Write([]byte("b " + bp.File + ":" + strconv.Itoa(bp.Line) + "\n"))
		d.readStdout1(ctx)
	}

	// 开始运行
	d.stdin.Write([]byte("r\n"))
	var gdbout string
	gdbout, debugResult.UserOutput, timeout = d.flushAndReadStdout1(ctx)

	debugResult.IsEnd = d.parseGdbOutput(gdbout)
	if timeout {
		d.process.Process.Kill()
		debugResult.IsEnd = true
	}

	// 读取堆栈信息
	debugResult.BackTrace = d.getBackTrace()

	// 读取行号等信息
	d.setRunPosition(debugResult, debugResult.BackTrace)

	return debugResult, nil
}

func (d *debugHandlerC) Restart(options *DebugOptions) (*DebugResult, error) {
	var limitTime int64 = 0
	if options != nil && options.LimitTime > 0 {
		limitTime = options.LimitTime
	}

	debugResult := &DebugResult{}
	ctx, cancel := d.getTimeoutContext(limitTime)
	defer cancel()

	// 开始运行
	d.stdin.Write([]byte("r\n"))
	d.readStdout1(ctx)
	d.stdin.Write([]byte("y\n"))

	d.flushAndReadStdout2(ctx, debugResult)
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
	staceData := d.readStdout2()
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

func (d *debugHandlerC) Next(num int, options *DebugOptions) (*DebugResult, error) {
	return d.next("n", num, options)
}

func (d *debugHandlerC) Step(num int, options *DebugOptions) (*DebugResult, error) {
	return d.next("s", num, options)
}

func (d *debugHandlerC) Continue(num int, options *DebugOptions) (*DebugResult, error) {
	return d.next("c", num, options)
}

func (d *debugHandlerC) next(cmd string, num int, options *DebugOptions) (*DebugResult, error) {
	var limitTime int64 = 0
	if options != nil && options.LimitTime > 0 {
		limitTime = options.LimitTime
	}
	d.stdin.Write([]byte(cmd + " " + strconv.Itoa(num) + "\n"))

	// 读取输出
	ctx, cancel := d.getTimeoutContext(limitTime)
	defer cancel()
	debugResult := &DebugResult{}

	// 解析输出
	d.flushAndReadStdout2(ctx, debugResult)
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
		d.readStdout1(context.Background())
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
		d.readStdout1(context.Background())
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
		d.flushAndReadStdout2(ctx, debugResult)
		return
	}

	// 跳出函数，直到回到关注文件
	for i := 0; i < index; i++ {
		d.stdin.Write([]byte("f\n"))
		d.flushAndReadStdout2(ctx, debugResult)
	}
}

// flushAndReadStdout1
// 刷新控制台的用户输出，并读取
// 返回1.gdb输出 2. 用户输出 3. 是否超时
func (d *debugHandlerC) flushAndReadStdout1(ctx context.Context) (string, string, bool) {
	gdbOutput, timeout1 := d.readStdout1(ctx)
	d.stdin.Write([]byte("call (void)fflush(0)\n"))
	userOutput, timeout2 := d.readStdout1(ctx)
	return gdbOutput, userOutput, timeout1 || timeout2
}

// flushAndReadStdout2
// 刷新控制台的用户输出，并读取数据添加到debugResult中
func (d *debugHandlerC) flushAndReadStdout2(ctx context.Context, debugResult *DebugResult) {
	timeout := false
	gdbout := ""
	userOutput := ""
	gdbout, userOutput, timeout = d.flushAndReadStdout1(ctx)
	debugResult.IsEnd = d.parseGdbOutput(gdbout)
	debugResult.UserOutput += userOutput
	if timeout {
		d.process.Process.Kill()
		debugResult.IsEnd = true
	}
}

// readStdout1
// 读取read中的输出，返回输出内容和是否超时
// 1. gdb输出， 2. 是否超时
func (d *debugHandlerC) readStdout1(ctx context.Context) (string, bool) {
	answer := ""
	readBuffer := make([]byte, 1024*4)
	for {
		n, _ := d.stdout.Read(readBuffer)
		if n > 0 {
			answer += string(readBuffer[0:n])
		}
		answer = strings.TrimRight(answer, " ")
		if strings.HasSuffix(answer, "(gdb)") {
			break
		}
		// 不停有输出，或者一直没有输出
		deadline, ok := ctx.Deadline()
		if ok {
			timeout := time.Until(deadline)
			if timeout < 0 {
				return answer, true
			}
		}
	}
	// 移除(gdb)
	answer = answer[:len(answer)-5]
	return answer, false
}

// 读取read中的输出，不设置超时
func (d *debugHandlerC) readStdout2() string {
	result, _ := d.readStdout1(context.Background())
	return result
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
