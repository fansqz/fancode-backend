package debugger

import (
	"FanCode/constants"
	"FanCode/service/judger"
	"errors"
	"fmt"
	"github.com/fansqz/GoDebugger/gdb"
	"io"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CompileLimitTime = int64(10 * time.Second)
)

type gdbDebugger struct {
	workPath string // 工作目录
	judger   *judger.JudgeCore
	gdb      *gdb.Gdb

	callback NotificationCallback

	// 保证map的线程安全
	lock sync.RWMutex
	// 断点映射
	// dap中没有断点编号，但是gdb却有，该映射是number:(file:line)的映射
	breakpointMap map[string]string
	// dap中没有断点编号，但是gdb却有，该映射是(file:line):number的映射
	breakpointInverseMap map[string]string
}

func NewGdbDebugger(gdbCallback NotificationCallback) Debugger {
	d := &gdbDebugger{
		judger:               judger.NewJudgeCore(),
		callback:             gdbCallback,
		breakpointMap:        make(map[string]string, 10),
		breakpointInverseMap: make(map[string]string, 10),
	}
	// 创建gdb对象
	gd, err := gdb.New(d.gdbNotificationCallback)
	if err != nil {
		log.Println(err)
	}
	d.gdb = gd
	return d
}

// gdbNotificationCallback 处理gdb异步响应的回调
func (g *gdbDebugger) gdbNotificationCallback(m map[string]interface{}) {
	fmt.Println(m)
	typ := g.getStringFromMap(m, "type")
	switch typ {
	case "exec":
		class := g.getStringFromMap(m, "class")
		switch class {
		case "stopped":
			// 程序停止
			payload := g.getInterfaceFromMap(m, "payload")
			g.processStoppedData(payload)
		case "running":
			// 程序执行
			g.callback(&ContinuedEvent{})
		}
	}

}

func (g *gdbDebugger) processStoppedData(m interface{}) {
	var file string
	var line int
	var reason constants.StoppedReasonType
	r := g.getStringFromMap(m, "reason")
	if r == "breakpoint-hit" {
		reason = constants.BreakpointStopped
		breakpointNum := g.getStringFromMap(m, "bkptno")
		// 读取断点
		g.lock.RLock()
		var bk string
		bk, _ = g.breakpointMap[breakpointNum]
		g.lock.RUnlock()
		file, line = g.parseBreakpoint(bk)
		// 返回停留的断点位置
		g.callback(&StoppedEvent{
			Reason: reason,
			File:   file,
			Line:   line,
		})
	} else if r == "end-stepping-range" {
		reason = constants.StepStopped
		frame := g.getInterfaceFromMap(m, "frame")
		file = g.maskPath(g.getStringFromMap(frame, "file"))
		lineStr := g.getStringFromMap(frame, "line")
		line, _ = strconv.Atoi(lineStr)
		// 返回停留的断点位置
		g.callback(&StoppedEvent{
			Reason: reason,
			File:   file,
			Line:   line,
		})
	} else if r == "exited-normally" {
		// 程序退出
		g.callback(&ExitedEvent{
			ExitCode: 0,
		})
	}

}

func (g *gdbDebugger) getStringFromMap(m interface{}, key string) string {
	s, _ := m.(map[string]interface{})[key].(string)
	return s
}

func (g *gdbDebugger) getIntFromMap(m interface{}, key string) int {
	s, _ := m.(map[string]interface{})[key].(string)
	num, _ := strconv.Atoi(s)
	return num
}

func (g *gdbDebugger) getInterfaceFromMap(m interface{}, key string) interface{} {
	s, _ := m.(map[string]interface{})[key]
	return s
}

// parseBreakpoint 解析断点
func (g *gdbDebugger) parseBreakpoint(bk string) (string, int) {
	l := strings.Split(bk, ":")
	if len(l) != 2 {
		return "", 0
	}
	file := l[0]
	lineStr := l[1]
	line, _ := strconv.Atoi(lineStr)
	return file, line
}

func (g *gdbDebugger) Launch(compileFiles []string, workPath string, language constants.LanguageType) error {
	// 设置gdb文件
	g.workPath = workPath
	// 启动协程编译运行
	go func() {
		execFile := path.Join(workPath, "main")
		// 进行编译
		r, err := g.judger.Compile(compileFiles, execFile, &judger.CompileOptions{
			Language:        language,
			LimitTime:       CompileLimitTime,
			ExcludedPaths:   []string{workPath},
			ReplacementPath: "/",
		})
		if err != nil {
			g.callback(&CompileEvent{
				Success: false,
				Message: err.Error(),
			})
		}
		if !r.Compiled {
			g.callback(&CompileEvent{
				Success: false,
				Message: r.ErrorMessage,
			})
		}
		g.callback(&CompileEvent{
			Success: true,
			Message: "用户代码编译成功",
		})
		// 创建命令
		m, _ := g.gdb.Send("file-exec-and-symbols", execFile)
		if result, ok := m["class"]; ok && result == "done" {
			g.callback(&LaunchEvent{
				Success: true,
				Message: "目标代码加载成功",
			})
		} else {
			g.callback(&LaunchEvent{
				Success: false,
				Message: "目标代码加载失败",
			})
		}
	}()
	return nil
}

