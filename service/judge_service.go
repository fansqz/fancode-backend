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
		if problemAttempt.Status == 0 && submission.Status == constants.Accepted {
			problemAttempt.Status = 1
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
	if problemAttempt.Status == 0 && submission.Status == constants.Accepted {
		problemAttempt.Status = 1
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
		CodeType:  judgeRequest.CodeType,
		Language:  judgeRequest.Language,
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
	defer os.RemoveAll(executePath)

	// 保存题目文件的路径
	problemPath := getLocalProblemPath(problem.Path)
	localCodePath, err2 := getCodePathByProblemPath(problemPath, judgeRequest.Language)
	if err2 != nil {
		return nil, err2
	}

	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	compileFiles, err2 = j.saveUserCode(judgeRequest.Language,
		judgeRequest.CodeType, judgeRequest.Code, localCodePath, executePath)
	if err2 != nil {
		return nil, err2
	}

	// 输出执行文件路劲
	executeFilePath := path.Join(executePath, "main")

	// 执行编译
	err = j.judgeCore.Compile(judgeRequest.Language, compileFiles, executeFilePath, LimitCompileTime)
	if err != nil {
		submission.Status = constants.CompileError
		submission.ErrorMessage = err.Error()
		submission.Status = constants.CompileError
		submission.ErrorMessage = err.Error()
		return submission, nil
	}

	// 运行
	caseFilePath := getCasePathByLocalProblemPath(problemPath)
	files, err3 := os.ReadDir(caseFilePath)
	if err3 != nil {
		return nil, e.ErrServer
	}
	inputCh := make(chan []byte)
	outputCh := make(chan judger.ExecuteResult)
	exitCh := make(chan string)
	defer func() {
		exitCh <- "exit"
	}()
	executeOption := &judger.ExecuteOption{
		ExecFile:    executeFilePath,
		Language:    constants.ProgramC,
		InputCh:     inputCh,
		OutputCh:    outputCh,
		ExitCh:      exitCh,
		LimitTime:   LimitExecuteTime,
		LimitMemory: LimitExecuteMemory,
	}
	// 运行可执行文件
	err = j.judgeCore.Execute(executeOption)
	if err != nil {
		submission.Status = constants.RuntimeError
		submission.ExpectedOutput = err.Error()
		return submission, nil
	}

	beginTime := time.Now()
	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".in") {

			// 输入数据
			var input []byte
			input, err = os.ReadFile(path.Join(caseFilePath, fileInfo.Name()))
			if err != nil {
				log.Println(err)
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
			outFilePath := path.Join(caseFilePath, strings.ReplaceAll(fileInfo.Name(), ".in", ".out"))
			var outFileContent []byte
			outFileContent, err = os.ReadFile(outFilePath)
			if err != nil {
				log.Println(err)
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

// saveUserCode
// 保存用户代码到用户的executePath，并返回需要编译的文件列表
func (j *judgeService) saveUserCode(language string, codeType string, codeStr string, localCodePath string, executePath string) ([]string, *e.Error) {
	var compileFiles []string
	var mainFile string
	var solutionFile string
	var err *e.Error
	var err2 error
	if codeType == constants.CodeTypeCore {

		mainFile, err = getMainFileNameByLanguage(language)
		if err != nil {
			return nil, err
		}
		solutionFile, err = getSolutionFileNameByLanguage(language)
		if err != nil {
			return nil, err
		}

		// 用户代码加上上下文，写到code.c中
		var code []byte
		code, err2 = os.ReadFile(path.Join(localCodePath, solutionFile))
		if err2 != nil {
			return nil, e.ErrServer
		}
		re := regexp.MustCompile(`/\*begin\*/(?s).*/\*end\*/`)
		code = re.ReplaceAll(code, []byte(codeStr))
		err2 = os.WriteFile(path.Join(executePath, solutionFile), code, 0644)
		if err2 != nil {
			return nil, e.ErrServer
		}
		// 将main文件和solution文件一起编译
		compileFiles = []string{path.Join(localCodePath, mainFile), path.Join(executePath, solutionFile)}
	} else {
		// acm
		mainFile, err = getMainFileNameByLanguage(language)
		if err != nil {
			return nil, err
		}
		err2 = os.WriteFile(path.Join(executePath, mainFile), []byte(codeStr), 0644)
		if err2 != nil {
			return nil, e.ErrServer
		}
		// 将main文件进行编译即可
		compileFiles = []string{path.Join(executePath, mainFile)}
	}
	return compileFiles, nil
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
	defer os.RemoveAll(executePath)

	// 保存题目文件的目录
	problemPath := getLocalProblemPath(problem.Path)
	localCodePath, err2 := getCodePathByProblemPath(problemPath, judgeRequest.Language)
	if err2 != nil {
		return nil, err2
	}

	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	compileFiles, err2 = j.saveUserCode(judgeRequest.Language,
		judgeRequest.CodeType, judgeRequest.Code, localCodePath, executePath)
	if err2 != nil {
		return nil, err2
	}

	// 输出的执行文件路劲
	executeFilePath := path.Join(executePath, "main")

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
	defer func() {
		exitCh <- "exit"
	}()
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
	localPath := path.Join(global.Conf.FilePathConfig.ProblemFileDir, questionPath)
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

// 根据题目的相对路径，获取题目的本地路径
func getLocalProblemPath(p string) string {
	return path.Join(global.Conf.FilePathConfig.ProblemFileDir, p)
}

// 给用户的此次运行生成一个临时目录
func getExecutePath() string {
	uuid := utils.GetUUID()
	executePath := path.Join(global.Conf.FilePathConfig.TempDir, uuid)
	return executePath
}

const (
	/* 一道题目的结构如下：
	// problemFile:
	//	c     //保存c代码
	//	java // 保存java代码
	//	go    // 保存go代码
	//	io    //保存用例
	*/
	CCodePath    = "c"
	JavaCodePath = "java"
	GoCodePath   = "go"
	CaseFilePath = "io"
)

const (
	CMainFile        = "main.c"
	CSolutionFile    = "solution.c"
	JavaMainFile     = "Main.java"
	JavaSolutionFile = "Solution.java"
	GoMainFile       = "main.go"
	GoSolutionFile   = "solution.go"
)

// 根据题目的路径获取题目中编程语言的路径
func getCodePathByProblemPath(problemPath string, language string) (string, *e.Error) {
	switch language {
	case constants.ProgramC:
		return path.Join(problemPath, CCodePath), nil
	case constants.ProgramJava:
		return path.Join(problemPath, JavaCodePath), nil
	case constants.ProgramGo:
		return path.Join(problemPath, GoCodePath), nil
	default:
		return "", e.ErrLanguageNotSupported
	}
}

// 根据编程语言获取该编程语言的Main文件名称
func getMainFileNameByLanguage(language string) (string, *e.Error) {
	switch language {
	case constants.ProgramC:
		return CMainFile, nil
	case constants.ProgramJava:
		return JavaMainFile, nil
	case constants.ProgramGo:
		return GoMainFile, nil
	default:
		return "", e.ErrLanguageNotSupported
	}
}

// 根据编程语言获取该编程语言的Solution文件名称
func getSolutionFileNameByLanguage(language string) (string, *e.Error) {
	switch language {
	case constants.ProgramC:
		return CSolutionFile, nil
	case constants.ProgramJava:
		return JavaSolutionFile, nil
	case constants.ProgramGo:
		return GoSolutionFile, nil
	default:
		return "", e.ErrLanguageNotSupported
	}
}

// 根据题目的路径获取题目中用例的路径
func getCasePathByLocalProblemPath(localProblemPath string) string {
	return path.Join(localProblemPath, CaseFilePath)
}
