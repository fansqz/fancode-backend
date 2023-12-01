package debug

import (
	"FanCode/constants"
	"FanCode/service/judger"
	"context"
	"fmt"
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

func (d *debugHandlerC) StartDebug(execFile string, workPath string, options *DebugOptions) (*DebugResult, error) {
	if workPath[len(workPath)-1:] != "/" {
		workPath = workPath + "/"
	}
	d.workPath = workPath
	if options != nil && len(options.Breakpoints) != 0 {
		d.breakpoints = options.Breakpoints
	}
	var limitTime int64 = 0
	if options != nil && options.LimitTime > 0 {
		limitTime = options.LimitTime
	}
	if err := d.configGDB(workPath); err != nil {
		return nil, err
	}
	d.process = exec.Command("gdb", "-x", path.Join(workPath, "init.gdb"), execFile)
	d.stdin, _ = d.process.StdinPipe()
	d.stdout, _ = d.process.StdoutPipe()
	d.stderr, _ = d.process.StderrPipe()

	// 执行启动
	if err := d.process.Start(); err != nil {
		return nil, err
	}

	debugResult := &DebugResult{}

	// 读取输出
	var ctx context.Context
	if limitTime != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(limitTime))
		defer cancel()
	} else {
		ctx = context.Background()
	}
	_, timeout := d.readStdout(ctx)
	if timeout {
		// 超时，杀死进程
		if err := d.process.Process.Kill(); err != nil {
			return nil, err
		}
		debugResult.IsEnd = true
		return debugResult, nil
	}

	// 设置一些断点
	d.stdin.Write([]byte("b scanf\n"))
	d.readStdout(ctx)
	d.stdin.Write([]byte("b getchar\n"))
	d.readStdout(ctx)
	for _, bp := range d.breakpoints {
		_, _ = d.stdin.Write([]byte("b " + bp.File + ":" + strconv.Itoa(bp.Line) + "\n"))
		d.readStdout(ctx)
	}

	// 开始运行
	d.stdin.Write([]byte("r\n"))
	output, timeout := d.readStdout(ctx)

	debugResult.UserOutput, debugResult.IsEnd = d.parseGdbOutput(output, "r")
	if timeout {
		d.process.Process.Kill()
		debugResult.IsEnd = true
	}

	// 读取堆栈信息
	debugResult.BackTrace = d.getBackTrace()

	// 读取行号等信息
	debugResult.File = debugResult.BackTrace[0].File
	debugResult.Function = debugResult.BackTrace[0].Function
	debugResult.Line = debugResult.BackTrace[0].Line
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
	result, err := fmt.Sscanf(s, "%d")
	if err != nil {
		return 0
	}
	return result
}

func (d *debugHandlerC) Next(num int, options *DebugOptions) (*DebugResult, error) {

	return nil, nil
}

func (d *debugHandlerC) Step(num int, options *DebugOptions) (*DebugResult, error) {
	return nil, nil
}

func (d *debugHandlerC) Continue(num int, options *DebugOptions) (*DebugResult, error) {
	return nil, nil
}

func (d *debugHandlerC) AddBreakpoints(breakpoints []Breakpoint) error {
	d.breakpoints = d.removeFromBreakpointSlice(d.breakpoints, breakpoints)
	for _, bp := range breakpoints {
		d.stdin.Write([]byte("b " + bp.File + ":" + strconv.Itoa(bp.Line) + "\n"))
		d.readStdout(context.Background())
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
		d.readStdout(context.Background())
	}
	return nil
}

// 读取read中的输出，返回输出内容和是否超时
func (d *debugHandlerC) readStdout(ctx context.Context) (string, bool) {
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
	result, _ := d.readStdout(context.Background())
	return result
}

// parseGdbOutput 解析gdb的输出，获取其中的用户输出，
// 返回用户的控制台输出，以及gdb是否终止
func (d *debugHandlerC) parseGdbOutput(cmd string, gdbOutput string) (string, bool) {
	isEnd := false
	// 如果continue，去除头部必含有的文本
	if cmd == "c" {
		gdbOutput = strings.Replace(gdbOutput, "Continuing.\n", "", 1)
	}

	// 解析末尾的 结束/代码行
	// 判断是否调试结束，并从文本中删除结束表示
	re := regexp.MustCompile(`(.*)(\[Inferior \d+ \(process \d+\) exited normally\])`)
	matches := re.FindStringSubmatch(gdbOutput)

	if len(matches) < 3 {
		lastSecondNewlineIndex := strings.LastIndex(gdbOutput[0:len(gdbOutput)-1], "\n")
		gdbOutput = gdbOutput[0 : lastSecondNewlineIndex+1]
	} else {
		isEnd = true
		gdbOutput = matches[1]
	}

	// 解析去除断点信息
	re = regexp.MustCompile(`(.*)(\nBreakpoint \d+, .+ at .+:\d+\n)`)
	matches = re.FindStringSubmatch(gdbOutput)
	if len(matches) >= 3 {
		gdbOutput = matches[1]
	}
	return gdbOutput, isEnd
}
