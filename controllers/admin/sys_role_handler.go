package admin

import (
	"FanCode/controllers"
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

// SysRoleController 角色管理相关功能
type SysRoleController interface {
	// GetRoleByID 根据id获取role
	GetRoleByID(ctx *gin.Context)
	// InsertSysRole 添加角色
	InsertSysRole(ctx *gin.Context)
	// DeleteSysRole 删除角色
	DeleteSysRole(ctx *gin.Context)
	// GetSysRoleList 获取角色列表
	GetSysRoleList(ctx *gin.Context)
	// UpdateSysRole 更新角色
	UpdateSysRole(ctx *gin.Context)
	// UpdateRoleApis 更新角色api
	UpdateRoleApis(ctx *gin.Context)
	// UpdateRoleMenus 更新角色menu
	UpdateRoleMenus(ctx *gin.Context)
	// GetApiIDsByRoleID 通过角
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

func (s *sysRoleController) GetRoleByID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	roleIDStr := ctx.Param("id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	user, err2 := s.sysRoleService.GetRoleByID(uint(roleID))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(user)
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
	pageQuery, err := controllers.GetPageQueryByQuery(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	role := &po.SysRole{
		Name:        ctx.Query("name"),
		Description: ctx.Query("description"),
	}
	pageQuery.Query = role
	pageInfo, err := s.sysRoleService.GetSysRoleList(pageQuery)
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

func (s *sysRoleController) UpdateRoleApis(ctx *gin.Context) {
	result := r.NewResult(ctx)
	var json insertApisToRoleRequest
	err := ctx.BindJSON(&json)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := s.sysRoleService.UpdateRoleApis(json.RoleID, json.ApiIDs)
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

func (s *sysRoleController) UpdateRoleMenus(ctx *gin.Context) {
	result := r.NewResult(ctx)
	var json insertMenusToRoleRequest
	err := ctx.BindJSON(&json)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := s.sysRoleService.UpdateRoleMenus(json.RoleID, json.MenuIDs)
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
