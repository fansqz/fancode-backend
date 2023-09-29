package dto

import (
	"FanCode/models/po"
	"FanCode/utils"
)

// ProblemDtoForGet 获取题目详细信息
type ProblemDtoForGet struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Number      string `json:"number"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Path        string `json:"path"`
	Difficulty  int    `json:"difficulty"`
	// 支持的语言用,分割
	Languages string `json:"languages"`
	Enable    int    `json:"enable"`
}

func NewProblemDtoForGet(problem *po.Problem) *ProblemDtoForGet {
	response := &ProblemDtoForGet{
		ID:          problem.ID,
		Name:        problem.Name,
		Number:      problem.Number,
		Description: problem.Description,
		Title:       problem.Title,
		Path:        problem.Path,
		Difficulty:  problem.Difficulty,
		Languages:   problem.Languages,
		Enable:      problem.Enable,
	}
	return response
}

// ProblemDtoForList 获取题目列表
type ProblemDtoForList struct {
	ID         uint       `json:"id"`
	CreatedAt  utils.Time `json:"createdAt"`
	UpdatedAt  utils.Time `json:"updatedAt"`
	Name       string     `json:"name"`
	Number     string     `json:"number"`
	Title      string     `json:"title"`
	Path       string     `json:"path"`
	Difficulty int        `json:"difficulty"`
	Enable     int        `json:"enable"`
}

func NewProblemDtoForList(problem *po.Problem) *ProblemDtoForList {
	response := &ProblemDtoForList{
		ID:         problem.ID,
		CreatedAt:  utils.Time(problem.CreatedAt),
		UpdatedAt:  utils.Time(problem.UpdatedAt),
		Name:       problem.Name,
		Number:     problem.Number,
		Title:      problem.Title,
		Path:       problem.Path,
		Difficulty: problem.Difficulty,
		Enable:     problem.Enable,
	}
	return response
}

// ProblemFileDto 获取题目文件列表的dto
type ProblemFileDto struct {
	// 输入文件分页
	Files []*FileDto
	// 输入或输出文件
	IOFileInfo PageInfo
	// 输入输出文件类型
	IOFileType string
}

// ProblemDtoForUserList 用户获取题目列表的时候返回的题目数据
type ProblemDtoForUserList struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Number      string `json:"number"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Difficulty  int    `json:"difficulty"`
	// 学生做题状态
	State int `json:"state"`
}

func NewProblemDtoForUserList(problem *po.Problem) *ProblemDtoForUserList {
	return &ProblemDtoForUserList{
		ID:          problem.ID,
		Name:        problem.Name,
		Number:      problem.Number,
		Description: problem.Description,
		Title:       problem.Title,
		Difficulty:  problem.Difficulty,
	}
}
