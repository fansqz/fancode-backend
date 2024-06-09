package gdb_debugger

import (
	"FanCode/constants"
	"FanCode/service/debug/debugger"
	"FanCode/service/judger"
	"errors"
	"fmt"
	"github.com/fansqz/GoDebugger/gdb"
	"log"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CompileLimitTime = int64(10 * time.Second)
	OptionTimeout    = time.Second * 2
)

type gdbDebugger struct {
	workPath string // 工作目录
	judger   *judger.JudgeCore
	gdb      *gdb.Gdb

	callback debugger.NotificationCallback

	runningLock sync.RWMutex
	running     bool // 是否在运行中，只有运行中才可以执行一些单步调试动作

	// 保证map的线程安全
	lock sync.RWMutex
	// dap中没有断点编号，但是gdb却有，该映射是number:(file:line)的映射
	breakpointMap map[string]string
	// dap中没有断点编号，但是gdb却有，该映射是(file:line):number的映射
	breakpointInverseMap map[string]string

	// 由于为了防止stepIn操作会进入系统依赖内部的特殊处理
	preDeleteContinues     int
	preDeleteContinuesLock sync.Mutex
}

func NewGdbDebugger(gdbCallback debugger.NotificationCallback) debugger.Debugger {
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
	log.Println(m)
	typ := g.getStringFromMap(m, "type")
	switch typ {
	case "exec":
		class := g.getStringFromMap(m, "class")
		switch class {
		case "stopped":
			// 程序停止
			payload := g.getInterfaceFromMap(m, "payload")
			g.processStoppedData(payload)
			// 标记程序状态为停止状态
			g.runningLock.Lock()
			g.running = false
			g.runningLock.Unlock()
		case "running":
			// 程序执行
			g.preDeleteContinuesLock.Lock()
			if g.preDeleteContinues > 0 {
				g.preDeleteContinues--
				g.preDeleteContinuesLock.Unlock()
				return
			}
			g.preDeleteContinuesLock.Unlock()
			// 设置用户程序为执行状态
			g.runningLock.Lock()
			g.running = true
			g.runningLock.Unlock()

			g.callback(&debugger.ContinuedEvent{})

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
		frame := g.getInterfaceFromMap(m, "frame")
		// 检测是否还在workpath范围
		fullname := g.getStringFromMap(frame, "fullname")
		file = g.maskPath(fullname)
		lineStr := g.getStringFromMap(frame, "line")
		line, _ = strconv.Atoi(lineStr)
		// 返回停留的断点位置
		g.callback(&debugger.StoppedEvent{
			Reason: reason,
			File:   file,
			Line:   line,
		})
	} else if r == "end-stepping-range" || r == "function-finished" {
		reason = constants.StepStopped
		frame := g.getInterfaceFromMap(m, "frame")
		// 检测是否还在workpath范围
		fullname := g.getStringFromMap(frame, "fullname")
		if !strings.HasPrefix(fullname, g.workPath) {
			// 说明通过step进入了系统依赖内部
			err := g.StepOut()
			if err == nil {
				g.preDeleteContinuesLock.Lock()
				g.preDeleteContinues++
				g.preDeleteContinuesLock.Unlock()
				return
			}
			log.Println(err)
		}
		file = g.maskPath(fullname)
		lineStr := g.getStringFromMap(frame, "line")
		line, _ = strconv.Atoi(lineStr)
		// 返回停留的断点位置
		g.callback(&debugger.StoppedEvent{
			Reason: reason,
			File:   file,
			Line:   line,
		})
	} else if r == "exited-normally" {
		// 程序退出
		g.callback(&debugger.ExitedEvent{
			ExitCode: 0,
		})
	}

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
			g.callback(&debugger.CompileEvent{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		if !r.Compiled {
			g.callback(&debugger.CompileEvent{
				Success: false,
				Message: r.ErrorMessage,
			})
			return
		}
		g.callback(&debugger.CompileEvent{
			Success: true,
			Message: "用户代码编译成功",
		})
		// 创建命令
		m, _ := g.gdb.Send("file-exec-and-symbols", execFile)
		if result, ok := m["class"]; ok && result == "done" {
			g.callback(&debugger.LaunchEvent{
				Success: true,
				Message: "目标代码加载成功",
			})
		} else {
			g.callback(&debugger.LaunchEvent{
				Success: false,
				Message: "目标代码加载失败",
			})
			return
		}
	}()
	return nil
}

func (g *gdbDebugger) Start() error {
	var gdbCallback gdb.AsyncCallback = func(m map[string]interface{}) {
		// 启动协程读取用户输出
		go g.processUserOutput()
	}
	err := g.gdb.SendAsync(gdbCallback, "exec-run")
	return err
}

func (g *gdbDebugger) processUserOutput() {
	b := make([]byte, 1024)
	for {
		n, err := g.gdb.Read(b)
		if err != nil {
			log.Println(err)
			return
		}
		output := string(b[0:n])
		g.callback(&debugger.OutputEvent{
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

func (g *gdbDebugger) StepOver() error {
	g.runningLock.Lock()
	defer g.runningLock.Unlock()
	if g.running {
		return errors.New("程序运行中，无法执行单步调试")
	}
	err := g.gdb.SendAsync(func(obj map[string]interface{}) {}, "exec-next")
	return err
}

func (g *gdbDebugger) StepIn() error {
	g.runningLock.Lock()
	defer g.runningLock.Unlock()
	if g.running {
		return errors.New("程序运行中，无法执行单步调试")
	}
	err := g.gdb.SendAsync(func(obj map[string]interface{}) {}, "exec-step")
	return err
}

func (g *gdbDebugger) StepOut() error {
	g.runningLock.Lock()
	defer g.runningLock.Unlock()
	if g.running {
		return errors.New("程序运行中，无法执行单步调试")
	}
	err := g.gdb.SendAsync(func(obj map[string]interface{}) {}, "exec-finish")
	return err
}

func (g *gdbDebugger) Continue() error {
	g.runningLock.Lock()
	defer g.runningLock.Unlock()
	if g.running {
		return errors.New("程序运行中，无法执行continue调试")
	}
	err := g.gdb.SendAsync(func(obj map[string]interface{}) {}, "exec-continue")
	return err
}

// todo: 添加断点和删除断点串行执行
func (g *gdbDebugger) AddBreakpoints(breakpoints []*debugger.Breakpoint) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	var callback gdb.AsyncCallback = func(m map[string]interface{}) {
		// 处理响应
		bkpts, _ := g.getPayloadFromMap(m)
		breakpoint, err := g.getBreakpoint(bkpts)
		if err != nil {
			log.Println(err)
		} else {
			file := g.getStringFromMap(breakpoint, "file")
			lineStr := g.getStringFromMap(breakpoint, "line")
			line, _ := strconv.Atoi(lineStr)
			g.callback(&debugger.BreakpointEvent{
				Reason: constants.NewType,
				Breakpoints: []*debugger.Breakpoint{
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
	for _, bp := range breakpoints {
		err := g.gdb.SendAsync(callback, "break-insert", path.Join(g.workPath, bp.File)+":"+strconv.Itoa(bp.Line))
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

func (g *gdbDebugger) RemoveBreakpoints(breakpoints []*debugger.Breakpoint) error {
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
				g.callback(&debugger.BreakpointEvent{
					Reason:      constants.RemovedType,
					Breakpoints: []*debugger.Breakpoint{{file, bp.Line}},
				})
			}
			log.Println(m)
		}
		if err := g.gdb.SendAsync(callback, "break-delete", number); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (g *gdbDebugger) GetStackTrace() ([]*debugger.StackFrame, error) {
	g.runningLock.Lock()
	defer g.runningLock.Unlock()
	if g.running {
		return nil, errors.New("程序未暂停无法获取栈帧信息")
	}
	m, err := g.sendWithTimeOut(OptionTimeout, "stack-list-frames")
	if err != nil {
		return nil, err
	}
	answer := make([]*debugger.StackFrame, 0, 5)
	stackMap, success := g.getPayloadFromMap(m)
	if !success {
		return nil, errors.New("操作失败，无数据")
	}
	stackList := g.getListFromMap(stackMap, "stack")
	for _, s := range stackList {
		frame := g.getInterfaceFromMap(s, "frame")
		id := g.getStringFromMap(frame, "level")
		fun := g.getStringFromMap(frame, "func")
		line := g.getIntFromMap(frame, "line")
		fullname := g.getStringFromMap(frame, "fullname")
		stack := &debugger.StackFrame{
			Id:   id,
			Name: fun,
			Line: line,
			Path: g.maskPath(fullname),
		}
		answer = append(answer, stack)
	}
	return answer, nil
}

func (g *gdbDebugger) GetFrameVariables(frameId string) ([]*debugger.Variable, error) {
	g.runningLock.Lock()
	defer g.runningLock.Unlock()
	// 获取当前线程id
	m, err := g.sendWithTimeOut(OptionTimeout, "thread-info")
	if err != nil {
		return nil, err
	}
	threadMap, success := g.getPayloadFromMap(m)
	if !success {
		return nil, errors.New("获取线程id失败")
	}
	currentThreadId := g.getStringFromMap(threadMap, "current-thread-id")
	// 获取栈帧信息
	m, err = g.sendWithTimeOut(OptionTimeout, "stack-list-variables",
		"--thread", currentThreadId, "--frame", frameId, "2")
	if err != nil {
		return nil, err
	}
	payload, success := g.getPayloadFromMap(m)
	if !success {
		return nil, errors.New("操作失败")
	}
	variables := g.getListFromMap(payload, "variables")
	return g.convertVariableMapToVariableStruct(variables, false, frameId), nil
}

func (g *gdbDebugger) GetVariables(reference string) ([]*debugger.Variable, error) {
	// 正则表达式，捕获栈帧ID和变量名
	t := strings.Split(reference, "-")
	// 在字符串中匹配正则表达式
	if len(t) < 3 {
		fmt.Println("No matches found")
		return nil, errors.New("引用格式有问题")
	}
	// 从匹配结果中获取捕获组
	frameId := t[1]
	structName := t[2]
	if t[0] == "p" {
		structName = "*" + structName
	}
	// 切换栈帧
	_, err := g.sendWithTimeOut(OptionTimeout, "stack-select-frame", frameId)
	if err != nil {
		return nil, err
	}
	// 创建变量
	m, err := g.sendWithTimeOut(OptionTimeout, "var-create", "structName", "@", structName)
	defer func() {
		_, _ = g.sendWithTimeOut(OptionTimeout, "var-delete", "structName")
	}()
	if t[0] == "p" {
		// 如果是指针类型，可以直接查看到值
		payload, success := g.getPayloadFromMap(m)
		if !success {
			return nil, errors.New("系统错误")
		}
		g.mapSet(payload, "name", structName)

		// 查看其children数量，如果大于0，说明是结构体或者数组之类的，将其value设置为空
		m2, err2 := g.sendWithTimeOut(OptionTimeout, "var-info-num-children", "structName")
		if err2 != nil {
			return nil, err2
		}
		payload2, success2 := g.getPayloadFromMap(m2)
		if !success2 {
			return nil, errors.New("系统错误")
		}
		n := g.getIntFromMap(payload2, "numchild")

		if n > 0 {
			g.mapDelete(payload, "value")
		}
		return g.convertVariableMapToVariableStruct([]interface{}{payload}, false, frameId), nil
	}

	if t[0] == "v" {
		// 读取栈帧变量信息
		m, err = g.sendWithTimeOut(OptionTimeout, "var-list-children", "2", "structName")
		if err != nil {
			return nil, err
		}
		payload, success := g.getPayloadFromMap(m)
		if !success {
			return nil, errors.New("系统错误")
		}
		children := g.getListFromMap(payload, "children")
		return g.convertVariableMapToVariableStruct(children, true, frameId), nil
	}
	return nil, errors.New("格式不支持")
}

func (g *gdbDebugger) convertVariableMapToVariableStruct(variables []interface{}, isChildren bool, frameId string) []*debugger.Variable {
	answer := make([]*debugger.Variable, 0, 10)
	for _, v := range variables {
		if isChildren {
			v = g.getInterfaceFromMap(v, "child")
		}
		variable := &debugger.Variable{
			Name:  g.getStringFromMap(v, "name"),
			Value: g.getStringFromMap(v, "value"),
			Type:  g.getStringFromMap(v, "type"),
		}
		valueExist := g.checkKeyFromMap(v, "value")
		if !valueExist {
			// 结构体类型
			variable.Reference = fmt.Sprintf("v-%s-%s", frameId, variable.Name)
		} else {
			// 判断类型是否是指针类型
			t := strings.Split(variable.Type, " ")
			if len(t) > 1 {
				variable.Reference = fmt.Sprintf("p-%s-%s", frameId, variable.Name)
			}
		}
		answer = append(answer, variable)
	}
	return answer
}

func (g *gdbDebugger) Terminate() error {
	// 发送终端给程序
	err := g.gdb.Interrupt()
	if err != nil {
		log.Println(err)
	}
	_ = g.gdb.Exit()
	// 保证map的线程安全
	g.lock.Lock()
	defer g.lock.Unlock()
	g.breakpointMap = make(map[string]string, 10)
	g.breakpointInverseMap = make(map[string]string, 10)
	g.preDeleteContinues = 0
	g.callback(&debugger.TerminatedEvent{})
	return nil
}
