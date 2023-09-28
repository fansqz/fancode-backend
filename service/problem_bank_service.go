package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"log"
	"mime/multipart"
	"path"
)

const (
	// ProblemBankIconPath cos中，题库图标存储位置
	ProblemBankIconPath = "/icon/problemBank"
)

// ProblemBankService 题库管理的service
type ProblemBankService interface {
	// UploadProblemBankIcon 上传题库图标
	UploadProblemBankIcon(file *multipart.FileHeader) (string, *e.Error)
	// ReadProblemBankIcon 读取题库图标
	ReadProblemBankIcon(ctx *gin.Context, iconName string)
	// InsertProblemBank 添加题库
	InsertProblemBank(problemBank *po.ProblemBank, ctx *gin.Context) (uint, *e.Error)
	// UpdateProblemBank 更新题库
	UpdateProblemBank(problemBank *po.ProblemBank) *e.Error
	// DeleteProblemBank 删除题库
	DeleteProblemBank(id uint, forceDelete bool) *e.Error
	// GetProblemBankList 获取题目列表
	GetProblemBankList(query *dto.PageQuery) (*dto.PageInfo, *e.Error)
	// GetProblemBankByID 获取题目信息
	GetProblemBankByID(id uint) (*po.ProblemBank, *e.Error)
}

type problemBankService struct {
}

func NewProblemBankService() ProblemBankService {
	return &problemBankService{}
}

func (p *problemBankService) UploadProblemBankIcon(file *multipart.FileHeader) (string, *e.Error) {
	cos := file_store.NewImageCOS()
	fileName := file.Filename
	fileName = utils.GetUUID() + "." + path.Base(fileName)
	file2, err := file.Open()
	if err != nil {
		return "", e.ErrBadRequest
	}
	err = cos.SaveFile(path.Join(ProblemBankIconPath, fileName), file2)
	if err != nil {
		log.Println(err)
		return "", e.ErrServer
	}
	return global.Conf.ProUrl + path.Join("/manage/problemBank/icon", fileName), nil
}

func (p *problemBankService) ReadProblemBankIcon(ctx *gin.Context, iconName string) {
	result := r.NewResult(ctx)
	cos := file_store.NewImageCOS()
	bytes, err := cos.ReadFile(path.Join(ProblemBankIconPath, iconName))
	if err != nil {
		result.Error(e.ErrServer)
		return
	}
	_, _ = ctx.Writer.Write(bytes)
}

func (p *problemBankService) InsertProblemBank(problemBank *po.ProblemBank, ctx *gin.Context) (uint, *e.Error) {
	// 对设置值的数据设置默认值
	if problemBank.Name == "" {
		problemBank.Name = "未命名题库"
	}
	if problemBank.Description == "" {
		problemBank.Description = "无描述信息"
	}
	problemBank.CreatorID = ctx.Keys["user"].(*dto.UserInfo).ID
	err := dao.InsertProblemBank(global.Mysql, problemBank)
	if err != nil {
		return 0, e.ErrMysql
	}
	return problemBank.ID, nil
}

func (p *problemBankService) UpdateProblemBank(problemBank *po.ProblemBank) *e.Error {
	err := dao.UpdateProblemBank(global.Mysql, problemBank)
	if err != nil {
		return e.ErrMysql
	}
	return nil
}

func (p *problemBankService) DeleteProblemBank(id uint, forceDelete bool) *e.Error {
	var err error
	// 非强制删除
	if !forceDelete {
		var count int64
		count, err = dao.GetProblemCount(global.Mysql, &po.Problem{
			BankID: id,
		})
		if count != 0 {
			return e.NewCustomMsg("题库不为空，请问是否需要强制删除")
		}
		err = dao.DeleteProblemBankByID(global.Mysql, id)
		if err != nil {
			return e.ErrMysql
		}
		return nil
	}

	// 强制删除
	err = dao.DeleteProblemBankByID(global.Mysql, id)
	if err != nil {
		return e.ErrMysql
	}
	return nil
}

func (p *problemBankService) GetProblemBankList(query *dto.PageQuery) (*dto.PageInfo, *e.Error) {
	var bankQuery *po.ProblemBank
	if query.Query != nil {
		bankQuery = query.Query.(*po.ProblemBank)
	}
	// 获取题库列表
	banks, err := dao.GetProblemBankList(global.Mysql, query)
	if err != nil {
		return nil, e.ErrMysql
	}
	newProblemBanks := make([]*dto.ProblemBankDtoForList, len(banks))
	for i := 0; i < len(banks); i++ {
		newProblemBanks[i] = dto.NewProblemBankDtoForList(banks[i])
		// 读取题库中的题目总数还有作者
		newProblemBanks[i].ProblemCount, err = dao.GetProblemCount(global.Mysql, &po.Problem{
			BankID: newProblemBanks[i].ID,
		})
		if err != nil {
			return nil, e.ErrMysql
		}
		newProblemBanks[i].CreatorName, err = dao.GetUserNameByID(global.Mysql, banks[i].CreatorID)
	}
	// 获取所有题库总数目
	var count int64
	count, err = dao.GetProblemBankCount(global.Mysql, bankQuery)
	if err != nil {
		return nil, e.ErrMysql
	}
	pageInfo := &dto.PageInfo{
		Total: count,
		Size:  int64(len(newProblemBanks)),
		List:  newProblemBanks,
	}
	return pageInfo, nil
}

func (p *problemBankService) GetProblemBankByID(id uint) (*po.ProblemBank, *e.Error) {
	bank, err := dao.GetProblemBankByID(global.Mysql, id)
	if err != nil {
		return nil, e.ErrMysql
	}
	return bank, nil
}
