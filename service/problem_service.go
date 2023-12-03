package service

import (
	conf "FanCode/config"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ProblemService interface {
	// CheckProblemNumber 检测题目编码
	CheckProblemNumber(problemCode string) (bool, *e.Error)
	// InsertProblem 添加题目
	InsertProblem(problem *po.Problem, ctx *gin.Context) (uint, *e.Error)
	// UpdateProblem 更新题目
	UpdateProblem(Problem *po.Problem, ctx *gin.Context, file *multipart.FileHeader) *e.Error
	// DeleteProblem 删除题目
	DeleteProblem(id uint) *e.Error
	// GetProblemList 获取题目列表
	GetProblemList(query *dto.PageQuery) (*dto.PageInfo, *e.Error)
	// GetUserProblemList 用户获取题目列表
	GetUserProblemList(ctx *gin.Context, query *dto.PageQuery) (*dto.PageInfo, *e.Error)
	// DownloadProblemZipFile 下载题目压缩文件
	DownloadProblemZipFile(ctx *gin.Context, problemID uint)
	// DownloadProblemTemplateFile 获取题目模板文件
	DownloadProblemTemplateFile(ctx *gin.Context)
	// GetProblemByID 获取题目信息
	GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error)
	// GetProblemByNumber 根据题目编号获取题目信息
	GetProblemByNumber(number string) (*dto.ProblemDtoForGet, *e.Error)
	// GetProblemTemplateCode 获取题目的模板代码
	GetProblemTemplateCode(problemID uint, language string) (string, *e.Error)
	// UpdateProblemEnable 设置题目可用
	UpdateProblemEnable(id uint, enable int) *e.Error

	// todo: 支持线上编辑题目
	GetProblemFileListByID(id uint) ([]*dto.FileDto, *e.Error)
	GetCaseFileByID(id uint, page int, pageSize int) (*dto.PageInfo, *e.Error)
	UpdateProblemField(id uint, field string, value string) *e.Error
}

type problemService struct {
	config            *conf.AppConfig
	problemDao        dao.ProblemDao
	problemAttemptDao dao.ProblemAttemptDao
}

func NewProblemService(config *conf.AppConfig, problemDao dao.ProblemDao, attemptDao dao.ProblemAttemptDao) ProblemService {
	return &problemService{
		config:            config,
		problemDao:        problemDao,
		problemAttemptDao: attemptDao,
	}
}

func (q *problemService) CheckProblemNumber(problemCode string) (bool, *e.Error) {
	b, err := q.problemDao.CheckProblemNumberExists(global.Mysql, problemCode)
	if err != nil {
		return !b, e.ErrProblemCodeCheckFailed
	}
	return !b, nil
}

func (q *problemService) InsertProblem(problem *po.Problem, ctx *gin.Context) (uint, *e.Error) {
	problem.CreatorID = ctx.Keys["user"].(*dto.UserInfo).ID
	// 对设置值的数据设置默认值
	if problem.Name == "" {
		problem.Name = "未命名题目"
	}
	if problem.Title == "" {
		problem.Title = "标题信息"
	}
	if problem.Description == "" {
		problemDescription, err := os.ReadFile(q.config.FilePathConfig.ProblemDescriptionTemplate)
		if err != nil {
			return 0, e.ErrProblemInsertFailed
		}
		problem.Description = string(problemDescription)
	}
	if problem.Number == "" {
		problem.Number = "未命名编号" + utils.GetGenerateUniqueCode()
	}
	// 检测编号是否重复
	if problem.Number != "" {
		b, checkError := q.problemDao.CheckProblemNumberExists(global.Mysql, problem.Number)
		if checkError != nil {
			return 0, e.ErrMysql
		}
		if b {
			return 0, e.ErrProblemCodeIsExist
		}
	}
	// 题目难度不在范围，那么都设置为1
	if problem.Difficulty > 5 || problem.Difficulty < 1 {
		problem.Difficulty = 1
	}
	problem.Enable = -1
	// 添加
	err := q.problemDao.InsertProblem(global.Mysql, problem)
	if err != nil {
		return 0, e.ErrMysql
	}
	return problem.ID, nil
}

func (q *problemService) UpdateProblem(problem *po.Problem, ctx *gin.Context, file *multipart.FileHeader) *e.Error {
	path, err := q.problemDao.GetProblemFilePathByID(global.Mysql, problem.ID)
	if err != nil {
		log.Println(err)
		return e.ErrProblemUpdateFailed
	}
	if problem.Enable == 1 && (path == "" && file == nil) {
		return e.NewCustomMsg("该题目没有上传编程文件，不可启动")
	}
	if file != nil {
		// 更新文件
		err2 := q.UploadProblemFile(ctx, file, problem.ID)
		return err2
	}
	problem.UpdatedAt = time.Now()
	// 更新题目
	err2 := q.problemDao.UpdateProblem(global.Mysql, problem)
	if err2 != nil {
		log.Println(err2)
		return e.ErrProblemUpdateFailed
	}
	return nil
}

