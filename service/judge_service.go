package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/models/dto"
	"FanCode/setting"
	"FanCode/utils"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type JudgeService interface {
	Submit(judgeRequest *dto.JudgingRequestDTO) (*dto.SubmitResultDTO, *e.Error)
	Execute(judgeRequest *dto.JudgingRequestDTO) (*dto.ExecuteResultDto, *e.Error)
}

type judgeService struct {
}

func NewJudgeService() JudgeService {
	return &judgeService{}
}

func (j *judgeService) Submit(judgeRequest *dto.JudgingRequestDTO) (*dto.SubmitResultDTO, *e.Error) {
	//读取题目
	_, err := dao.GetQuestionByQuestionID(judgeRequest.QuestionID)
	if err != nil {
		return nil, e.ErrSubmitFailed
	}
	return nil, nil
}

func (j *judgeService) Execute(judgeRequest *dto.JudgingRequestDTO) (*dto.ExecuteResultDto, *e.Error) {
	uuid := utils.GetUUID()
	//读取题目到本地，并编译
	question, err := dao.GetQuestionByQuestionID(judgeRequest.QuestionID)
	if err != nil {
		return nil, e.ErrExecuteFailed
	}
	err = checkAndDownloadQuestionFile(question.Path)
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
	err = os.WriteFile(executePath+"/user_code.c", []byte(judgeRequest.Code), 0644)
	if err != nil {
		log.Println(err)
		return nil, e.ErrExecuteFailed
	}
	// 执行编译
	localPath := setting.Conf.FilePathConfig.QuestionFileDir + "/" + question.Path
	cmd := exec.Command("gcc", "-o", executePath+"/main", localPath+"/main.c", executePath+"/user_code.c")
	err = cmd.Run()
	if err != nil {
		return &dto.ExecuteResultDto{
			ProblemId:    question.ID,
			Status:       CompileError,
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
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".out") {
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
			// 读取.in文件
			inFilePath := localPath + "/" + strings.ReplaceAll(fileInfo.Name(), ".out", ".in")
			inFileContent, err4 := os.ReadFile(inFilePath)
			if err4 != nil {
				log.Println(err4)
				return nil, e.ErrExecuteFailed
			}
			// 将输出结果与.in文件对比
			if bytes.Equal(cmd2.Stdout.(*bytes.Buffer).Bytes(), inFileContent) {
				continue
			} else {
				return &dto.ExecuteResultDto{
					ProblemId:    question.ID,
					Status:       AnswerError,
					ErrorMessage: "",
					Timestamp:    nil,
				}, nil
			}
		}
		if i > 3 {
			break
		}
	}
	return &dto.ExecuteResultDto{
		ProblemId:    question.ID,
		Status:       ExecuteSuccess,
		ErrorMessage: "",
		Timestamp:    nil,
	}, nil
}

func checkAndDownloadQuestionFile(questionPath string) error {
	localPath := setting.Conf.FilePathConfig.QuestionFileDir + "/" + questionPath
	if !checkFolderExists(localPath) {
		//拉取文件
		store := file_store.NewCOS()
		err := store.DownloadFolder(questionPath, localPath)
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
