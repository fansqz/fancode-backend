package controllers

import (
	"FanCode/dao"
	"FanCode/models"
	r "FanCode/result"
	"FanCode/setting"
	"FanCode/store"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
	"strconv"
)

// QuestionController
// @Description: 题目管理相关功能
type QuestionController interface {
	InsertQuestion(ctx *gin.Context)
	UpdateQuestion(ctx *gin.Context)
	DeleteQuestion(ctx *gin.Context)
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

// 读取一个列表的题目
func (q *questionController) GetQuestionList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageStr := ctx.Param("page")
	pageSizeStr := ctx.Param("pageSize")
	var page int
	var pageSize int
	var convertErr error
	page, convertErr = strconv.Atoi(pageStr)
	if convertErr != nil {
		result.SimpleErrorMessage("参数错误")
	}
	pageSize, convertErr = strconv.Atoi(pageSizeStr)
	questions, err := dao.GetQuestionList(page, pageSize)
	if err != nil {
		result.SimpleErrorMessage("读取失败")
	}
	result.SuccessData(questions)
}

func (q *questionController) UploadQuestionFile(ctx *gin.Context) {
	result := r.NewResult(ctx)
	file, err := ctx.FormFile("questionFile")
	if err != nil {
		result.SimpleErrorMessage("文件上传失败")
		return
	}
	filename := file.Filename
	questionNumber := ctx.PostForm("questionNumber")
	// 保存文件到本地
	tempPath := setting.Conf.FilePathConfig.TempDir
	tempPath = tempPath + "/" + utils.GetUUID()
	err = ctx.SaveUploadedFile(file, tempPath)
	if err != nil {
		result.SimpleErrorMessage("文件存储失败")
		return
	}
	//解压
	err = utils.Extract(tempPath+"/"+filename, tempPath+"/"+questionNumber)
	if err != nil {
		result.SimpleErrorMessage("文件解压失败")
		return
	}
	//检测文件内有一个文件夹，或者是多个文件
	questionPathInLocal, _ := getSingleDirectoryPath(tempPath + "/" + questionNumber)
	s := store.NewCOS()
	s.DeleteFolder(questionNumber)
	s.UploadFolder(questionNumber, questionPathInLocal)
	// 存储到数据库
	updateError := dao.UpdatePathByNumber(questionNumber, questionNumber)
	if updateError != nil {
		result.SimpleErrorMessage("题目数据标识存储到数据库失败")
		return
	}
	//删除temp中所有文件
	os.RemoveAll(tempPath)
	result.SuccessData("题目文件上传成功")
}

// 如果文件夹内有且仅有一个文件夹，返回内部文件夹路径
func getSingleDirectoryPath(path string) (string, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return path, err
	}

	// 检查目录中文件和文件夹的数量
	if len(dirEntries) != 1 || !dirEntries[0].IsDir() {
		return path, nil
	}

	return filepath.Join(path, dirEntries[0].Name()), nil
}
