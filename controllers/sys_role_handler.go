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
	// InsertApisToRole 给角色添加api
	InsertApisToRole(ctx *gin.Context)
	// InsertMenusToRole 给角色添加menu
	InsertMenusToRole(ctx *gin.Context)
	// GetApiIDsByRoleID 通过角色id获取该角色拥有的apiID
	GetApiIDsByRoleID(ctx *gin.Context)
	// GetMenuIDsByRoleID 通过角色id获取该角色拥有的menuID
	GetMenuIDsByRoleID(ctx *gin.Context)
}

type sysRoleController struct {
	sysRoleService service.SysRoleService
}

func NewSysRoleController() SysRoleController {
	return &sysRoleController{
		sysRoleService: service.NewSysRoleService(),
	}
}

func (s *sysRoleController) InsertSysRole(ctx *gin.Context) {
	result := r.NewResult(ctx)
	sysRole := &po.SysRole{}
	sysRole.Name = ctx.PostForm("name")
	sysRole.Description = ctx.PostForm("description")
	//插入
	pID, err := s.sysRoleService.InsertSysRole(sysRole)
	if err != nil {
		result.Error(err)
		return
	}
	result.Success("角色添加成功", pID)
}

func (s *sysRoleController) UpdateSysRole(ctx *gin.Context) {
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
	err2 := s.sysRoleService.UpdateSysRole(sysRole)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("角色修改成功")
}

func (s *sysRoleController) DeleteSysRole(ctx *gin.Context) {
	result := r.NewResult(ctx)
	ids := ctx.Param("id")
	id, convertErr := strconv.Atoi(ids)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := s.sysRoleService.DeleteSysRole(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("删除成功")
}

// GetSysRoleList 读取一个列表的角色
func (s *sysRoleController) GetSysRoleList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
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
	roleName := ctx.Query("roleName")
	pageInfo, err := s.sysRoleService.GetSysRoleList(roleName, page, pageSize)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}

type insertApisToRoleRequest struct {
	RoleID uint   `json:"roleID"`
	ApiIDs []uint `json:"apiIDs"`
}

func (s *sysRoleController) InsertApisToRole(ctx *gin.Context) {
	result := r.NewResult(ctx)
	var json insertApisToRoleRequest
	err := ctx.BindJSON(&json)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := s.sysRoleService.InsertApisToRole(json.RoleID, json.ApiIDs)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("添加成功")
}

type insertMenusToRoleRequest struct {
	RoleID  uint   `json:"roleID"`
	MenuIDs []uint `json:"menuIDs"`
}

func (s *sysRoleController) InsertMenusToRole(ctx *gin.Context) {
	result := r.NewResult(ctx)
	var json insertMenusToRoleRequest
	err := ctx.BindJSON(&json)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := s.sysRoleService.InsertMenusToRole(json.RoleID, json.MenuIDs)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("添加成功")
}

func (s *sysRoleController) GetApiIDsByRoleID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	apiIDs, err2 := s.sysRoleService.GetApiIDsByRoleID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(apiIDs)
}

func (s *sysRoleController) GetMenuIDsByRoleID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	menuIDs, err2 := s.sysRoleService.GetMenuIDsByRoleID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(menuIDs)
}