func (g *gdbDebugger) Start() error {
	var gdbCallback gdb.AsyncCallback = func(m map[string]interface{}) {
		// 启动协程读取用户输出
		go g.processUserOutput()
	}
	err := g.gdb.SendAsync("exec-run", gdbCallback)
	return err
}

func (g *gdbDebugger) processUserOutput() {
	b := make([]byte, 1024)
	for {
		n, err := g.gdb.Read(b)
		if err != nil {
			if err != io.EOF {
				// 如果不是EOF，打印出错误信息
				log.Printf("读取数据时发生错误: %v", err)
			} else {
				return
			}
		}
		output := string(b[0:n])
		g.callback(&OutputEvent{
			Output: output,
		})
	}
}

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
	err := g.gdb.SendAsync("exec-next", func(obj map[string]interface{}) {})
	return err
}

func (g *gdbDebugger) Step() error {
	err := g.gdb.SendAsync("exec-step", func(obj map[string]interface{}) {})
	return err
}

func (g *gdbDebugger) Continue() error {
	err := g.gdb.SendAsync("exec-continue", func(obj map[string]interface{}) {})
	return err
}

// todo: 添加断点和删除断点串行执行
func (g *gdbDebugger) AddBreakpoints(breakpoints []*Breakpoint) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	var callback gdb.AsyncCallback = func(m map[string]interface{}) {
		// 处理响应
		if result, ok := m["class"]; ok && result == "done" {
			if bkpts, ok2 := m["payload"]; ok2 {
				breakpoint, err := g.getBreakpoint(bkpts)
				if err != nil {
					log.Println(err)
				} else {
					file := g.getStringFromMap(breakpoint, "file")
					lineStr := g.getStringFromMap(breakpoint, "line")
					line, _ := strconv.Atoi(lineStr)
					g.callback(&BreakpointEvent{
						Reason: constants.NewType,
						Breakpoints: []*Breakpoint{
							{
								File: g.maskPath(file),
								Line: line,
							},
						},
					})
					// 设置map
					number := g.getStringFromMap(breakpoint, "number")
					g.lock.Lock()
					g.breakpointMap[number] = g.maskPath(file) + ":" + lineStr
					g.breakpointInverseMap[g.maskPath(file)+":"+lineStr] = number
					g.lock.Unlock()
				}
			}

		}
	}
	for _, bp := range breakpoints {
		err := g.gdb.SendAsync("break-insert", callback, path.Join(g.workPath, bp.File)+":"+strconv.Itoa(bp.Line))
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (g *gdbDebugger) getBreakpoint(data interface{}) (map[string]interface{}, error) {
	bkpt := data.(map[string]interface{})
	if bkpt2, ok := bkpt["bkpt"]; ok {
		if breakpoint, ok2 := bkpt2.(map[string]interface{}); ok2 {
			return breakpoint, nil
		}
	}
	return nil, errors.New("解析失败")
}

func (g *gdbDebugger) RemoveBreakpoints(breakpoints []*Breakpoint) error {
	for _, bp := range breakpoints {
		file := bp.File
		line := strconv.Itoa(bp.Line)
		if file[0] != '/' {
			file = "/" + file
		}
		g.lock.Lock()
		number := g.breakpointInverseMap[file+":"+line]
		g.lock.Unlock()
		var callback gdb.AsyncCallback = func(m map[string]interface{}) {
			// 断点删除成功，则移除map中的断点记录，并响应结果
			if class := g.getStringFromMap(m, "class"); class == "done" {
				g.lock.Lock()
				number = g.breakpointInverseMap[file+":"+line]
				delete(g.breakpointMap, number)
				delete(g.breakpointInverseMap, file+":"+line)
				g.callback(&BreakpointEvent{
					Reason:      constants.RemovedType,
					Breakpoints: []*Breakpoint{{file, bp.Line}},
				})
			}
			log.Println(m)
		}
		if err := g.gdb.SendAsync("break-delete", callback, number); err != nil {
			log.Println(err)
		}
	}
	return nil
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
