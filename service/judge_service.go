package service

import (
	"FanCode/constants"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/service/judger"
	"FanCode/utils"
	"bytes"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	time "time"
)

const (
	// 限制时间和内存
	LimitExecuteTime   = 15 * time.Second
	LimitExecuteMemory = 20 * 1024 * 1024
	// 限制编译时间
	LimitCompileTime = 10 * time.Second
)

type JudgeService interface {
	// Submit 答案提交
	Submit(ctx *gin.Context, judgeRequest *dto.SubmitRequestDto) (*dto.SubmitResultDto, *e.Error)
	// Execute 执行
	Execute(judgeRequest *dto.ExecuteRequestDto) (*dto.ExecuteResultDto, *e.Error)
}

type judgeService struct {
	judgeCore *judger.JudgeCore
}

func NewJudgeService() JudgeService {
	return &judgeService{
		judgeCore: judger.NewJudgeCore(),
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
	_ = dao.InsertSubmission(tx, submission)

	// 检测用户是否保存了attempt
	userId := ctx.Keys["user"].(*dto.UserInfo).ID
	problemAttempt, err2 := dao.GetProblemAttempt(tx, userId, judgeRequest.ProblemID)
	if err2 != nil {
		log.Println(err2)
		return nil, e.ErrSubmitFailed
	}

	// 如果本身就没有记录，就插入
	if problemAttempt.ID == 0 {
		problemAttempt = &po.ProblemAttempt{
			UserID:    userId,
			ProblemID: judgeRequest.ProblemID,
		}
		problemAttempt.SubmissionCount++
		if submission.Status == constants.Accepted {
			problemAttempt.SuccessCount++
		} else {
			problemAttempt.ErrCount++
		}
		problemAttempt.Code = judgeRequest.Code
		if problemAttempt.State == 0 && submission.Status == constants.Accepted {
			problemAttempt.State = 1
		}
		err2 = dao.InsertProblemAttempt(tx, problemAttempt)
		if err2 != nil {
			log.Println(err2)
			tx.Rollback()
			return nil, e.ErrSubmitFailed
		}
		tx.Commit()
		return dto.NewSubmitResultDto(submission), nil
	}

	// 有记录则更新
	problemAttempt.SubmissionCount++
	if submission.Status == constants.Accepted {
		problemAttempt.SuccessCount++
	} else {
		problemAttempt.ErrCount++
	}
	problemAttempt.Code = judgeRequest.Code
	if problemAttempt.State == 0 && submission.Status == constants.Accepted {
		problemAttempt.State = 1
	}
	err2 = dao.UpdateProblemAttempt(tx, problemAttempt)
	if err2 != nil {
		log.Println(err2)
		tx.Rollback()
		return nil, e.ErrSubmitFailed
	}
	tx.Commit()
	return dto.NewSubmitResultDto(submission), nil
}

func (j *judgeService) submit(ctx *gin.Context, judgeRequest *dto.SubmitRequestDto) (*po.Submission, *e.Error) {

	// 提交结果对象
	submission := &po.Submission{
		Code:      judgeRequest.Code,
		ProblemID: judgeRequest.ProblemID,
		UserID:    ctx.Keys["user"].(*dto.UserInfo).ID,
	}

	//读取题目到本地
	problem, err := dao.GetProblemByID(global.Mysql, judgeRequest.ProblemID)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	err = checkAndDownloadQuestionFile(problem.Path)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}

	// executePath 执行路径，用户的临时文件
	executePath := getExecutePath()
	err = os.MkdirAll(executePath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}

	// 保存题目文件的路径
	localPath := getLocalPathByPath(problem.Path)

	// 用户代码加上上下文，写到code.c中
	var code []byte
	code, err = os.ReadFile(localPath + "/code.c")
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}
	re := regexp.MustCompile(`/\*begin\*/(?s).*/\*end\*/`)
	code = re.ReplaceAll(code, []byte(judgeRequest.Code))
	err = os.WriteFile(executePath+"/code.c", code, 0644)
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}

	// 编译的文件
	compileFiles := []string{localPath + "/main.c", executePath + "/code.c"}
	// 输出的执行文件路劲
	executeFilePath := executePath + "/main"

	// 执行编译
	err = j.judgeCore.Compile(constants.ProgramC, compileFiles, executeFilePath, LimitCompileTime)
	if err != nil {
		submission.Status = constants.CompileError
		submission.ErrorMessage = err.Error()
		submission.Status = constants.CompileError
		submission.ErrorMessage = err.Error()
		return submission, nil
	}

	// 运行
	files, err2 := os.ReadDir(path.Join(localPath, "io"))
	if err2 != nil {
		submission.Status = constants.RuntimeError
		submission.ErrorMessage = err2.Error()
		_ = dao.InsertSubmission(global.Mysql, submission)
		submission.Status = constants.RuntimeError
		submission.ErrorMessage = err2.Error()
		return submission, nil
	}
	inputCh := make(chan []byte)
	outputCh := make(chan judger.ExecuteResult)
	exitCh := make(chan string)
	executeOption := &judger.ExecuteOption{
		ExecFile:    executeFilePath,
		Language:    constants.ProgramC,
		InputCh:     inputCh,
		OutputCh:    outputCh,
		ExitCh:      exitCh,
		LimitTime:   LimitExecuteTime,
		LimitMemory: LimitExecuteMemory,
	}

	beginTime := time.Now()
	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".in") {
			// 运行可执行文件
			err = j.judgeCore.Execute(executeOption)
			if err != nil {
				submission.Status = constants.RuntimeError
				submission.ExpectedOutput = err.Error()
				return submission, nil
			}

			// 输入数据
			input, err3 := os.ReadFile(localPath + "/" + fileInfo.Name())
			if err3 != nil {
				log.Println(err3)
				return nil, e.ErrExecuteFailed
			}
			inputCh <- input

			// 读取输出数据
			executeResult := <-outputCh

			// 运行出错
			if !executeResult.Executed {
				submission.Status = constants.RuntimeError
				submission.ErrorMessage = executeResult.Error.Error()
				return submission, nil
			}

			// 读取.out文件
			outFilePath := localPath + "/" + strings.ReplaceAll(fileInfo.Name(), ".in", ".out")
			outFileContent, err4 := os.ReadFile(outFilePath)
			if err4 != nil {
				log.Println(err4)
				return nil, e.ErrExecuteFailed
			}

			// 结果不正确则结束
			if !bytes.Equal(executeResult.Output, outFileContent) {
				submission.Status = constants.WrongAnswer
				submission.ExpectedOutput = string(outFileContent)
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

func (j *judgeService) Execute(judgeRequest *dto.ExecuteRequestDto) (*dto.ExecuteResultDto, *e.Error) {

	//读取题目到本地，并编译
	problem, err := dao.GetProblemByID(global.Mysql, judgeRequest.ProblemID)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	err = checkAndDownloadQuestionFile(problem.Path)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}

	// executePath 用户执行目录
	executePath := getExecutePath()
	err = os.MkdirAll(executePath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}

	// 保存题目文件的目录
	localPath := getLocalPathByPath(problem.Path)

	// 读取用户输入文件
	var code []byte
	code, err = os.ReadFile(localPath + "/code.c")
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}
	re := regexp.MustCompile(`/\*begin\*/(?s).*/\*end\*/`)
	code = re.ReplaceAll(code, []byte(judgeRequest.Code))

	// 使用空格替换所有非单词字符
	err = os.WriteFile(executePath+"/code.c", code, 0644)
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}

	// 编译的文件
	compileFiles := []string{localPath + "/main.c", executePath + "/code.c"}
	// 输出的执行文件路劲
	executeFilePath := executePath + "/main"

	// 执行编译
	err = j.judgeCore.Compile(constants.ProgramC, compileFiles, executeFilePath, LimitCompileTime)
	if err != nil {
		return &dto.ExecuteResultDto{
			ProblemID:    problem.ID,
			Status:       constants.CompileError,
			ErrorMessage: err.Error(),
		}, nil
	}

	//执行
	inputCh := make(chan []byte)
	outputCh := make(chan judger.ExecuteResult)
	exitCh := make(chan string)
	executeOption := &judger.ExecuteOption{
		ExecFile:    executeFilePath,
		Language:    constants.ProgramC,
		InputCh:     inputCh,
		OutputCh:    outputCh,
		ExitCh:      exitCh,
		LimitTime:   LimitExecuteTime,
		LimitMemory: LimitExecuteMemory,
	}

	err = j.judgeCore.Execute(executeOption)
	if err != nil {
		return &dto.ExecuteResultDto{
			ProblemID:    problem.ID,
			Status:       constants.RuntimeError,
			ErrorMessage: err.Error(),
		}, nil
	}

	inputCh <- []byte(judgeRequest.Input)
	output := <-outputCh

	if !output.Executed {
		return &dto.ExecuteResultDto{
			ProblemID:    problem.ID,
			Status:       constants.RuntimeError,
			ErrorMessage: output.Error.Error(),
		}, nil
	}

	return &dto.ExecuteResultDto{
		ProblemID:    problem.ID,
		Status:       constants.RunSuccess,
		ErrorMessage: "",
		UserOutput:   string(output.Output),
	}, nil
}

func checkAndDownloadQuestionFile(questionPath string) error {
	localPath := global.Conf.FilePathConfig.ProblemFileDir + "/" + questionPath
	if !utils.CheckFolderExists(localPath) {
		// 拉取文件
		store := file_store.NewProblemCOS()
		err := store.DownloadFolder(questionPath, localPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func getLocalPathByPath(path string) string {
	return global.Conf.FilePathConfig.ProblemFileDir + "/" + path
}

func getExecutePath() string {
	uuid := utils.GetUUID()
	executePath := global.Conf.FilePathConfig.TempDir + "/" + uuid
	return executePath
}
