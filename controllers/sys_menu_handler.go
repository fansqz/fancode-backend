package controllers

import (
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type SysMenuController interface {
	// GetMenuCount 获取menu数目
	GetMenuCount(ctx *gin.Context)
	// DeleteMenuByID 删除menu
	DeleteMenuByID(ctx *gin.Context)
	// UpdateMenu 更新menu
	UpdateMenu(ctx *gin.Context)
	// GetMenuByID 根据id获取menu
	GetMenuByID(ctx *gin.Context)
	// GetMenuTree 获取menu树
	GetMenuTree(ctx *gin.Context)
	// InsertMenu 添加menu
	InsertMenu(ctx *gin.Context)
}

type sysMenuController struct {
	sysMenuService service.SysMenuService
}

func NewSysMenuController() SysMenuController {
	return &sysMenuController{
		sysMenuService: service.NewSysMenuService(),
	}
}

func (s *sysMenuController) GetMenuCount(ctx *gin.Context) {
	result := r.NewResult(ctx)
	count, err := s.sysMenuService.GetMenuCount()
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(count)
}

func (s *sysMenuController) DeleteMenuByID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := s.sysMenuService.DeleteMenuByID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("删除成功")
}

func (s *sysMenuController) UpdateMenu(ctx *gin.Context) {
	result := r.NewResult(ctx)

	idStr := ctx.PostForm("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	parentIDStr := ctx.PostForm("parentMenuID")
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	code := ctx.PostForm("code")
	name := ctx.PostForm("name")
	description := ctx.PostForm("description")
	menu := &po.SysMenu{
		ParentMenuID: uint(parentID),
		Code:         code,
		Name:         name,
		Description:  description,
	}
	menu.ID = uint(id)
	err2 := s.sysMenuService.UpdateMenu(menu)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("修改成功")
}

func (s *sysMenuController) GetMenuByID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
	}
	menu, err2 := s.sysMenuService.GetMenuByID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(menu)
}

func (s *sysMenuController) GetMenuTree(ctx *gin.Context) {
	result := r.NewResult(ctx)
	menuTree, err := s.sysMenuService.GetMenuTree()
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(menuTree)
}

func (s *sysMenuController) InsertMenu(ctx *gin.Context) {
	result := r.NewResult(ctx)
	parentIDStr := ctx.PostForm("parentMenuID")
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	code := ctx.PostForm("code")
	name := ctx.PostForm("name")
	description := ctx.PostForm("description")
	menu := &po.SysMenu{
		ParentMenuID: uint(parentID),
		Code:         code,
		Name:         name,
		Description:  description,
	}
	id, err2 := s.sysMenuService.InsertMenu(menu)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(id)
}
