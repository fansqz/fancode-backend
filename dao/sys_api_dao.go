package dao

import (
	"FanCode/models/po"
	"github.com/jinzhu/gorm"
)

// InsertApi 创建角色
func InsertApi(db *gorm.DB, api *po.SysApi) error {
	return db.Create(api).Error
}

// GetApiByID 通过api的id获取api
func GetApiByID(db *gorm.DB, id uint) (*po.SysApi, error) {
	var api po.SysApi
	err := db.First(&api, id).Error
	return &api, err
}

// GetSysApiListByParentID 通过父id找到其所有子api
func GetSysApiListByParentID(db *gorm.DB, parentID int32) ([]*po.SysApi, error) {
	var sysApis []*po.SysApi
	err := db.Where("parent_api_id = ?", parentID).Find(&sysApis).Error
	if err != nil {
		return nil, err
	}
	return sysApis, nil
}

// GetApiCount 获取api总数
func GetApiCount(db *gorm.DB) (uint, error) {
	var count uint
	err := db.Model(&po.SysApi{}).Count(&count).Error
	return count, err
}

// GetSysApiListByPathKeyword 模糊查询api
func GetSysApiListByPathKeyword(db *gorm.DB, keyword string, page int, pageSize int) ([]*po.SysApi, error) {
	var sysApis []*po.SysApi
	err := db.Where("path LIKE ?", "%"+keyword+"%").Offset((page - 1) * pageSize).Limit(pageSize).Find(&sysApis).Error
	if err != nil {
		return nil, err
	}
	return sysApis, nil
}

// DeleteApiByID 根据api的id进行删除
func DeleteApiByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.SysApi{}, id).Error
}

// UpdateApi 修改api
func UpdateApi(db *gorm.DB, api *po.SysApi) error {
	return db.Save(api).Error
}
