package gdb_debugger

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// referenceStruct 定义的引用结构体
type referenceStruct struct {
	Type         string
	FrameId      string
	VariableName string
	PointType    string
	Address      string
	FieldPath    string
}

// parseReference 解析引用，gdbDebugger将地址或者结构体都抽象成引用
// 目前一共有两种类型的引用，分别是：
// 1. 普通类型：v-<栈帧id>-<变量名>-<属性（如果有）>;
// 2.指针：p-<指针类型>-<地址>-<地址的变量名称>-<属性（如果有）>
func parseReference(reference string) (*referenceStruct, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(reference)
	if err != nil {
		return nil, err
	}
	reference = string(decodedBytes)
	// 正则表达式，捕获栈帧ID和变量名
	t := strings.Split(reference, "-")
	// 在字符串中匹配正则表达式
	if len(t) < 3 {
		fmt.Println("No matches found")
		return nil, errors.New("引用格式有问题")
	}
	answer := &referenceStruct{}
	if t[0] == "p" {
		answer.Type = t[0]
		answer.PointType = t[1]
		answer.Address = t[2]
		answer.VariableName = t[3]
		if len(t) > 4 {
			answer.FieldPath = t[4]
		}
		return answer, nil
	}
	if t[0] == "v" {
		answer.Type = t[0]
		answer.FrameId = t[1]
		answer.VariableName = t[2]
		if len(t) > 3 {
			answer.FieldPath = t[3]
		}
		return answer, nil
	}
	return nil, errors.New("引用格式有问题")
}

// convertReference 把refStruct结构体转成引用字符串
func convertReference(refStruct *referenceStruct) string {
	var answer string
	if refStruct.Type == "p" {
		answer = fmt.Sprintf("p-%s-%s-%s", refStruct.PointType, refStruct.Address, refStruct.VariableName)
	} else if refStruct.Type == "v" {
		answer = fmt.Sprintf("v-%s-%s", refStruct.FrameId, refStruct.VariableName)
	}
	if refStruct.FieldPath == "" {
		return base64.StdEncoding.EncodeToString([]byte(answer))
	}
	ref := fmt.Sprintf("%s-%s", answer, refStruct.FieldPath)
	return base64.StdEncoding.EncodeToString([]byte(ref))
}

// getFieldReference 通过一个已知结构体的引用，获取该结构体的某个属性的引用
func getFieldReference(reference string, field string) string {
	ref, _ := parseReference(reference)
	if ref.FieldPath == "" {
		return fmt.Sprintf("%s-%s", reference, field)
	} else {
		return fmt.Sprintf("%s.%s", reference, field)
	}
}
