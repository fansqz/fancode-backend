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
	currentThreadId, err := g.getCurrentThreadId()
	if err != nil {
		return nil, err
	}
	// 获取栈帧中所有局部变量
	m, err := g.sendWithTimeOut(OptionTimeout, "stack-list-variables",
		"--thread", currentThreadId, "--frame", frameId, "2")
	if err != nil {
		return nil, err
	}
	payload, success := g.getPayloadFromMap(m)
	if !success {
		return nil, errors.New("操作失败")
	}
	variables := g.getListFromMap(payload, "variables")
	return g.convertToVariableStruct4GetFrameVariables(frameId, variables), nil
}

func (g *gdbDebugger) GetVariables(reference string) ([]*debugger.Variable, error) {
	// 正则表达式，捕获栈帧ID和变量名
	refStruct, err := parseReference(reference)
	if err != nil {
		return nil, err
	}

	if refStruct.Type == "v" {
		// 如果是普通类型需要切换栈帧，同一个变量名，可能在不同栈帧中会有重复，需要定位栈帧和变量名称才能读取到变量值。
		if _, err = g.sendWithTimeOut(OptionTimeout, "stack-select-frame", refStruct.FrameId); err != nil {
			return nil, err
		}
	}

	// 获取所有children列表并解析
	var m map[string]interface{}

	name := "structName"
	// 创建变量
	if refStruct.Type == "v" {
		m, err = g.sendWithTimeOut(OptionTimeout, "var-create", name, "@",
			refStruct.VariableName)
	} else if refStruct.Type == "p" {
		m, err = g.sendWithTimeOut(OptionTimeout, "var-create", name, "*",
			fmt.Sprintf("(%s)%s", refStruct.PointType, refStruct.Address))
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = g.sendWithTimeOut(OptionTimeout, "var-delete", "structName")
	}()

	if refStruct.FieldPath == "" {
		m, err = g.sendWithTimeOut(OptionTimeout, "var-list-children", "1",
			name)
	} else {
		m, err = g.sendWithTimeOut(OptionTimeout, "var-list-children", "1",
			fmt.Sprintf("%s.%s", name, refStruct.FieldPath))
	}
	if err != nil {
		return nil, err
	}
	payload, success := g.getPayloadFromMap(m)
	if !success {
		return nil, errors.New("系统错误")
	}
	children := g.getListFromMap(payload, "children")
	return g.convertToVariableStruct4GetVariables(reference, children), nil
}

func (g *gdbDebugger) convertToVariableStruct4GetVariables(ref string, variables []interface{}) []*debugger.Variable {
	answer := make([]*debugger.Variable, 0, 10)
	for _, v := range variables {
		v = g.getInterfaceFromMap(v, "child")
		field := &debugger.Variable{
			Name:  g.convertVariableName(ref, g.getStringFromMap(v, "name")),
			Value: g.getStringFromMap(v, "value"),
			Type:  g.getStringFromMap(v, "type"),
		}

		// 如果value为空说明是结构体类型
		if !g.checkKeyFromMap(v, "value") {
			// 已经定位了的结构体下的某个属性，直接加路径即可。
			field.Reference = getFieldReference(ref, field.Name)
		}

		// 指针类型，如果有值，但是children又不为0说明是指针类型
		if g.checkKeyFromMap(v, "value") && g.getIntFromMap(v, "numchild") != 0 {
			if field.Type != "char *" {
				address := g.getAddress(field.Value)
				if !isNullPoint(address) {
					field.Reference = convertReference(
						&referenceStruct{Type: "p", PointType: field.Type, Address: address, VariableName: field.Name})
				}
			}
		}
		answer = append(answer, field)
	}
	return answer
}

// convertToVariableStruct4GetFrameVariables 将从gdb中获取到的变量列表转换成变量结构体
func (g *gdbDebugger) convertToVariableStruct4GetFrameVariables(frameId string, variables []interface{}) []*debugger.Variable {
	answer := make([]*debugger.Variable, 0, 10)
	for _, v := range variables {
		variable := &debugger.Variable{
			Name:  g.convertVariableName("", g.getStringFromMap(v, "name")),
			Value: g.getStringFromMap(v, "value"),
			Type:  g.getStringFromMap(v, "type"),
		}

		// 结构体类型，如果value为空说明是结构体类型
		if !g.checkKeyFromMap(v, "value") {
			// 如果parentRef不为空，说明是栈帧中的某个结构体变量
			variable.Reference = convertReference(&referenceStruct{Type: "v", FrameId: frameId, VariableName: variable.Name})
		}

		// 指针类型，如果有值，但是children又不为0说明是指针类型
		if g.checkKeyFromMap(v, "value") && g.getChildrenNumber(variable.Name) != 0 {
			if variable.Type != "char *" {
				address := g.getAddress(variable.Value)
				if !isNullPoint(address) {
					variable.Reference = convertReference(
						&referenceStruct{Type: "p", PointType: variable.Type, Address: address, VariableName: variable.Name})
				}
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

// getCurrentThreadId 获取当前线程id
func (g *gdbDebugger) getCurrentThreadId() (string, error) {
	// 获取当前线程id
	m, err := g.sendWithTimeOut(OptionTimeout, "thread-info")
	if err != nil {
		return "", err
	}
	threadMap, success := g.getPayloadFromMap(m)
	if !success {
		return "", errors.New("获取线程id失败")
	}
	currentThreadId := g.getStringFromMap(threadMap, "current-thread-id")
	return currentThreadId, nil
}

// getAddress value中获取地址，由于一个存储地址的value有时候会有各种类型
// 比如：0x555555602260; 0x555555601020 <globalItem>; 0x5555554008d4 "Hello, World!"
// 通过该方法从value中获取到地址
func (g *gdbDebugger) getAddress(value string) string {
	i := strings.Index(value, " ")
	if i == -1 {
		return value
	} else {
		return value[0:i]
	}
}

// convertVariableName 解析变量名称
// 由于某些结构体或者指针返回的名称不太美观，所以在这里进行一个转换
// 比如获取一个结构体的属性，属性名：localItem.id  ->  id
// 解引用情况：dynamicInt.*(int *)0x555555602260 -> *dynamicInt
// 数组情况：array.0 -> 0
func (g *gdbDebugger) convertVariableName(ref string, variableName string) string {
	index := strings.LastIndex(variableName, ".")
	if index == -1 {
		return variableName
	}
	if variableName[index+1] == '*' {
		refStruct, _ := parseReference(ref)
		return fmt.Sprintf("*%s", refStruct.VariableName)
	}
	if variableName[index+1] >= '0' && variableName[index+1] <= '9' {
		return variableName[index+1:]
	}
	return variableName[index+1:]
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
			g.parseStoppedData(payload)
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

func (g *gdbDebugger) parseStoppedData(m interface{}) {
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

// getChildrenNumber 获取children数量
func (g *gdbDebugger) getChildrenNumber(name string) int {
	_, _ = g.sendWithTimeOut(OptionTimeout, "var-create", name, "@", name)
	defer func() {
		_, _ = g.sendWithTimeOut(OptionTimeout, "var-delete", name)
	}()
	m, err := g.sendWithTimeOut(OptionTimeout, "var-info-num-children", name)
	if err != nil {
		return 0
	}
	payload, success := g.getPayloadFromMap(m)
	if !success {
		return 0
	}
	return g.getIntFromMap(payload, "numchild")
}
