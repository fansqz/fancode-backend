package controllers

import (
	"FanCode/dao"
	"FanCode/models"
	r "FanCode/result"
	"FanCode/store"
	"github.com/gin-gonic/gin"
	"strconv"
)

// QuestionController
// @Description: 题目管理相关功能
type QuestionController interface {
	InsertQuestion(ctx *gin.Context)
}

type questionController struct {
}

func NewQuestionController() QuestionController {
	return &questionController{}
}

func (q *questionController) InsertQuestion(ctx *gin.Context) {
	result := r.NewResult(ctx)
	questionNumber := ctx.PostForm("number")
	questionName := ctx.PostForm("name")
	description := ctx.PostForm("description")
	title := ctx.PostForm("title")
	path := ctx.PostForm("path")
	if dao.CheckQuestionNumber(questionNumber) {
		result.SimpleErrorMessage("题目编号已存在")
		return
	}
	question := &models.Question{}
	question.Number = questionNumber
	question.Name = questionName
	question.Description = description
	question.Title = title
	question.Path = path
	//插入
	dao.InsertQuestion(question)

	result.SuccessMessage("题库添加成功")

}

func (q *questionController) UpdateQuestion(ctx *gin.Context) {
	result := r.NewResult(ctx)
	questionIDString := ctx.PostForm("id")
	quesetionID, err := strconv.Atoi(questionIDString)
	if err != nil {
		result.SimpleErrorMessage("题目id出错")
	}
	questionNumber := ctx.PostForm("number")
	questionName := ctx.PostForm("name")
	description := ctx.PostForm("description")
	title := ctx.PostForm("title")
	path := ctx.PostForm("path")

	question := &models.Question{}
	question.ID = uint(quesetionID)
	question.Number = questionNumber
	question.Name = questionName
	question.Description = description
	question.Title = title
	question.Path = path

	dao.UpdateQuestion(question)
	result.SuccessData("修改成功")
}

func (q *questionController) DeleteQuestion(ctx *gin.Context) {
	result := r.NewResult(ctx)
	ids := ctx.Param("id")
	id, convertErr := strconv.Atoi(ids)
	if convertErr != nil {
		result.SimpleErrorMessage("id错误")
		return
	}
	// 读取question
	question, err := dao.GetQuestionByQuestionID(uint(id))
	if err != nil {
		result.SimpleErrorMessage("不存在该题目")
		return
	}
	// 删除题目文件
	s := store.NewCOS()
	s.DeleteFolder(question.Path)
	result.SuccessData("删除成功")
}
