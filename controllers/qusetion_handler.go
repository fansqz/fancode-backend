package controllers

import (
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

// QuestionController
// @Description: 题目管理相关功能
type QuestionController interface {
	InsertQuestion(ctx *gin.Context)
	UpdateQuestion(ctx *gin.Context)
	DeleteQuestion(ctx *gin.Context)
	GetQuestionList(ctx *gin.Context)
	UploadQuestionFile(ctx *gin.Context)
}

type questionController struct {
	questionService service.QuestionService
}

func NewQuestionController() QuestionController {
	return &questionController{
		questionService: service.NewQuestionService(),
	}
}

func (q *questionController) InsertQuestion(ctx *gin.Context) {
	result := r.NewResult(ctx)
	question := &po.Question{}
	question.Number = ctx.PostForm("number")
	question.Name = ctx.PostForm("name")
	question.Description = ctx.PostForm("description")
	question.Title = ctx.PostForm("title")
	question.Path = ctx.PostForm("path")
	//插入
	err := q.questionService.InsertQuestion(question)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("题库添加成功")
}

func (q *questionController) UpdateQuestion(ctx *gin.Context) {
	result := r.NewResult(ctx)
	questionIDString := ctx.PostForm("id")
	quesetionID, err := strconv.Atoi(questionIDString)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	question := &po.Question{}
	question.ID = uint(quesetionID)
	question.Number = ctx.PostForm("number")
	question.Name = ctx.PostForm("name")
	question.Description = ctx.PostForm("description")
	question.Title = ctx.PostForm("title")
	question.Path = ctx.PostForm("path")

	err2 := q.questionService.InsertQuestion(question)
	if err != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("修改成功")
}

func (q *questionController) DeleteQuestion(ctx *gin.Context) {
	result := r.NewResult(ctx)
	ids := ctx.Param("id")
	id, convertErr := strconv.Atoi(ids)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := q.questionService.DeleteQuestion(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("删除成功")
}

// 读取一个列表的题目
func (q *questionController) GetQuestionList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	var page int
	var pageSize int
	var convertErr error
	page, convertErr = strconv.Atoi(pageStr)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	pageSize, convertErr = strconv.Atoi(pageSizeStr)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	questions, err := q.questionService.GetQuestionList(page, pageSize)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(questions)
}

func (q *questionController) UploadQuestionFile(ctx *gin.Context) {
	result := r.NewResult(ctx)
	file, err := ctx.FormFile("questionFile")
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	questionNumber := ctx.PostForm("questionNumber")
	// 保存文件到本地
	uploadErr := q.questionService.UploadQuestionFile(ctx, file, questionNumber)
	if uploadErr != nil {
		result.Error(uploadErr)
		return
	}
	result.SuccessData("题目文件上传成功")
}
