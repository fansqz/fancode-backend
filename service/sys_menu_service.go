package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"gorm.io/gorm"
	"log"
)

type SysMenuService interface {

	// GetMenuCount 获取menu数目
	GetMenuCount() (int64, *e.Error)
	// DeleteMenuByID 删除menu
	DeleteMenuByID(id uint) *e.Error
	// UpdateMenu 更新menu
	UpdateMenu(menu *po.SysMenu) *e.Error
	// GetMenuByID 根据id获取menu
	GetMenuByID(id uint) (*po.SysMenu, *e.Error)
	// GetMenuTree 获取menu树
	GetMenuTree() ([]*dto.SysMenuTreeDto, *e.Error)
	// InsertMenu 添加menu
	InsertMenu(menu *po.SysMenu) (uint, *e.Error)
}

type sysMenuService struct {
	db *gorm.DB
}

func NewSysMenuService() SysMenuService {
	return &sysMenuService{}
}

func (s *sysMenuService) GetMenuCount() (int64, *e.Error) {
	count, err := dao.GetMenuCount(global.Mysql)
	if err != nil {
		log.Println(err)
		return 0, e.ErrMenuUnknownError
	}
	return count, nil
}

// DeleteMenuByID 根据menu的id进行删除
func (s *sysMenuService) DeleteMenuByID(id uint) *e.Error {
	err := global.Mysql.Transaction(func(tx *gorm.DB) error {
		// 递归删除API
		return s.deleteMenusRecursive(tx, id)
	})

	if err != nil {
		return e.ErrMenuUnknownError
	}

	return nil
}

// deleteMenusRecursive 递归删除API
func (s *sysMenuService) deleteMenusRecursive(db *gorm.DB, parentID uint) error {
	childMenus, err := dao.GetChildMenusByParentID(db, parentID)
	if err != nil {
		return err
	}
	for _, childAPI := range childMenus {
		// 删除子menu的子menu
		if err = s.deleteMenusRecursive(db, childAPI.ID); err != nil {
			return err
		}
	}
	// 当前menu
	if err = dao.DeleteMenuByID(db, parentID); err != nil {
		return err
	}
	return nil
}

func (s *sysMenuService) UpdateMenu(menu *po.SysMenu) *e.Error {
	err := dao.UpdateMenu(global.Mysql, menu)
	if gorm.ErrRecordNotFound == err {
		return e.ErrMenuNotExist
	}
	return nil
}

func (s *sysMenuService) GetMenuByID(id uint) (*po.SysMenu, *e.Error) {
	menu, err := dao.GetMenuByID(global.Mysql, id)
	if err != nil {
		return nil, e.ErrMenuUnknownError
	}
	return menu, nil
}

func (s *sysMenuService) GetMenuTree() ([]*dto.SysMenuTreeDto, *e.Error) {
	var menuList []*po.SysMenu
	var err error
	if menuList, err = dao.GetAllMenu(global.Mysql); err != nil {
		log.Println(err)
		return nil, e.ErrMenuUnknownError
	}

	menuMap := make(map[uint]*dto.SysMenuTreeDto)
	var rootMenus []*dto.SysMenuTreeDto

	// 添加到map中保存
	for _, menu := range menuList {
		menuMap[menu.ID] = dto.NewSysMenuTreeDto(menu)
	}

	// 遍历并添加到父节点中
	for _, menu := range menuList {
		if menu.ParentMenuID == 0 {
			rootMenus = append(rootMenus, menuMap[menu.ID])
		} else {
			parentMenu, exists := menuMap[menu.ParentMenuID]
			if !exists {
				return nil, e.ErrMenuUnknownError
			}
			parentMenu.Children = append(parentMenu.Children, menuMap[menu.ID])
		}
	}

	return rootMenus, nil
}

func (s *sysMenuService) InsertMenu(menu *po.SysMenu) (uint, *e.Error) {
	err := dao.InsertMenu(global.Mysql, menu)
	if err != nil {
		return 0, e.ErrMenuUnknownError
	}
	return menu.ID, nil
}