// todo: 这里有事务相关的问题
func (q *problemService) DeleteProblem(id uint) *e.Error {
	// 读取Problem
	problem, err := q.problemDao.GetProblemByID(global.Mysql, id)
	if err != nil {
		return e.ErrMysql
	}
	if problem == nil || problem.Number == "" {
		return e.ErrProblemNotExist
	}
	if problem.Path != "" {
		// 删除题目文件
		s := file_store.NewProblemCOS(q.config.COSConfig)
		err = s.DeleteFolder(problem.Path)
		if err != nil {
			return e.ErrProblemDeleteFailed
		}
		// 删除本地文件
		localPath := getLocalProblemPath(q.config, problem.Path)
		err = utils.CheckAndDeletePath(localPath)
		if err != nil {
			return e.ErrProblemDeleteFailed
		}
	}
	// 删除题目
	err = q.problemDao.DeleteProblemByID(global.Mysql, id)
	if err != nil {
		return e.ErrMysql
	}
	return nil
}

func (q *problemService) GetProblemList(query *dto.PageQuery) (*dto.PageInfo, *e.Error) {
	var bankQuery *po.Problem
	if query.Query != nil {
		bankQuery = query.Query.(*po.Problem)
	}
	// 获取题目列表
	problems, err := q.problemDao.GetProblemList(global.Mysql, query)
	if err != nil {
		return nil, e.ErrMysql
	}
	newProblems := make([]*dto.ProblemDtoForList, len(problems))
	for i := 0; i < len(problems); i++ {
		newProblems[i] = dto.NewProblemDtoForList(problems[i])
	}
	// 获取所有题目总数目
	var count int64
	count, err = q.problemDao.GetProblemCount(global.Mysql, bankQuery)
	if err != nil {
		return nil, e.ErrMysql
	}
	pageInfo := &dto.PageInfo{
		Total: count,
		Size:  int64(len(newProblems)),
		List:  newProblems,
	}
	return pageInfo, nil
}

func (q *problemService) GetUserProblemList(ctx *gin.Context, query *dto.PageQuery) (*dto.PageInfo, *e.Error) {
	userId := ctx.Keys["user"].(*dto.UserInfo).ID
	if query.Query != nil {
		query.Query.(*po.Problem).Enable = 1
	} else {
		query.Query = &po.Problem{
			Enable: 1,
		}
	}
	// 获取题目列表
	problems, err := q.problemDao.GetProblemList(global.Mysql, query)
	if err != nil {
		return nil, e.ErrMysql
	}
	newProblems := make([]*dto.ProblemDtoForUserList, len(problems))
	for i := 0; i < len(problems); i++ {
		newProblems[i] = dto.NewProblemDtoForUserList(problems[i])
		// 读取题目完成情况
		var status int
		status, err = q.problemAttemptDao.GetProblemAttemptStatus(global.Mysql, userId, problems[i].ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, e.ErrProblemListFailed
		}
		newProblems[i].Status = status
	}
	// 获取所有题目总数目
	var count int64
	count, err = q.problemDao.GetProblemCount(global.Mysql, query.Query.(*po.Problem))
	if err != nil {
		return nil, e.ErrMysql
	}
	pageInfo := &dto.PageInfo{
		Total: count,
		Size:  int64(len(newProblems)),
		List:  newProblems,
	}
	return pageInfo, nil
}

// UploadProblemFile 保存到oss的时候，以id做为文件名
func (q *problemService) UploadProblemFile(ctx *gin.Context, file *multipart.FileHeader, problemID uint) *e.Error {
	path := strconv.Itoa(int(problemID))
	filename := file.Filename
	// 保存文件到本地
	tempPath := q.config.FilePathConfig.TempDir
	tempPath = tempPath + "/" + utils.GetUUID()
	err := ctx.SaveUploadedFile(file, tempPath+"/"+filename)
	if err != nil {
		log.Println(err)
		return e.ErrProblemFileUploadFailed
	}
	//解压
	err = utils.Extract(tempPath+"/"+filename, tempPath+"/"+path)
	if err != nil {
		log.Println(err)
		return e.ErrProblemFileUploadFailed
	}
	//检测文件内有一个文件夹，或者是多个文件
	ProblemPathInLocal, _ := getSingleDirectoryPath(tempPath + "/" + path)
	s := file_store.NewProblemCOS(q.config.COSConfig)
	err = s.DeleteFolder(strconv.Itoa(int(problemID)))
	s.UploadFolder(path, ProblemPathInLocal)
	// 存储到数据库
	updateError := q.problemDao.UpdatePathByID(global.Mysql, path, problemID)
	if updateError != nil {
		return e.ErrMysql
	}
	// 删除temp中所有文件
	os.RemoveAll(tempPath)
	// 删除本地题目的文件
	localPath := getLocalProblemPath(q.config, path)
	err = utils.CheckAndDeletePath(localPath)
	if err != nil {
		_ = utils.CheckAndDeletePath(localPath)
	}
	return nil
}

func (q *problemService) GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error) {
	problem, err := q.problemDao.GetProblemByID(global.Mysql, id)
	if err == gorm.ErrRecordNotFound {
		return nil, e.ErrProblemNotExist
	}
	if err != nil {
		return nil, e.ErrMysql
	}
	return dto.NewProblemDtoForGet(problem), nil
}

