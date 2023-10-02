package dao

import (
	"FanCode/models/po"
	"gorm.io/gorm"
)

// GetApiByID 通过api的id获取api
func GetApiByID(db *gorm.DB, id uint) (*po.SysApi, error) {
	var api po.SysApi
	err := db.First(&api, id).Error
	return &api, err
}

// GetApiListByParentID 通过父id找到其所有子api
func GetApiListByParentID(db *gorm.DB, parentID int32) ([]*po.SysApi, error) {
	var sysApis []*po.SysApi
	err := db.Where("parent_api_id = ?", parentID).Find(&sysApis).Error
	if err != nil {
		return nil, err
	}
	return sysApis, nil
}

// GetApiCount 获取api总数
func GetApiCount(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&po.SysApi{}).Count(&count).Error
	return count, err
}

// GetApiListByPathKeyword 模糊查询api
func GetApiListByPathKeyword(db *gorm.DB, keyword string, page int, pageSize int) ([]*po.SysApi, error) {
	var sysApis []*po.SysApi
	err := db.Where("path LIKE ?", "%"+keyword+"%").Offset((page - 1) * pageSize).Limit(pageSize).Find(&sysApis).Error
	if err != nil {
		return nil, err
	}
	return sysApis, nil
}

// GetChildApisByParentID 根据父API的ID获取所有子API
func GetChildApisByParentID(db *gorm.DB, parentID uint) ([]*po.SysApi, error) {
	var childApis []*po.SysApi
	if err := db.Where("parent_api_id = ?", parentID).Find(&childApis).Error; err != nil {
		return nil, err
	}
	return childApis, nil
}

// GetAllApi 获取所有api
func GetAllApi(db *gorm.DB) ([]*po.SysApi, error) {
	var apiList []*po.SysApi
	err := db.Find(&apiList).Error
	return apiList, err
}

// InsertApi 创建角色
func InsertApi(db *gorm.DB, api *po.SysApi) error {
	return db.Create(api).Error
}

// UpdateApi 修改api
func UpdateApi(db *gorm.DB, api *po.SysApi) error {
	return db.Model(api).Updates(api).Error
}

// DeleteApiByID 根据api的id进行删除
func DeleteApiByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.SysApi{}, id).Error
}
