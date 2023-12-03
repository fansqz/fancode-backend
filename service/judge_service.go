package service

import (
	conf "FanCode/config"
	"FanCode/constants"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/service/judger"
	"FanCode/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"os"
	"path"
	"strings"
	time "time"
)

const (
	// 限制时间和内存
	LimitExecuteTime   = int64(15 * time.Second)
	LimitExecuteMemory = 100 * 1024 * 1024
	QuotaExecuteCpu    = 100000
	// 限制编译时间
	LimitCompileTime = int64(10 * time.Second)
)

type JudgeService interface {
	// Submit 答案提交
	Submit(ctx *gin.Context, judgeRequest *dto.SubmitRequestDto) (*dto.SubmitResultDto, *e.Error)
	// Execute 执行
	Execute(judgeRequest *dto.ExecuteRequestDto) (*dto.ExecuteResultDto, *e.Error)
	// SaveCode 保存用户代码
	SaveCode(ctx *gin.Context, problemID uint, language string, code string) *e.Error
	// GetCode 读取用户代码
	GetCode(ctx *gin.Context, problemID uint) (*dto.UserCodeDto, *e.Error)
}

type judgeService struct {
	config            *conf.AppConfig
	judgeCore         *judger.JudgeCore
	problemService    ProblemService
	submissionDao     dao.SubmissionDao
	problemAttemptDao dao.ProblemAttemptDao
	problemDao        dao.ProblemDao
}

func NewJudgeService(config *conf.AppConfig, problemService ProblemService, submissionDao dao.SubmissionDao,
	attemptDao dao.ProblemAttemptDao, problemDao dao.ProblemDao) JudgeService {
	return &judgeService{
		config:            config,
		judgeCore:         judger.NewJudgeCore(),
		problemService:    problemService,
		submissionDao:     submissionDao,
		problemAttemptDao: attemptDao,
		problemDao:        problemDao,
	}
}

func (j *judgeService) Submit(ctx *gin.Context, judgeRequest *dto.SubmitRequestDto) (*dto.SubmitResultDto, *e.Error) {
	// 提交获取结果
	submission, err := j.submit(ctx, judgeRequest)
	if err != nil {
		return nil, err
	}

	// 插入提交数据
	tx := global.Mysql.Begin()
	_ = j.submissionDao.InsertSubmission(tx, submission)

	// 检测用户是否保存了attempt
	userId := ctx.Keys["user"].(*dto.UserInfo).ID
	problemAttempt, err2 := j.problemAttemptDao.GetProblemAttemptByID(tx, userId, judgeRequest.ProblemID)
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		return nil, e.ErrSubmitFailed
	}

	// 如果本身就没有记录，就插入
	if errors.Is(err2, gorm.ErrRecordNotFound) {
		problemAttempt = &po.ProblemAttempt{
			UserID:    userId,
			ProblemID: judgeRequest.ProblemID,
			Code:      judgeRequest.Code,
			Language:  judgeRequest.Language,
			Status:    constants.InProgress,
		}
		problemAttempt.SubmissionCount++
		if submission.Status == constants.Accepted {
			problemAttempt.SuccessCount++
		} else {
			problemAttempt.ErrCount++
		}
		problemAttempt.Code = judgeRequest.Code
		if problemAttempt.Status == 0 && submission.Status == constants.Accepted {
			problemAttempt.Status = 1
		}
		err2 = j.problemAttemptDao.InsertProblemAttempt(tx, problemAttempt)
		if err2 != nil {
			tx.Rollback()
			return nil, e.ErrSubmitFailed
		}
		tx.Commit()
		return dto.NewSubmitResultDto(submission), nil
	}

	problemAttempt.Code = judgeRequest.Code
	problemAttempt.Language = judgeRequest.Language
	// 有记录则更新
	problemAttempt.SubmissionCount++
	if submission.Status == constants.Accepted {
		problemAttempt.Status = constants.Success
		problemAttempt.SuccessCount++
	} else {
		if problemAttempt.Status != constants.Success {
			problemAttempt.Status = constants.InProgress
		}
		problemAttempt.ErrCount++
	}
	if err2 = j.problemAttemptDao.UpdateProblemAttempt(tx, problemAttempt); err2 != nil {
		tx.Rollback()
		return nil, e.ErrSubmitFailed
	}
	tx.Commit()
	return dto.NewSubmitResultDto(submission), nil
}

