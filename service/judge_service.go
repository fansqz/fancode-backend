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
	problemCaseDao    dao.ProblemCaseDao
	submissionDao     dao.SubmissionDao
	problemAttemptDao dao.ProblemAttemptDao
	problemDao        dao.ProblemDao
}

func NewJudgeService(config *conf.AppConfig, ps ProblemService, sd dao.SubmissionDao,
	ad dao.ProblemAttemptDao, pd dao.ProblemDao, pcd dao.ProblemCaseDao) JudgeService {
	return &judgeService{
		config:            config,
		judgeCore:         judger.NewJudgeCore(),
		problemService:    ps,
		problemCaseDao:    pcd,
		submissionDao:     sd,
		problemAttemptDao: ad,
		problemDao:        pd,
	}
}

func (j *judgeService) Submit(ctx *gin.Context, judgeRequest *dto.SubmitRequestDto) (*dto.SubmitResultDto, *e.Error) {
	// 提交获取结果
	submission, err := j.submit(ctx, judgeRequest)
	if err != nil {
		// Add logging for error
		log.Printf("Submit error: %v\n", err)
		return nil, err
	}

	// 插入提交数据
	tx := global.Mysql.Begin()
	_ = j.submissionDao.InsertSubmission(tx, submission)

	// 检测用户是否保存了attempt
	userId := ctx.Keys["user"].(*dto.UserInfo).ID
	problemAttempt, err2 := j.problemAttemptDao.GetProblemAttemptByID(tx, userId, judgeRequest.ProblemID)
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		// Add logging for error
		log.Println("GetProblemAttemptByID error: %v\n", err2)
		return nil, e.ErrSubmitFailed
	}

	// 如果本身就没有记录，就添加
	if errors.Is(err2, gorm.ErrRecordNotFound) {
		problemAttempt = &po.ProblemAttempt{
			UserID:    userId,
			ProblemID: judgeRequest.ProblemID,
			Code:      judgeRequest.Code,
			Language:  string(judgeRequest.Language),
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
			// Add logging for error
			log.Printf("InsertProblemAttempt error: %v\n", err2)
			return nil, e.ErrSubmitFailed
		}
		tx.Commit()
		return dto.NewSubmitResultDto(submission), nil
	}

	problemAttempt.Code = judgeRequest.Code
	problemAttempt.Language = string(judgeRequest.Language)
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
		// Add logging for error
		log.Printf("UpdateProblemAttempt error: %v\n", err2)
		return nil, e.ErrSubmitFailed
	}
	tx.Commit()
	return dto.NewSubmitResultDto(submission), nil
}

