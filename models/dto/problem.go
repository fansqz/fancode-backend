package dto

import (
	"FanCode/models/po"
	"FanCode/utils"
)

// 获取题目详细信息
type ProblemDtoForGet struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Path        string `json:"path"`
}

func NewProblemDtoForGet(problem *po.Problem) *ProblemDtoForGet {
	response := &ProblemDtoForGet{
		ID:          problem.ID,
		Name:        problem.Name,
		Code:        problem.Code,
		Description: problem.Description,
		Title:       problem.Title,
		Path:        problem.Path,
	}
	return response
}

// 获取题目列表
type ProblemDtoForList struct {
	ID        uint       `json:"id"`
	CreatedAt utils.Time `json:"createdAt"`
	UpdatedAt utils.Time `json:"updatedAt"`
	Name      string     `json:"name"`
	Code      string     `json:"code"`
	Title     string     `json:"title"`
	Path      string     `json:"path"`
}

func NewProblemDtoForList(question *po.Problem) *ProblemDtoForList {
	response := &ProblemDtoForList{
		ID:        question.ID,
		CreatedAt: utils.Time(question.CreatedAt),
		UpdatedAt: utils.Time(question.UpdatedAt),
		Name:      question.Name,
		Code:      question.Code,
		Title:     question.Title,
		Path:      question.Path,
	}
	return response
}

// 获取题目文件列表的dto
type ProblemFileDto struct {
	// 输入文件分页
	Files []*FileDto
	// 输入或输出文件
	IOFileInfo PageInfo
	// 输入输出文件类型
	IOFileType string
}