func (j *judgeService) submit(ctx *gin.Context, judgeRequest *dto.SubmitRequestDto) (*po.Submission, *e.Error) {

	// 提交结果对象
	submission := &po.Submission{
		Language:  judgeRequest.Language,
		Code:      judgeRequest.Code,
		ProblemID: judgeRequest.ProblemID,
		UserID:    ctx.Keys["user"].(*dto.UserInfo).ID,
	}

	//读取题目到本地
	problem, err := j.problemDao.GetProblemByID(global.Mysql, judgeRequest.ProblemID)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	if err = checkAndDownloadQuestionFile(j.config, problem.Path); err != nil {
		return nil, e.ErrExecuteFailed
	}

	// executePath 执行路径，用户的临时文件
	executePath := getExecutePath(j.config)
	if err = os.MkdirAll(executePath, os.ModePerm); err != nil {
		return nil, e.ErrExecuteFailed
	}
	defer os.RemoveAll(executePath)

	// 保存题目文件的路径
	problemPath := getLocalProblemPath(j.config, problem.Path)

	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err2 *e.Error
	if compileFiles, err2 = j.saveUserCode(judgeRequest.Language,
		judgeRequest.Code, executePath); err2 != nil {
		return nil, err2
	}

	// 输出执行文件路劲
	executeFilePath := path.Join(executePath, "main")

	// 执行编译
	compileOptions := &judger.CompileOptions{
		ExcludedPaths: []string{executePath},
		Language:      judgeRequest.Language,
		LimitTime:     LimitCompileTime,
	}
	var compileResult *judger.CompileResult
	if compileResult, err = j.judgeCore.Compile(compileFiles, executeFilePath, compileOptions); err != nil {
		return nil, e.ErrUnknown
	}
	if !compileResult.Compiled {
		submission.Status = constants.CompileError
		submission.ErrorMessage = err.Error()
		return submission, nil
	}

	// 运行
	casePathForIn := path.Join(problemPath, "in")
	casePathForOut := path.Join(problemPath, "out")
	files, err3 := os.ReadDir(casePathForIn)
	if err3 != nil {
		return nil, e.ErrServer
	}
	inputCh := make(chan []byte)
	outputCh := make(chan judger.ExecuteResult)
	exitCh := make(chan string)
	defer func() {
		exitCh <- "exit"
	}()
	executeOption := &judger.ExecuteOptions{
		Language:      judgeRequest.Language,
		LimitTime:     LimitExecuteTime,
		MemoryLimit:   LimitExecuteMemory,
		CPUQuota:      QuotaExecuteCpu,
		ExcludedPaths: []string{executePath},
	}
	// 运行可执行文件
	if err = j.judgeCore.Execute(executeFilePath, inputCh, outputCh, exitCh, executeOption); err != nil {
		return nil, e.ErrUnknown
	}

	beginTime := time.Now()
	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".in") {

			// 输入数据
			var input []byte
			input, err = os.ReadFile(path.Join(casePathForIn, fileInfo.Name()))
			if err != nil {
				log.Println(err)
				return nil, e.ErrExecuteFailed
			}
			inputCh <- input

			// 读取输出数据
			select {
			case executeResult := <-outputCh:
				// 运行出错
				if !executeResult.Executed {
					submission.Status = constants.RuntimeError
					submission.ErrorMessage = executeResult.ErrorMessage
					return submission, nil
				}

				// 读取.out文件
				outFilePath := path.Join(casePathForOut, strings.ReplaceAll(fileInfo.Name(), ".in", ".out"))
				var outFileContent []byte
				if outFileContent, err = os.ReadFile(outFilePath); err != nil {
					return nil, e.ErrExecuteFailed
				}

				// 结果不正确则结束
				if !j.compareAnswer(string(executeResult.Output), string(outFileContent)) {
					submission.Status = constants.WrongAnswer
					submission.CaseName = strings.Split(fileInfo.Name(), ".")[0]
					submission.CaseData = string(input)
					submission.ExpectedOutput = string(outFileContent)
					submission.UserOutput = string(executeResult.Output)
					return submission, nil
				}
			}
		}
	}
	endTime := time.Now()
	submission.Status = constants.Accepted
	submission.TimeUsed = endTime.Sub(beginTime)
	return submission, nil
}

