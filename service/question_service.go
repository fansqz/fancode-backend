package service

import (
	"FanCode/api_models/response"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/models"
	"FanCode/setting"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

type QuestionService interface {
	InsertQuestion(question *models.Question) *e.Error
	UpdateQuestion(question *models.Question) *e.Error
	DeleteQuestion(id uint) *e.Error
	GetQuestionList(page int, pageSize int) ([]*response.QuestionResponseForList, *e.Error)
	UploadQuestionFile(ctx *gin.Context, file *multipart.FileHeader, questionNumber string) *e.Error
}

type questionService struct {
}

func NewQuestionService() QuestionService {
	return &questionService{}
}

func (q *questionService) InsertQuestion(question *models.Question) *e.Error {
	if dao.CheckQuestionNumber(question.Number) {
		return e.ErrQuestionNumberIsExist
	}
	//插入
	dao.InsertQuestion(question)
	return nil
}

func (q *questionService) UpdateQuestion(question *models.Question) *e.Error {
	err := dao.UpdateQuestion(question)
	if err != nil {
		log.Println(err)
		return e.ErrQuestionUpdateFailed
	}
	return nil
}

func (q *questionService) DeleteQuestion(id uint) *e.Error {
	// 读取question
	question, err := dao.GetQuestionByQuestionID(id)
	if err != nil {
		log.Println(err)
		return e.ErrQuestionDeleteFailed
	}
	if question == nil || question.Number == "" {
		return e.ErrQuestionNotExist
	}
	// 删除题目文件
	s := file_store.NewCOS()
	err = s.DeleteFolder(question.Path)
	if err != nil {
		return e.ErrQuestionDeleteFailed
	}
	// 删除题目
	err = dao.DeleteQuestionByID(id)
	if err != nil {
		return e.ErrQuestionDeleteFailed
	}
	return nil
}

// 读取一个列表的题目
func (q *questionService) GetQuestionList(page int, pageSize int) ([]*response.QuestionResponseForList, *e.Error) {

	questions, err := dao.GetQuestionList(page, pageSize)
	if err != nil {
		return nil, e.ErrQuestionListFailed
	}
	newQuestions := make([]*response.QuestionResponseForList, len(questions))
	for i := 0; i < len(questions); i++ {
		newQuestions[i] = response.NewQuestionResponseForList(questions[i])
	}
	return newQuestions, nil
}

func (q *questionService) UploadQuestionFile(ctx *gin.Context, file *multipart.FileHeader, questionNumber string) *e.Error {
	filename := file.Filename
	// 保存文件到本地
	tempPath := setting.Conf.FilePathConfig.TempDir
	tempPath = tempPath + "/" + utils.GetUUID()
	err := ctx.SaveUploadedFile(file, tempPath+"/"+filename)
	if err != nil {
		log.Println(err)
		return e.ErrQuestionFileUploadFailed
	}
	//解压
	err = utils.Extract(tempPath+"/"+filename, tempPath+"/"+questionNumber)
	if err != nil {
		log.Println(err)
		return e.ErrQuestionFileUploadFailed
	}
	//检测文件内有一个文件夹，或者是多个文件
	questionPathInLocal, _ := getSingleDirectoryPath(tempPath + "/" + questionNumber)
	s := file_store.NewCOS()
	err = s.DeleteFolder(questionNumber)
	s.UploadFolder(questionNumber, questionPathInLocal)
	// 存储到数据库
	updateError := dao.UpdatePathByNumber(questionNumber, questionNumber)
	if updateError != nil {
		return e.ErrQuestionFileUploadFailed
	}
	//删除temp中所有文件
	err = os.RemoveAll(tempPath)
	if err != nil {
		log.Println(err)
	}
	return nil
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
