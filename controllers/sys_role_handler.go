package controllers

import (
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

// SysRoleController 角色管理相关功能
type SysRoleController interface {
	// InsertSysRole 添加角色
	InsertSysRole(ctx *gin.Context)
	// DeleteSysRole 删除角色
	DeleteSysRole(ctx *gin.Context)
	// GetSysRoleList 获取角色列表
	GetSysRoleList(ctx *gin.Context)
	// UpdateSysRole 更新角色
	UpdateSysRole(ctx *gin.Context)
}

type sysRoleController struct {
	sysRoleService service.SysRoleService
}

func NewSysRoleController() SysRoleController {
	return &sysRoleController{
		sysRoleService: service.NewSysRoleService(),
	}
}

func (q *sysRoleController) InsertSysRole(ctx *gin.Context) {
	result := r.NewResult(ctx)
	sysRole := &po.SysRole{}
	sysRole.Name = ctx.PostForm("name")
	sysRole.Description = ctx.PostForm("description")
	//插入
	pID, err := q.sysRoleService.InsertSysRole(sysRole)
	if err != nil {
		result.Error(err)
		return
	}
	result.Success("角色添加成功", pID)
}

func (q *sysRoleController) UpdateSysRole(ctx *gin.Context) {
	result := r.NewResult(ctx)
	sysRoleIDString := ctx.PostForm("id")
	sysRoleID, err := strconv.Atoi(sysRoleIDString)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	sysRole := &po.SysRole{}
	sysRole.ID = uint(sysRoleID)
	sysRole.Name = ctx.PostForm("name")
	sysRole.Description = ctx.PostForm("description")
	err2 := q.sysRoleService.UpdateSysRole(sysRole)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("角色修改成功")
}

func (q *sysRoleController) DeleteSysRole(ctx *gin.Context) {
	result := r.NewResult(ctx)
	ids := ctx.Param("id")
	id, convertErr := strconv.Atoi(ids)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := q.sysRoleService.DeleteSysRole(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("删除成功")
}

// GetSysRoleList 读取一个列表的角色
func (q *sysRoleController) GetSysRoleList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageStr := ctx.Param("page")
	pageSizeStr := ctx.Param("pageSize")
	var page int
	var pageSize int
	var convertErr error
	page, convertErr = strconv.Atoi(pageStr)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	pageSize, convertErr = strconv.Atoi(pageSizeStr)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	pageInfo, err := q.sysRoleService.GetSysRoleList(page, pageSize)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}
