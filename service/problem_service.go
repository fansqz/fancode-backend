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
	CheckProblemCode(problemCode string) (bool, *e.Error)
	InsertProblem(problem *po.Problem) (uint, *e.Error)
	UpdateProblem(Problem *po.Problem) *e.Error
	GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error)
	DeleteProblem(id uint) *e.Error
	GetProblemList(page int, pageSize int) (*dto.PageInfo, *e.Error)
	UploadProblemFile(ctx *gin.Context, file *multipart.FileHeader, ProblemCode string) *e.Error
}

type problemService struct {
}

func NewProblemService() ProblemService {
	return &problemService{}
}

func (q *problemService) CheckProblemCode(problemCode string) (bool, *e.Error) {
	b, err := dao.CheckProblemCodeExists(problemCode)
	if err != nil {
		return !b, e.ErrProblemCodeCheckFailed
	}
	return !b, nil
}

func (q *problemService) GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error) {
	problem, err := dao.GetProblemByProblemID(id)
	if err != nil {
		log.Println(err)
		return nil, e.ErrProblemGetFailed
	}
	return dto.NewProblemDtoForGet(problem), nil
}

func (q *problemService) InsertProblem(problem *po.Problem) (uint, *e.Error) {
	// 对设置值的数据设置默认值
	if problem.Name == "" {
		problem.Name = "未命名题目"
	}
	if problem.Title == "" {
		problem.Title = "标题信息"
	}
	if problem.Description == "" {
		problemDescription, err := os.ReadFile(setting.Conf.FilePathConfig.ProblemDescriptionTemplate)
		if err != nil {
			log.Println(err)
			return 0, e.ErrProblemInsertFailed
		}
		problem.Description = string(problemDescription)
	}
	if problem.Code == "" {
		problem.Code = "未命名编号" + utils.GetGenerateUniqueCode()
	}
	// 检测编号是否重复
	if problem.Code != "" {
		b, checkError := dao.CheckProblemCodeExists(problem.Code)
		if checkError != nil {
			return 0, e.ErrProblemInsertFailed
		}
		if b {
			return 0, e.ErrProblemCodeIsExist
		}
	}
	// 添加
	err := dao.InsertProblem(problem)
	if err != nil {
		return 0, e.ErrProblemInsertFailed
	}
	return problem.ID, nil
}

func (q *problemService) UpdateProblem(problem *po.Problem) *e.Error {
	err := dao.UpdateProblem(problem)
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
	if Problem == nil || Problem.Code == "" {
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

func (q *problemService) UploadProblemFile(ctx *gin.Context, file *multipart.FileHeader, problemCode string) *e.Error {
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
	err = utils.Extract(tempPath+"/"+filename, tempPath+"/"+problemCode)
	if err != nil {
		log.Println(err)
		return e.ErrProblemFileUploadFailed
	}
	//检测文件内有一个文件夹，或者是多个文件
	ProblemPathInLocal, _ := getSingleDirectoryPath(tempPath + "/" + problemCode)
	s := file_store.NewCOS()
	err = s.DeleteFolder(problemCode)
	s.UploadFolder(problemCode, ProblemPathInLocal)
	// 存储到数据库
	updateError := dao.UpdatePathByCode(problemCode, problemCode)
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
