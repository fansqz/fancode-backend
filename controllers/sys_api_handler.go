package controllers

import (
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type SysApiController interface {
	// GetApiCount 获取api数目
	GetApiCount(ctx *gin.Context)
	// DeleteApiByID 删除api
	DeleteApiByID(ctx *gin.Context)
	// UpdateApi 更新api
	UpdateApi(ctx *gin.Context)
	// GetApiByID 根据id获取api
	GetApiByID(ctx *gin.Context)
	// GetApiTree 获取api树
	GetApiTree(ctx *gin.Context)
	// InsertApi 添加api
	InsertApi(ctx *gin.Context)
}

type sysApiController struct {
	sysApiService service.SysApiService
}

func NewSysApiController() SysApiController {
	return &sysApiController{
		sysApiService: service.NewSysApiService(),
	}
}

func (s *sysApiController) GetApiCount(ctx *gin.Context) {
	result := r.NewResult(ctx)
	count, err := s.sysApiService.GetApiCount()
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(count)
}

func (s *sysApiController) DeleteApiByID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := s.sysApiService.DeleteApiByID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("删除成功")
}

func (s *sysApiController) UpdateApi(ctx *gin.Context) {
	result := r.NewResult(ctx)

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	parentIDStr := ctx.Param("parentApiID")
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	path := ctx.Param("path")
	method := ctx.Param("method")
	name := ctx.Param("name")
	description := ctx.Param("description")
	api := &po.SysApi{
		ParentApiID: uint(parentID),
		Path:        path,
		Method:      method,
		Name:        name,
		Description: description,
	}
	api.ID = uint(id)
	err2 := s.sysApiService.UpdateApi(api)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("修改成功")
}

func (s *sysApiController) GetApiByID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
	}
	api, err2 := s.sysApiService.GetApiByID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(api)
}

func (s *sysApiController) GetApiTree(ctx *gin.Context) {
	result := r.NewResult(ctx)
	apiTree, err := s.sysApiService.GetApiTree()
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(apiTree)
}

func (s *sysApiController) InsertApi(ctx *gin.Context) {
	result := r.NewResult(ctx)
	parentIDStr := ctx.Param("parentApiID")
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	path := ctx.Param("path")
	method := ctx.Param("method")
	name := ctx.Param("name")
	description := ctx.Param("description")
	api := &po.SysApi{
		ParentApiID: uint(parentID),
		Path:        path,
		Method:      method,
		Name:        name,
		Description: description,
	}
	id, err2 := s.sysApiService.InsertApi(api)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(id)
}