// saveUserCode
// 保存用户代码到用户的executePath，并返回需要编译的文件列表
func (j *judgeService) saveUserCode(language string, codeStr string, executePath string) ([]string, *e.Error) {
	var compileFiles []string
	var mainFile string
	var err *e.Error
	var err2 error

	if mainFile, err = getMainFileNameByLanguage(language); err != nil {
		return nil, err
	}
	if err2 = os.WriteFile(path.Join(executePath, mainFile), []byte(codeStr), 0644); err2 != nil {
		return nil, e.ErrServer
	}
	// 将main文件进行编译即可
	compileFiles = []string{path.Join(executePath, mainFile)}

	return compileFiles, nil
}

// 根据编程语言获取该编程语言的Main文件名称
func getMainFileNameByLanguage(language string) (string, *e.Error) {
	switch language {
	case constants.ProgramC:
		return "main.c", nil
	case constants.ProgramJava:
		return "Main.java", nil
	case constants.ProgramGo:
		return "main.go", nil
	default:
		return "", e.ErrLanguageNotSupported
	}
}

func (j *judgeService) Execute(judgeRequest *dto.ExecuteRequestDto) (*dto.ExecuteResultDto, *e.Error) {
	executeResult := &dto.ExecuteResultDto{
		ProblemID: judgeRequest.ProblemID,
	}

	//读取题目到本地，并编译
	problem, err := j.problemDao.GetProblemByID(global.Mysql, judgeRequest.ProblemID)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	if err = checkAndDownloadQuestionFile(j.config, problem.Path); err != nil {
		return nil, e.ErrExecuteFailed
	}

	// executePath 用户执行目录
	executePath := getExecutePath(j.config)
	if err = os.MkdirAll(executePath, os.ModePerm); err != nil {
		return nil, e.ErrExecuteFailed
	}
	defer os.RemoveAll(executePath)

	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err2 *e.Error
	if compileFiles, err2 = j.saveUserCode(judgeRequest.Language, judgeRequest.Code, executePath); err2 != nil {
		return nil, err2
	}

	// 输出的执行文件路劲
	executeFilePath := path.Join(executePath, "main")

	// 执行编译
	compileOptions := &judger.CompileOptions{
		Language:      judgeRequest.Language,
		LimitTime:     LimitCompileTime,
		ExcludedPaths: []string{executePath},
	}
	var compileResult *judger.CompileResult
	if compileResult, err = j.judgeCore.Compile(compileFiles, executeFilePath, compileOptions); err != nil {
		return nil, e.ErrUnknown
	}
	if !compileResult.Compiled {
		executeResult.Status = constants.CompileError
		executeResult.ErrorMessage = compileResult.ErrorMessage
		return executeResult, nil
	}

	//执行
	inputCh := make(chan []byte)
	outputCh := make(chan judger.ExecuteResult)
	exitCh := make(chan string)
	defer func() {
		exitCh <- "exit"
	}()
	executeOptions := &judger.ExecuteOptions{
		Language:      judgeRequest.Language,
		LimitTime:     LimitExecuteTime,
		MemoryLimit:   LimitExecuteMemory,
		CPUQuota:      QuotaExecuteCpu,
		ExcludedPaths: []string{executePath},
	}
	if err = j.judgeCore.Execute(executeFilePath, inputCh, outputCh, exitCh, executeOptions); err != nil {
		return nil, e.ErrUnknown
	}

	// 输入输入用例
	inputCh <- []byte(judgeRequest.Input)
	output := <-outputCh

	if !output.Executed {
		executeResult.Status = constants.RuntimeError
		executeResult.ErrorMessage = output.ErrorMessage
		return executeResult, nil
	}

	executeResult.Status = constants.RunSuccess
	executeResult.UserOutput = string(output.Output)
	return executeResult, nil
}

