package controllers

import (
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type SysUserController interface {
	// InsertSysUser 添加用户
	InsertSysUser(ctx *gin.Context)
	// UpdateSysUser 更新用户，但是不更新密码
	UpdateSysUser(ctx *gin.Context)
	// DeleteSysUser 删除用户
	DeleteSysUser(ctx *gin.Context)
	// GetSysUserList 获取用户列表
	GetSysUserList(ctx *gin.Context)
	// UpdateUserRoles 更新角色roleIDs
	UpdateUserRoles(ctx *gin.Context)
	// GetRoleIDsByUserID 通过用户id获取所有角色id
	GetRoleIDsByUserID(ctx *gin.Context)
}

type sysUserController struct {
	sysUserService service.SysUserService
}

func NewSysUserController() SysUserController {
	return &sysUserController{
		sysUserService: service.NewSysUserService(),
	}
}

func (s *sysUserController) InsertSysUser(ctx *gin.Context) {
	result := r.NewResult(ctx)
	loginName := ctx.PostForm("loginName")
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	email := ctx.PostForm("email")
	phone := ctx.PostForm("phone")
	id, err := s.sysUserService.InsertSysUser(&po.SysUser{
		LoginName: loginName,
		Username:  username,
		Password:  password,
		Email:     email,
		Phone:     phone,
	})
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(id)
}

func (s *sysUserController) UpdateSysUser(ctx *gin.Context) {
	result := r.NewResult(ctx)
	idStr := ctx.PostForm("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	loginName := ctx.PostForm("loginName")
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	email := ctx.PostForm("email")
	phone := ctx.PostForm("phone")
	user := &po.SysUser{}
	user.ID = uint(id)
	user.LoginName = loginName
	user.Username = username
	user.Password = password
	user.Email = email
	user.Phone = phone
	err2 := s.sysUserService.UpdateSysUser(user)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("更新成功")
}

func (s *sysUserController) DeleteSysUser(ctx *gin.Context) {
	result := r.NewResult(ctx)
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := s.sysUserService.DeleteSysUser(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("删除成功")
}

func (s *sysUserController) GetSysUserList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	username := ctx.Query("username")
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	var pageSize int
	pageSize, err = strconv.Atoi(pageSizeStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	pageInfo, err2 := s.sysUserService.GetSysUserList(username, page, pageSize)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(pageInfo)
}

type updateUserRolesRequest struct {
	UserID  uint   `json:"userID"`
	RoleIDs []uint `json:"roleIDs"`
}

func (s *sysUserController) UpdateUserRoles(ctx *gin.Context) {
	result := r.NewResult(ctx)
	var request updateUserRolesRequest
	err := ctx.BindJSON(&request)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := s.sysUserService.UpdateUserRoles(request.UserID, request.RoleIDs)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("更新成功")
}

func (s *sysUserController) GetRoleIDsByUserID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	roleIDs, err2 := s.sysUserService.GetRoleIDsByUserID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(roleIDs)
}
