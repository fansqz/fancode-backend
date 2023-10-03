package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"gorm.io/gorm"
	"log"
	"time"
)

type SysApiService interface {

	// GetApiCount 获取api数目
	GetApiCount() (int64, *e.Error)
	// DeleteApiByID 删除api
	DeleteApiByID(id uint) *e.Error
	// UpdateApi 更新api
	UpdateApi(api *po.SysApi) *e.Error
	// GetApiByID 根据id获取api
	GetApiByID(id uint) (*po.SysApi, *e.Error)
	// GetApiTree 获取api树
	GetApiTree() ([]*dto.SysApiTreeDto, *e.Error)
	// InsertApi 添加api
	InsertApi(api *po.SysApi) (uint, *e.Error)
}

type sysApiService struct {
	sysApiDao dao.SysApiDao
}

func NewSysApiService(apiDao dao.SysApiDao) SysApiService {
	return &sysApiService{
		sysApiDao: apiDao,
	}
}

func (s *sysApiService) GetApiCount() (int64, *e.Error) {
	count, err := s.sysApiDao.GetApiCount(global.Mysql)
	if err != nil {
		log.Println(err)
		return 0, e.ErrApiUnknownError
	}
	return count, nil
}

// DeleteApiByID 根据api的id进行删除
func (s *sysApiService) DeleteApiByID(id uint) *e.Error {
	err := global.Mysql.Transaction(func(tx *gorm.DB) error {
		// 递归删除API
		return s.deleteApisRecursive(tx, id)
	})

	if err != nil {
		return e.ErrApiUnknownError
	}

	return nil
}

// deleteApisRecursive 递归删除API
func (s *sysApiService) deleteApisRecursive(db *gorm.DB, parentID uint) error {
	childApis, err := s.sysApiDao.GetChildApisByParentID(db, parentID)
	if err != nil {
		return err
	}
	for _, childAPI := range childApis {
		// 删除子api的子api
		if err = s.deleteApisRecursive(db, childAPI.ID); err != nil {
			return err
		}
	}
	// 当前api
	if err = s.sysApiDao.DeleteApiByID(db, parentID); err != nil {
		return err
	}
	return nil
}

func (s *sysApiService) UpdateApi(api *po.SysApi) *e.Error {
	api.UpdatedAt = time.Now()
	err := s.sysApiDao.UpdateApi(global.Mysql, api)
	if gorm.ErrRecordNotFound == err {
		return e.ErrApiNotExist
	} else if err != nil {
		return e.ErrMysql
	}
	return nil
}

func (s *sysApiService) GetApiByID(id uint) (*po.SysApi, *e.Error) {
	api, err := s.sysApiDao.GetApiByID(global.Mysql, id)
	if err != nil {
		return nil, e.ErrApiUnknownError
	}
	return api, nil
}

func (s *sysApiService) GetApiTree() ([]*dto.SysApiTreeDto, *e.Error) {
	var apiList []*po.SysApi
	var err error
	if apiList, err = s.sysApiDao.GetAllApi(global.Mysql); err != nil {
		log.Println(err)
		return nil, e.ErrApiUnknownError
	}

	apiMap := make(map[uint]*dto.SysApiTreeDto)
	var rootApis []*dto.SysApiTreeDto

	// 添加到map中保存
	for _, api := range apiList {
		apiMap[api.ID] = dto.NewSysApiTreeDto(api)
	}

	// 遍历并添加到父节点中
	for _, api := range apiList {
		if api.ParentApiID == 0 {
			rootApis = append(rootApis, apiMap[api.ID])
		} else {
			parentApi, exists := apiMap[api.ParentApiID]
			if !exists {
				return nil, e.ErrApiUnknownError
			}
			parentApi.Children = append(parentApi.Children, apiMap[api.ID])
		}
	}

	return rootApis, nil
}

func (s *sysApiService) InsertApi(api *po.SysApi) (uint, *e.Error) {
	err := s.sysApiDao.InsertApi(global.Mysql, api)
	if err != nil {
		return 0, e.ErrApiUnknownError
	}
	return api.ID, nil
}
