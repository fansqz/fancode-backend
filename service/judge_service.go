package service

import (
	"FanCode/constants"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"bytes"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type JudgeService interface {
	// Submit 答案提交
	Submit(ctx *gin.Context, judgeRequest *dto.SubmitRequestDto) (*dto.SubmitResultDto, *e.Error)
	// Execute 执行
	Execute(judgeRequest *dto.ExecuteRequestDto) (*dto.ExecuteResultDto, *e.Error)
}

type judgeService struct {
}

func NewJudgeService() JudgeService {
	return &judgeService{}
}

func (j *judgeService) Submit(ctx *gin.Context, judgeRequest *dto.SubmitRequestDto) (*dto.SubmitResultDto, *e.Error) {
	submission, err := j.submit(ctx, judgeRequest)
	if err != nil {
		return nil, err
	}
	tx := global.Mysql.Begin()
	_ = dao.InsertSubmission(tx, submission)
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
	uuid := utils.GetUUID()
	// 提交结果对象
	submission := &po.Submission{
		Code:      judgeRequest.Code,
		ProblemID: judgeRequest.ProblemID,
		UserID:    ctx.Keys["user"].(*dto.UserInfo).ID,
	}
	//读取题目到本地，并编译
	problem, err := dao.GetProblemByID(global.Mysql, judgeRequest.ProblemID)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	err = checkAndDownloadQuestionFile(problem.Path)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	// executePath
	executePath := global.Conf.FilePathConfig.TempDir + "/" + uuid
	err = os.MkdirAll(executePath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}
	// 保存code文件
	localPath := global.Conf.FilePathConfig.ProblemFileDir + "/" + problem.Path
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
	// 执行编译
	cmd := exec.Command("gcc", "-o", executePath+"/main",
		localPath+"/test_execute.c", executePath+"/code.c")
	err = cmd.Run()
	if err != nil {
		submission.Status = constants.CompileError
		submission.ErrorMessage = err.Error()
		submission.Status = constants.CompileError
		submission.ErrorMessage = err.Error()
		return submission, nil
	}
	// 运行
	files, err2 := os.ReadDir(localPath)
	if err2 != nil {
		submission.Status = constants.RuntimeError
		submission.ErrorMessage = err2.Error()
		_ = dao.InsertSubmission(global.Mysql, submission)
		submission.Status = constants.RuntimeError
		submission.ErrorMessage = err2.Error()
		return submission, nil
	}
	i := 0
	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".in") {
			i++
			input, err3 := os.Open(localPath + "/" + fileInfo.Name())
			if err3 != nil {
				log.Println(err3)
				return nil, e.ErrExecuteFailed
			}
			//执行
			cmd2 := exec.Command(executePath + "/main")
			cmd2.Stdin = input
			cmd2.Stdout = &bytes.Buffer{}
			err = cmd2.Run()
			if err != nil {
				log.Println(err)
				return nil, e.ErrExecuteFailed
			}
			// 读取.out文件
			outFilePath := localPath + "/" + strings.ReplaceAll(fileInfo.Name(), ".in", ".out")
			outFileContent, err4 := os.ReadFile(outFilePath)
			if err4 != nil {
				log.Println(err4)
				return nil, e.ErrExecuteFailed
			}
			// 将输出结果与.out文件对比
			if !bytes.Equal(cmd2.Stdout.(*bytes.Buffer).Bytes(), outFileContent) {
				submission.Status = constants.WrongAnswer
				submission.ExpectedOutput = string(outFileContent)
				submission.UserOutput = string(cmd2.Stdout.(*bytes.Buffer).Bytes())
				cmd2.Stdout.(*bytes.Buffer).Reset()
				return submission, nil
			}
			// 释放buffer
			cmd2.Stdout.(*bytes.Buffer).Reset()
		}
	}
	submission.Status = constants.Accepted
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
	// executePath
	executePath := getExecutePath()
	err = os.MkdirAll(executePath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}
	// 保存code文件
	localPath := getLocalPathByPath(problem.Path)
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
	// 执行编译
	cmd := exec.Command("gcc", "-o", executePath+"/main",
		localPath+"/test_execute.c", executePath+"/code.c")
	err = cmd.Run()
	if err != nil {
		return &dto.ExecuteResultDto{
			ProblemID:    problem.ID,
			Status:       constants.CompileError,
			ErrorMessage: err.Error(),
			Timestamp:    nil,
		}, nil
	}
	//执行
	cmd2 := exec.Command(executePath + "/main")
	cmd2.Stdin = strings.NewReader(judgeRequest.Input)
	cmd2.Stdout = &bytes.Buffer{}
	err = cmd2.Run()
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}
	output := cmd2.Stdout.(*bytes.Buffer).Bytes()
	cmd2.Stdout.(*bytes.Buffer).Reset()
	return &dto.ExecuteResultDto{
		ProblemID:    problem.ID,
		Status:       constants.RunSuccess,
		ErrorMessage: "",
		UserOutput:   string(output),
		Timestamp:    nil,
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
