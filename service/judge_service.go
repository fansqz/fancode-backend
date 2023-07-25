package service

import (
	"FanCode/constants"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/setting"
	"FanCode/utils"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type JudgeService interface {
	// 答案提交
	Submit(ctx *gin.Context, judgeRequest *dto.JudgingRequestDTO) (*dto.SubmitResultDTO, *e.Error)
	// 执行
	Execute(judgeRequest *dto.JudgingRequestDTO) (*dto.ExecuteResultDto, *e.Error)
}

type judgeService struct {
}

func NewJudgeService() JudgeService {
	return &judgeService{}
}

func (j *judgeService) Submit(ctx *gin.Context, judgeRequest *dto.JudgingRequestDTO) (*dto.SubmitResultDTO, *e.Error) {
	uuid := utils.GetUUID()
	// 提交结果对象
	submission := &po.Submission{
		Code:      judgeRequest.Code,
		ProblemID: judgeRequest.ProblemID,
		UserID:    ctx.Keys["user"].(*po.User).ID,
	}
	//读取题目到本地，并编译
	problem, err := dao.GetProblemByProblemID(judgeRequest.ProblemID)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	err = checkAndDownloadQuestionFile(problem.Path)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	// executePath
	executePath := setting.Conf.FilePathConfig.TempDir + "/" + uuid
	err = os.MkdirAll(executePath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}
	// 保存code文件
	localPath := setting.Conf.FilePathConfig.ProblemFileDir + "/" + problem.Path
	var code []byte
	code, err = os.ReadFile(localPath + "/code")
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
		localPath+"/main.c", executePath+"/code.c")
	err = cmd.Run()
	if err != nil {
		submission.Status = constants.CompileError
		submission.ErrorMessage = err.Error()
		_ = dao.InsertSubmission(submission)
		return &dto.SubmitResultDTO{
			ProblemID:    problem.ID,
			Status:       constants.CompileError,
			ErrorMessage: err.Error(),
			Timestamp:    nil,
		}, nil
	}
	// 运行
	files, err2 := os.ReadDir(localPath)
	if err2 != nil {
		submission.Status = constants.RuntimeError
		submission.ErrorMessage = err2.Error()
		_ = dao.InsertSubmission(submission)
		return &dto.SubmitResultDTO{
			ProblemID:    problem.ID,
			Status:       constants.RuntimeError,
			ErrorMessage: err2.Error(),
			Timestamp:    nil,
		}, nil
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
			if bytes.Equal(cmd2.Stdout.(*bytes.Buffer).Bytes(), outFileContent) {
				continue
			} else {
				submission.Status = constants.WrongAnswer
				_ = dao.InsertSubmission(submission)
				return &dto.SubmitResultDTO{
					ProblemID:      problem.ID,
					Status:         constants.WrongAnswer,
					ErrorMessage:   "",
					ExpectedOutput: string(outFileContent),
					UserOutput:     string(cmd2.Stdout.(*bytes.Buffer).Bytes()),
					Timestamp:      nil,
				}, nil
			}
			// 释放buffer
			cmd2.Stdout.(*bytes.Buffer).Reset()
		}
	}
	submission.Status = constants.Accepted
	_ = dao.InsertSubmission(submission)
	return &dto.SubmitResultDTO{
		ProblemID:    problem.ID,
		Status:       constants.Accepted,
		ErrorMessage: "",
		Timestamp:    nil,
	}, nil
}

func (j *judgeService) Execute(judgeRequest *dto.JudgingRequestDTO) (*dto.ExecuteResultDto, *e.Error) {
	uuid := utils.GetUUID()
	//读取题目到本地，并编译
	problem, err := dao.GetProblemByProblemID(judgeRequest.ProblemID)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	err = checkAndDownloadQuestionFile(problem.Path)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	// executePath
	executePath := setting.Conf.FilePathConfig.TempDir + "/" + uuid
	err = os.MkdirAll(executePath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}
	// 保存code文件
	localPath := setting.Conf.FilePathConfig.ProblemFileDir + "/" + problem.Path
	var code []byte
	code, err = os.ReadFile(localPath + "/code")
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
		localPath+"/main.c", executePath+"/code.c")
	err = cmd.Run()
	if err != nil {
		return &dto.ExecuteResultDto{
			ProblemID:    problem.ID,
			Status:       constants.CompileError,
			ErrorMessage: err.Error(),
			Timestamp:    nil,
		}, nil
	}
	// 运行
	files, err2 := os.ReadDir(localPath)
	if err2 != nil {
		return nil, e.ErrExecuteFailed
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
			if bytes.Equal(cmd2.Stdout.(*bytes.Buffer).Bytes(), outFileContent) {
				continue
			} else {
				return &dto.ExecuteResultDto{
					ProblemID:      problem.ID,
					Status:         constants.WrongAnswer,
					ErrorMessage:   "",
					ExpectedOutput: string(outFileContent),
					UserOutput:     string(cmd2.Stdout.(*bytes.Buffer).Bytes()),
					Timestamp:      nil,
				}, nil
			}
			// 释放buffer
			cmd2.Stdout.(*bytes.Buffer).Reset()
		}
		if i > 3 {
			break
		}
	}
	return &dto.ExecuteResultDto{
		ProblemID:    problem.ID,
		Status:       constants.Accepted,
		ErrorMessage: "",
		Timestamp:    nil,
	}, nil
}

func checkAndDownloadQuestionFile(questionPath string) error {
	localPath := setting.Conf.FilePathConfig.ProblemFileDir + "/" + questionPath
	if !checkFolderExists(localPath) {
		// 拉取文件
		store := file_store.NewCOS()
		err := store.DownloadFolder(questionPath, localPath)
		if err != nil {
			return err
		}
		// 将code.c改为code
		err = os.Rename(localPath+"/code.c", localPath+"/code")
		if err != nil {
			return err
		}
	}
	return nil
}

func checkFolderExists(folderPath string) bool {
	fileInfo, err := os.Stat(folderPath)
	if err == nil && fileInfo.IsDir() {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		fmt.Println("发生错误:", err)
		return false
	}
}
