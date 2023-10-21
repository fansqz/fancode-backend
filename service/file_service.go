package service

import (
	conf "FanCode/config"
	e "FanCode/error"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
)

// FileService 文件上传相关service
type FileService interface {
	// StartUpload 启动上传命令
	StartUpload() (string, *e.Error)
	// Upload 上传分片
	Upload(path string, ctx *gin.Context, file *multipart.FileHeader) *e.Error
	// CheckChunkSet 检测分片的文件名称集合
	CheckChunkSet(path string) ([]string, *e.Error)
	// CancelUpload 取消上传
	CancelUpload(path string) *e.Error
	// CompleteUpload 完成大文件上传功能
	CompleteUpload(path string, fileName string, hash string, hashType string) *e.Error
}

type fileService struct {
	config *conf.AppConfig
}

func NewFileService(config *conf.AppConfig) FileService {
	return &fileService{
		config: config,
	}
}

func (f *fileService) StartUpload() (string, *e.Error) {
	tempPath := getTempDir(f.config)
	err := os.MkdirAll(tempPath, 0755)
	if err != nil {
		return "", e.ErrServer
	}
	return tempPath, nil
}

func (f *fileService) Upload(path string, ctx *gin.Context, file *multipart.FileHeader) *e.Error {
	err := ctx.SaveUploadedFile(file, path)
	if err != nil {
		return e.ErrServer
	}
	return nil
}

func (f *fileService) CheckChunkSet(path string) ([]string, *e.Error) {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return nil, e.ErrServer
	}
	answer := make([]string, len(dirs))
	for i, a := range dirs {
		answer[i] = a.Name()
	}
	return answer, nil
}

// CancelUpload 取消上传
func (f *fileService) CancelUpload(path string) *e.Error {
	err := os.RemoveAll(path)
	if err != nil {
		return e.ErrServer
	}
	return nil
}

// CompleteUpload 完成大文件上传功能
func (f *fileService) CompleteUpload(p string, fileName string, h string, hashType string) *e.Error {
	// 读取path中的所有文件
	files, err := ioutil.ReadDir(p)
	if err != nil {
		// 处理错误
		return e.ErrServer
	}

	// 创建结果文件
	resultFile, err := os.Create(fileName)
	if err != nil {
		// 处理错误
		return e.ErrServer
	}
	defer resultFile.Close()

	// 遍历所有文件，逐个写入结果文件
	for _, file := range files {
		filePath := path.Join(p, file.Name())
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			// 处理错误
			return e.ErrServer
		}
		resultFile.Write(fileData)
	}

	hash2, err2 := hash(p, hashType)
	if err2 != nil {
		return err2
	}
	if hash2 != h {
		return e.ErrHashMissMatch
	}
	return nil
}

func hash(filePath string, hashType string) (string, *e.Error) {
	// 计算结果文件的哈希值，并与传入的哈希值进行比较
	switch hashType {
	case "md5":
		resultHash, err := calculateMD5(filePath)
		if err != nil {
			return "", e.ErrServer
		}
		return resultHash, nil
	case "sha1":
		resultHash, err := calculateSHA1(filePath)
		if err != nil {
			return "", e.ErrServer
		}
		return resultHash, nil
	case "sha256":
		resultHash, err := calculateSHA256(filePath)
		if err != nil {
			return "", e.ErrServer
		}
		return resultHash, nil
	default:
		// 不支持的哈希算法类型，处理错误
		return "", e.ErrHashTypeNotSupport
	}
}

// 计算MD5哈希值
func calculateMD5(filePath string) (string, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		// 处理错误
		return "", err
	}
	hash := md5.Sum(fileData)
	return hex.EncodeToString(hash[:]), nil
}

// 计算SHA1哈希值
func calculateSHA1(filePath string) (string, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		// 处理错误
		return "", err
	}
	hash := sha1.Sum(fileData)
	return hex.EncodeToString(hash[:]), nil
}

// 计算SHA256哈希值
func calculateSHA256(filePath string) (string, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		// 处理错误
		return "", err
	}
	hash := sha256.Sum256(fileData)
	return hex.EncodeToString(hash[:]), nil
}