func (q *problemService) GetProblemByNumber(number string) (*dto.ProblemDtoForGet, *e.Error) {
	problem, err := q.problemDao.GetProblemByNumber(global.Mysql, number)
	if err != nil {
		return nil, e.ErrMysql
	}
	return dto.NewProblemDtoForGet(problem), nil
}

func (q *problemService) GetProblemTemplateCode(problemID uint, language string) (string, *e.Error) {
	// 读取acm模板
	code, err := getAcmCodeTemplate(language)
	if err != nil {
		return "", e.ErrProblemGetFailed
	}
	return code, nil
}

func (q *problemService) GetProblemFileListByID(id uint) ([]*dto.FileDto, *e.Error) {
	// 获取题目文件
	problem, err := q.problemDao.GetProblemByID(global.Mysql, id)
	if err != nil {
		return nil, e.ErrMysql
	}
	if problem.Path == "" {
		return nil, e.ErrProblemFileNotExist
	}
	// 下载文件到本地
	err = checkAndDownloadQuestionFile(q.config, problem.Path)
	if err != nil {
		return nil, e.ErrProblemGetFailed
	}
	// 读取文件,仅会读取一个层级的文件
	localPath := getLocalProblemPath(q.config, problem.Path)
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

func (q *problemService) GetCaseFileByID(id uint, page int, pageSize int) (*dto.PageInfo, *e.Error) {
	// 获取题目文件
	problem, err := q.problemDao.GetProblemByID(global.Mysql, id)
	if err != nil {
		return nil, e.ErrMysql
	}
	if problem.Path == "" {
		return nil, e.ErrProblemFileNotExist
	}
	// 下载文件到本地
	err = checkAndDownloadQuestionFile(q.config, problem.Path)
	if err != nil {
		return nil, e.ErrProblemGetFailed
	}
	//根据输入类型获取输入文件列表
	ioFileList := make([]string, 10)
	files, _ := os.ReadDir(getCaseFolderByPath(q.config, problem.Path))
	for _, fileInfo := range files {
		if strings.HasSuffix(fileInfo.Name(), ".in") {
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
		Total: int64(len(ioFileList)),
		Size:  int64(len(ioFileList2)),
		List:  ioFileList2,
	}
	return ioFilePageInfo, nil
}

func (q *problemService) UpdateProblemField(id uint, field string, value string) *e.Error {
	if field == "name" || field == "code" || field == "description" || field == "title" {
		err := q.problemDao.UpdateProblemField(global.Mysql, id, field, value)
		if err != nil {
			return e.ErrProblemUpdateFailed
		}
		return nil
	}
	return e.ErrProblemUpdateFailed
}

func getCaseFolderByPath(config *conf.AppConfig, path string) string {
	localpath := getLocalProblemPath(config, path)
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

func (q *problemService) DownloadProblemZipFile(ctx *gin.Context, problemID uint) {
	result := r.NewResult(ctx)
	path, err := q.problemDao.GetProblemFilePathByID(global.Mysql, problemID)
	if err != nil {
		result.Error(e.ErrProblemZipFileDownloadFailed)
		return
	}
	temp := getTempDir(q.config)
	// 最后删除临时文件夹
	defer func() {
		// 删除临时文件夹和压缩包
		err = os.RemoveAll(temp)
		if err != nil {
			os.RemoveAll(temp)
		}
	}()
	localPath := temp + "/" + strconv.Itoa(int(problemID))
	zipPath := localPath + ".zip"
	store := file_store.NewProblemCOS(q.config.COSConfig)
	err = store.DownloadAndCompressFolder(path, localPath, zipPath)
	if err != nil {
		result.Error(e.ErrProblemZipFileDownloadFailed)
		return
	}
	var content []byte
	content, err = os.ReadFile(zipPath)
	if err != nil {
		result.Error(e.ErrProblemZipFileDownloadFailed)
		return
	}
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Header("Content-Disposition", "attachment; filename="+strconv.Itoa(int(problemID))+".zip")
	ctx.Header("Content-Type", "application/zip")
	ctx.Writer.Write(content)
}

func (q *problemService) DownloadProblemTemplateFile(ctx *gin.Context) {
	result := r.NewResult(ctx)
	path := q.config.FilePathConfig.ProblemFileTemplate
	content, err := os.ReadFile(path)
	if err != nil {
		result.Error(e.ErrProblemZipFileDownloadFailed)
		return
	}
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Header("Content-Disposition", "attachment; filename="+"编程文件模板.zip")
	ctx.Header("Content-Type", "application/zip")
	ctx.Writer.Write(content)
}

// todo: 是否要加事务
func (q *problemService) UpdateProblemEnable(id uint, enable int) *e.Error {
	//检测题目文件是否存在
	problem, err := q.problemDao.GetProblemByID(global.Mysql, id)
	if err != nil {
		return e.ErrMysql
	}
	if problem.Path == "" {
		return e.ErrProblemFilePathNotExist
	}
	err = q.problemDao.SetProblemEnable(global.Mysql, id, enable)
	if err != nil {
		return e.ErrMysql
	}
	return nil
}