func (j *judgeService) submit(ctx *gin.Context, judgeRequest *dto.SubmitRequestDto) (*po.Submission, *e.Error) {
	// 提交结果对象
	submission := &po.Submission{
		Language:  string(judgeRequest.Language),
		Code:      judgeRequest.Code,
		ProblemID: judgeRequest.ProblemID,
		UserID:    ctx.Keys["user"].(*dto.UserInfo).ID,
	}

	// executePath 执行路径，用户的临时文件
	executePath := getExecutePath(j.config)
	if err := os.MkdirAll(executePath, os.ModePerm); err != nil {
		// Add logging for error
		log.Printf("MkdirAll error: %v\n", err)
		return nil, e.ErrExecuteFailed
	}
	defer os.RemoveAll(executePath)

	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err2 *e.Error
	if compileFiles, err2 = j.saveUserCode(judgeRequest.Language,
		judgeRequest.Code, executePath); err2 != nil {
		// Add logging for error
		log.Printf("SaveUserCode error: %v\n", err2)
		return nil, err2
	}

	// 输出执行文件路径
	executeFilePath := path.Join(executePath, "main")

	// 执行编译
	compileOptions := &judger.CompileOptions{
		ExcludedPaths: []string{executePath},
		Language:      judgeRequest.Language,
		LimitTime:     LimitCompileTime,
	}
	var compileResult *judger.CompileResult
	var err error
	if compileResult, err = j.judgeCore.Compile(compileFiles, executeFilePath, compileOptions); err != nil {
		// Add logging for error
		log.Printf("Compile error: %v\n", err)
		return nil, e.ErrUnknown
	}
	if !compileResult.Compiled {
		submission.Status = constants.CompileError
		submission.ErrorMessage = compileResult.ErrorMessage
		return submission, nil
	}

	// 运行
	caseList, err := j.problemCaseDao.GetProblemCaseList2(global.Mysql, judgeRequest.ProblemID)
	if err != nil {
		// Add logging for error
		log.Printf("GetProblemCaseList2 error: %v\n", err)
		return nil, e.ErrUnknown
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
		// Add logging for error
		log.Printf("Execute error: %v\n", err)
		return nil, e.ErrUnknown
	}

	beginTime := time.Now()
	for _, c := range caseList {
		// 输入数据
		inputCh <- []byte(c.Input)

		// 读取输出数据
		select {
		case executeResult := <-outputCh:
			// 运行出错
			if !executeResult.Executed {
				submission.Status = constants.RuntimeError
				submission.ErrorMessage = executeResult.ErrorMessage
				return submission, nil
			}

			// 结果不正确则结束
			if !j.compareAnswer(string(executeResult.Output), c.Output) {
				submission.Status = constants.WrongAnswer
				submission.CaseName = c.Name
				submission.CaseData = c.Input
				submission.ExpectedOutput = c.Output
				submission.UserOutput = string(executeResult.Output)
				return submission, nil
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
func (j *judgeService) saveUserCode(language constants.LanguageType, codeStr string, executePath string) ([]string, *e.Error) {
	var compileFiles []string
	var mainFile string
	var err2 *e.Error

	if mainFile, err2 = getMainFileNameByLanguage(language); err2 != nil {
		log.Println(err2)
		return nil, err2
	}
	if err := os.WriteFile(path.Join(executePath, mainFile), []byte(codeStr), 0644); err != nil {
		log.Println(err)
		return nil, e.ErrServer
	}
	// 将main文件进行编译即可
	compileFiles = []string{path.Join(executePath, mainFile)}

	return compileFiles, nil
}

// 根据编程语言获取该编程语言的Main文件名称
func getMainFileNameByLanguage(language constants.LanguageType) (string, *e.Error) {
	switch language {
	case constants.LanguageC:
		return "main.c", nil
	case constants.LanguageJava:
		return "Main.java", nil
	case constants.LanguageGo:
		return "main.go", nil
	default:
		return "", e.ErrLanguageNotSupported
	}
}

func (j *judgeService) Execute(judgeRequest *dto.ExecuteRequestDto) (*dto.ExecuteResultDto, *e.Error) {
	executeResult := &dto.ExecuteResultDto{
		ProblemID: judgeRequest.ProblemID,
	}

	// executePath 用户执行目录
	executePath := getExecutePath(j.config)
	if err := os.MkdirAll(executePath, os.ModePerm); err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}
	defer os.RemoveAll(executePath)

	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err *e.Error
	if compileFiles, err = j.saveUserCode(judgeRequest.Language, judgeRequest.Code, executePath); err != nil {
		log.Println(err)
		return nil, err
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
	var err2 error
	if compileResult, err2 = j.judgeCore.Compile(compileFiles, executeFilePath, compileOptions); err2 != nil {
		log.Println(err2)
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
	if err2 = j.judgeCore.Execute(executeFilePath, inputCh, outputCh, exitCh, executeOptions); err2 != nil {
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
		log.Println(err)
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
			log.Println(err2)
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
		log.Println(err)
		return nil, e.ErrMysql
	}

	// 读取代码模板
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 读取题目
		problem, err := j.problemDao.GetProblemByID(global.Mysql, problemID)
		if err != nil {
			log.Println(err)
			return nil, e.ErrMysql
		}
		code, err2 := j.problemService.GetProblemTemplateCode(problemID,
			strings.Split(problem.Languages, ",")[0])
		if err2 != nil {
			log.Println(err2)
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
			log.Println(err)
			return err
		}
	}
	return nil
}