func (j *judgeService) SaveCode(ctx *gin.Context, problemID uint, language string, code string) *e.Error {
	userInfo := ctx.Keys["user"].(*dto.UserInfo)
	tx := global.Mysql.Begin()
	problemAttempt, err := j.problemAttemptDao.GetProblemAttemptByID(tx, userInfo.ID, problemID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return e.ErrMysql
	}

	// attempt不存在则添加
	if errors.Is(err, gorm.ErrRecordNotFound) {
		problemAttempt = &po.ProblemAttempt{
			UserID:          userInfo.ID,
			ProblemID:       problemID,
			Code:            code,
			Language:        language,
			SubmissionCount: 0,
			Status:          0,
		}
		if err2 := j.problemAttemptDao.InsertProblemAttempt(tx, problemAttempt); err2 != nil {
			tx.Rollback()
			return e.ErrMysql
		}
		tx.Commit()
		return nil
	}

	// 存在则更新
	problemAttempt2 := &po.ProblemAttempt{
		UserID:    userInfo.ID,
		ProblemID: problemID,
		Code:      code,
		Language:  language,
	}
	if err = j.problemAttemptDao.UpdateProblemAttempt(tx, problemAttempt2); err != nil {
		tx.Rollback()
		return e.ErrMysql
	}
	tx.Commit()
	return nil
}

func (j *judgeService) GetCode(ctx *gin.Context, problemID uint) (*dto.UserCodeDto, *e.Error) {
	userInfo := ctx.Keys["user"].(*dto.UserInfo)
	problemAttempt, err := j.problemAttemptDao.GetProblemAttemptByID(global.Mysql, userInfo.ID, problemID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.ErrMysql
	}

	// 读取代码模板
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 读取题目
		problem, err := j.problemDao.GetProblemByID(global.Mysql, problemID)
		if err != nil {
			return nil, e.ErrMysql
		}
		code, err2 := j.problemService.GetProblemTemplateCode(problemID,
			strings.Split(problem.Languages, ",")[0])
		if err2 != nil {
			return nil, err2
		}
		return &dto.UserCodeDto{
			ProblemID: problemID,
			Code:      code,
			Language:  strings.Split(problem.Languages, ",")[0],
		}, nil
	}

	return dto.NewUserCodeDto(problemAttempt), nil
}

// 比较用户的答案，忽略\n和空格
func (j *judgeService) compareAnswer(data1 string, data2 string) bool {
	data1 = strings.Trim(data1, " ")
	data1 = strings.Trim(data1, "\n")
	data2 = strings.Trim(data2, " ")
	data2 = strings.Trim(data2, "\n")
	return data1 == data2
}

func checkAndDownloadQuestionFile(config *conf.AppConfig, questionPath string) error {
	localPath := path.Join(config.FilePathConfig.ProblemFileDir, questionPath)
	if !utils.CheckFolderExists(localPath) {
		// 拉取文件
		store := file_store.NewProblemCOS(config.COSConfig)
		if err := store.DownloadFolder(questionPath, localPath); err != nil {
			return err
		}
	}
	return nil
}
