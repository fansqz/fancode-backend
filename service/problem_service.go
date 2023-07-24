package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/setting"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

type ProblemService interface {
	GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error)
	InsertProblem(Problem *po.Problem) *e.Error
	UpdateProblem(Problem *po.Problem) *e.Error
	DeleteProblem(id uint) *e.Error
	GetProblemList(page int, pageSize int) (*dto.PageInfo, *e.Error)
	UploadProblemFile(ctx *gin.Context, file *multipart.FileHeader, ProblemNumber string) *e.Error
}

type problemService struct {
}

func NewProblemService() ProblemService {
	return &problemService{}
}

func (q *problemService) GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error) {
	problem, err := dao.GetProblemByProblemID(id)
	if err != nil {
		log.Println(err)
		return nil, e.ErrProblemGetFailed
	}
	return dto.NewProblemDtoForGet(problem), nil
}

func (q *problemService) InsertProblem(Problem *po.Problem) *e.Error {
	if dao.CheckProblemNumber(Problem.Number) {
		return e.ErrProblemNumberIsExist
	}
	//插入
	dao.InsertProblem(Problem)
	return nil
}

func (q *problemService) UpdateProblem(Problem *po.Problem) *e.Error {
	err := dao.UpdateProblem(Problem)
	if err != nil {
		log.Println(err)
		return e.ErrProblemUpdateFailed
	}
	return nil
}

func (q *problemService) DeleteProblem(id uint) *e.Error {
	// 读取Problem
	Problem, err := dao.GetProblemByProblemID(id)
	if err != nil {
		log.Println(err)
		return e.ErrProblemDeleteFailed
	}
	if Problem == nil || Problem.Number == "" {
		return e.ErrProblemNotExist
	}
	// 删除题目文件
	s := file_store.NewCOS()
	err = s.DeleteFolder(Problem.Path)
	if err != nil {
		return e.ErrProblemDeleteFailed
	}
	// 删除题目
	err = dao.DeleteProblemByID(id)
	if err != nil {
		return e.ErrProblemDeleteFailed
	}
	return nil
}

// 读取一个列表的题目
func (q *problemService) GetProblemList(page int, pageSize int) (*dto.PageInfo, *e.Error) {
	// 获取题目列表
	problems, err := dao.GetProblemList(page, pageSize)
	if err != nil {
		return nil, e.ErrProblemListFailed
	}
	newProblems := make([]*dto.ProblemDtoForList, len(problems))
	for i := 0; i < len(problems); i++ {
		newProblems[i] = dto.NewProblemDtoForList(problems[i])
	}
	// 获取所有题目总数目
	var count uint
	count, err = dao.GetProblemCount()
	if err != nil {
		return nil, e.ErrProblemListFailed
	}
	pageInfo := &dto.PageInfo{
		Total: count,
		Size:  uint(len(newProblems)),
		List:  newProblems,
	}
	return pageInfo, nil
}

func (q *problemService) UploadProblemFile(ctx *gin.Context, file *multipart.FileHeader, ProblemNumber string) *e.Error {
	filename := file.Filename
	// 保存文件到本地
	tempPath := setting.Conf.FilePathConfig.TempDir
	tempPath = tempPath + "/" + utils.GetUUID()
	err := ctx.SaveUploadedFile(file, tempPath+"/"+filename)
	if err != nil {
		log.Println(err)
		return e.ErrProblemFileUploadFailed
	}
	//解压
	err = utils.Extract(tempPath+"/"+filename, tempPath+"/"+ProblemNumber)
	if err != nil {
		log.Println(err)
		return e.ErrProblemFileUploadFailed
	}
	//检测文件内有一个文件夹，或者是多个文件
	ProblemPathInLocal, _ := getSingleDirectoryPath(tempPath + "/" + ProblemNumber)
	s := file_store.NewCOS()
	err = s.DeleteFolder(ProblemNumber)
	s.UploadFolder(ProblemNumber, ProblemPathInLocal)
	// 存储到数据库
	updateError := dao.UpdatePathByNumber(ProblemNumber, ProblemNumber)
	if updateError != nil {
		return e.ErrProblemFileUploadFailed
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
