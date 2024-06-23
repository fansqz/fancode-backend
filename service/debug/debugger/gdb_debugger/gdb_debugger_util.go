package gdb_debugger

import (
	"errors"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (g *gdbDebugger) sendWithTimeOut(timeout time.Duration, operation string, args ...string) (map[string]interface{}, error) {
	channel := make(chan map[string]interface{}, 1)

	err := g.gdb.SendAsync(func(obj map[string]interface{}) {
		channel <- obj
	}, operation, args...)
	if err != nil {
		return nil, err
	}
	select {
	case m := <-channel:
		return m, nil
	case <-time.After(timeout):
		return nil, errors.New("GetStackTrace time out")
	}
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

func (g *gdbDebugger) getInterfaceFromMap(m interface{}, key string) interface{} {
	s, ok := m.(map[string]interface{})
	if !ok {
		return nil
	}
	answer, _ := s[key]
	return answer
}

func (g *gdbDebugger) getStringFromMap(m interface{}, key string) string {
	answer := g.getInterfaceFromMap(m, key)
	if answer == nil {
		return ""
	}
	strAnswer, _ := answer.(string)
	return strAnswer
}

func (g *gdbDebugger) getIntFromMap(m interface{}, key string) int {
	answer := g.getStringFromMap(m, key)
	numAnswer, _ := strconv.Atoi(answer)
	return numAnswer
}

func (g *gdbDebugger) getListFromMap(m interface{}, key string) []interface{} {
	s, _ := m.(map[string]interface{})[key]
	s2, _ := s.([]interface{})
	return s2
}

func (g *gdbDebugger) mapSet(m interface{}, key string, value string) {
	m2, _ := m.(map[string]interface{})
	m2[key] = value
}

func (g *gdbDebugger) mapDelete(m interface{}, key string) {
	m2, _ := m.(map[string]interface{})
	delete(m2, key)
}

// 检查map中是否有某个key
func (g *gdbDebugger) checkKeyFromMap(m interface{}, key string) bool {
	s, _ := m.(map[string]interface{})
	_, exist := s[key]
	return exist
}

func (g *gdbDebugger) getPayloadFromMap(m map[string]interface{}) (interface{}, bool) {
	if class := g.getStringFromMap(m, "class"); class == "done" {
		if payload, ok := m["payload"]; ok {
			return payload, true
		} else {
			return nil, false
		}
	}
	return nil, false
}

// isNullPoint 判断是否是空指针
// 0x0为空指针。解析16进制，如果为0则为null
func isNullPoint(address string) bool {
	address = address[2:]
	num, _ := strconv.ParseInt(address, 16, 64)
	return num == 0
}

// isPointType 判断一个类型是否是指针类型
// 比如：int *; Color *
func isPointType(typ string) bool {
	return strings.LastIndex(typ, "*") != -1
}

// parseBreakpoint 解析断点
func parseBreakpoint(bk string) (string, int) {
	l := strings.Split(bk, ":")
	if len(l) != 2 {
		return "", 0
	}
	file := l[0]
	lineStr := l[1]
	line, _ := strconv.Atoi(lineStr)
	return file, line
}
