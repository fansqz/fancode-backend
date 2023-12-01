package debug

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	GdbInitFileName = "init.gdb"
)

type debugHandlerC struct {
	process     *exec.Cmd // 调试进程
	stdin       io.Writer
	stdout      io.Reader
	stderr      io.Reader
	codeFile    string   // 用户文件，用于打断点
	breakpoints []string // 断点
}

func (d *debugHandlerC) Reset() error {
	d.stdin = nil
	d.stdout = nil
	d.stderr = nil
	d.codeFile = ""
	d.breakpoints = nil
	if d.process != nil {
		return d.process.Process.Kill()
	}
	return nil
}

func (d *debugHandlerC) StartDebug(execFile string, workPath string, options *DebugOptions) (*DebugResult, error) {
	if options != nil && len(options.Breakpoints) != 0 {
		d.breakpoints = options.Breakpoints
		d.codeFile = options.CodeFile
	}
	var limitTime int64 = 0
	if options != nil && options.LimitTime > 0 {
		limitTime = options.LimitTime
	}
	if err := d.configGDB(workPath); err != nil {
		return nil, err
	}

	d.process = exec.Command("gdb", "-x", path.Join(workPath, "init.gdb"), execFile)
	d.stdin = &bytes.Buffer{}
	d.stdout = &bytes.Buffer{}
	d.stderr = &bytes.Buffer{}
	d.process.Stdin = d.stdin.(*bytes.Buffer)
	d.process.Stdout = d.stdout.(*bytes.Buffer)
	d.process.Stderr = d.stderr.(*bytes.Buffer)

	// 执行启动
	if err := d.process.Start(); err != nil {
		return nil, err
	}

	debugResult := &DebugResult{}
	// 读取输出
	var ctx context.Context
	if limitTime == 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(limitTime))
		defer cancel()
	} else {
		ctx = context.Background()
	}
	out, timeout := d.read(d.stdout, ctx)
	if timeout {
		// 超时，杀死进程
		if err := d.process.Process.Kill(); err != nil {
			return nil, err
		}
		debugResult.IsEnd = true
		debugResult.Output = out
		return debugResult, nil
	}

	// 设置一些断点
	d.stdin.Write([]byte("b scanf\n"))
	d.read(d.stdout, ctx)
	d.stdin.Write([]byte("b getchar\n"))
	d.read(d.stdout, ctx)
	for _, bp := range d.breakpoints {
		_, _ = d.stdin.Write([]byte("b " + d.codeFile + ":" + bp))
		d.read(d.stdout, ctx)
	}

	// 开始运行
	d.stdin.Write([]byte("run\n"))
	debugResult.Output = "success"
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

// 读取read中的输出，返回输出内容和是否超时
func (d *debugHandlerC) read(read io.Reader, ctx context.Context) (string, bool) {
	answer := ""
	readBuffer := make([]byte, 1024*4)
	for {
		_, _ = read.Read(readBuffer)
		answer += string(readBuffer)
		if strings.HasSuffix(answer, "(gdb)") {
			break
		}
		// 不停有输出，或者一直没有输出都可以判
		deadline, ok := ctx.Deadline()
		if ok {
			timeout := time.Until(deadline)
			if timeout < 0 {
				return answer, true
			}
		}
	}
	return answer, false
}

func (d *debugHandlerC)

func (d *debugHandlerC) Next(num int) (*DebugResult, error) {

}

func (d *debugHandlerC) Step(num int) (*DebugResult, error) {

}

func (d *debugHandlerC) Continue(num int) (*DebugResult, error) {

}

func (d *debugHandlerC) AddBreakpoints(breakpoints []string) error {
	d.breakpoints = d.removeFromSlice(d.breakpoints, breakpoints)
	for _, bp := range breakpoints {
		d.stdin.Write([]byte("b " + d.codeFile + ":" + bp + "\n"))
		d.read(d.stdout, context.Background())
	}
	return nil
}

func (d *debugHandlerC) removeFromSlice(source, toRemove []string) []string {
	removeSet := make(map[string]struct{})

	// 将待删除的元素添加到Set中
	for _, v := range toRemove {
		removeSet[v] = struct{}{}
	}

	result := make([]string, 0, len(source))

	// 将不在Set中的元素添加到结果切片中
	for _, s := range source {
		if _, exists := removeSet[s]; !exists {
			result = append(result, s)
		}
	}

	return result
}

func (d *debugHandlerC) RemoveBreakpoints(breakpoints []string) error {
	for _, bp := range breakpoints {
		d.stdin.Write([]byte("clear " + d.codeFile + ":" + bp + "\n"))
		d.read(d.stdout, context.Background())
	}
	return nil
}
