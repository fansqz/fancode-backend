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
	"strings"
)

type ProblemService interface {
	CheckProblemCode(problemCode string) (bool, *e.Error)
	InsertProblem(problem *po.Problem) (uint, *e.Error)
	UpdateProblem(Problem *po.Problem) *e.Error
	DeleteProblem(id uint) *e.Error
	GetProblemList(page int, pageSize int) (*dto.PageInfo, *e.Error)
	UploadProblemFile(ctx *gin.Context, file *multipart.FileHeader, ProblemCode string) *e.Error

	// 根据io获取题目
	GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error)
	// 根据id获取题目文件
	GetProblemFileByID(id uint) ([]*dto.FileDto, *e.Error)
	// 根据id获取用例文件
	GetCaseFileByID(id uint, page int, pageSize int, fileType string) (*dto.PageInfo, *e.Error)
	// 更新一个字段
	UpdateProblemField(id uint, field string, value string) *e.Error
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

func (q *problemService) GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error) {
	problem, err := dao.GetProblemByProblemID(id)
	if err != nil {
		log.Println(err)
		return nil, e.ErrProblemGetFailed
	}
	return dto.NewProblemDtoForGet(problem), nil
}

func (q *problemService) GetProblemFileByID(id uint) ([]*dto.FileDto, *e.Error) {
	// 获取题目文件
	problem, err := dao.GetProblemByProblemID(id)
	if err != nil {
		return nil, e.ErrProblemGetFailed
	}
	if problem.Path == "" {
		return nil, e.ErrProblemFileNotExist
	}
	// 下载文件到本地
	err = checkAndDownloadQuestionFile(problem.Path)
	if err != nil {
		return nil, e.ErrProblemGetFailed
	}
	// 读取文件,仅会读取一个层级的文件
	localPath := getLocalPathByPath(problem.Path)
	files, err2 := os.ReadDir(localPath)
	if err2 != nil {
		return nil, e.ErrExecuteFailed
	}
	fileDtoList := make([]*dto.FileDto, 10)
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}
		var content []byte
		content, err = os.ReadFile(localPath + "/" + fileInfo.Name())
		if err != nil {
			continue
		}
		fileDto := &dto.FileDto{
			Name:    fileInfo.Name(),
			Content: string(content),
		}
		fileDtoList = append(fileDtoList, fileDto)
	}
	return fileDtoList, nil
}

func (q *problemService) GetCaseFileByID(id uint, page int, pageSize int, fileType string) (*dto.PageInfo, *e.Error) {
	// 获取题目文件
	problem, err := dao.GetProblemByProblemID(id)
	if err != nil {
		return nil, e.ErrProblemGetFailed
	}
	if problem.Path == "" {
		return nil, e.ErrProblemFileNotExist
	}
	// 下载文件到本地
	err = checkAndDownloadQuestionFile(problem.Path)
	if err != nil {
		return nil, e.ErrProblemGetFailed
	}
	//根据输入类型获取输入文件列表
	ioFileList := make([]string, 10)
	files, _ := os.ReadDir(getCaseFolderByPath(problem.Path))
	for _, fileInfo := range files {
		if strings.HasSuffix(fileInfo.Name(), "."+fileType) {
			ioFileList = append(ioFileList, fileInfo.Name())
		}
	}
	// 分页查询
	index1 := (page - 1) * pageSize
	index2 := index1 + pageSize
	if index2 > len(ioFileList) {
		index2 = len(ioFileList)
	}
	ioFileList2 := ioFileList[index1:index2]
	ioFilePageInfo := &dto.PageInfo{
		Total: uint(len(ioFileList)),
		Size:  uint(len(ioFileList2)),
		List:  ioFileList2,
	}
	return ioFilePageInfo, nil
}

func (q *problemService) UpdateProblemField(id uint, field string, value string) *e.Error {
	if field == "name" || field == "code" || field == "description" || field == "title" {
		err := dao.UpdateProblemField(id, field, value)
		if err != nil {
			log.Println(err)
			return e.ErrProblemFieldUpdateFailed
		}
		return nil
	}
	return e.ErrProblemFieldForbiddenUpdate
}

func getCaseFolderByPath(path string) string {
	localpath := getLocalPathByPath(path)
	return localpath + "/io"
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
