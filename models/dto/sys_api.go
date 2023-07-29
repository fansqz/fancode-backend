package dto

import "FanCode/models/po"

type SysApiTreeDto struct {
	ID          uint             `json:"id"`
	ParentApiID uint             `json:"parentApiID"`
	Path        string           `json:"path"`        // 请求路径
	Method      string           `json:"method"`      // 请求方法
	Name        string           `json:"name"`        // 请求名称
	Description string           `json:"description"` // 描述
	Sort        int              `json:"sort"`        // 排序
	Children    []*SysApiTreeDto `json:"children"`    //子api
}

func NewSysApiTreeDto(sysApi *po.SysApi) *SysApiTreeDto {
	return &SysApiTreeDto{
		ID:          sysApi.ID,
		ParentApiID: sysApi.ParentApiID,
		Path:        sysApi.Path,
		Method:      sysApi.Method,
		Name:        sysApi.Name,
		Description: sysApi.Description,
		Sort:        sysApi.Sort,
	}
}
