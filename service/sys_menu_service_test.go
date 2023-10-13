package service

import (
	"FanCode/dao/mock"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"testing"
	"time"
)

func TestSysMenuService_GetMenuCount(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	menuDao := mock.NewMockSysMenuDao(mockCtl)
	menuDao.EXPECT().GetMenuCount(gomock.Any()).Return(int64(10), nil)
	menuDao.EXPECT().GetMenuCount(gomock.Any()).Return(int64(0), gorm.ErrInvalidDB)
	menuService := NewSysMenuService(menuDao)
	count, err := menuService.GetMenuCount()
	assert.Equal(t, int64(10), count)
	assert.Nil(t, err)
	count, err = menuService.GetMenuCount()
	assert.Equal(t, err, e.ErrMysql)
	assert.Equal(t, int64(0), count)

}

func TestSysMenuService_DeleteMenuByID(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	menuDao := mock.NewMockSysMenuDao(mockCtl)

	// mock数据库
	db, mock, err := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectCommit()
	assert.Nil(t, err)
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		SkipInitializeWithVersion: true,
		Conn:                      db,
	}), &gorm.Config{})
	assert.Nil(t, err)
	global.Mysql = gormDB

	// 生成api树
	menus1 := make([]*po.SysMenu, 8)
	menus2 := make([]*po.SysMenu, 8)
	for i := 0; i < 8; i++ {
		menu := &po.SysMenu{}
		menu.ID = uint(i + 2)
		menus1[i] = menu
		menu = &po.SysMenu{}
		menu.ID = uint(i + 10)
		menus2[i] = menu
	}
	menuDao.EXPECT().GetChildMenusByParentID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(db2 *gorm.DB, id uint) ([]*po.SysMenu, *e.Error) {
			if id == 1 {
				return menus1, nil
			}
			if id == 9 {
				return menus2, nil
			}
			return []*po.SysMenu{}, nil
		}).AnyTimes()

	deleteCount := 0
	menuDao.EXPECT().DeleteMenuByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(db *gorm.DB, menuID uint) error {
			deleteCount += int(menuID)
			return nil
		}).AnyTimes()

	// 测试
	menuService := NewSysMenuService(menuDao)
	err2 := menuService.DeleteMenuByID(1)
	assert.Nil(t, err2)
	assert.Equal(t, deleteCount, ((1+17)/2)*17)
}

func TestSysMenuService_GetMenuByID(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock apiDao
	menuDao := mock.NewMockSysMenuDao(mockCtl)
	menu := &po.SysMenu{
		Name:         "menu名",
		Description:  "menu描述",
		Code:         "menu权限值",
		ParentMenuID: 10,
	}
	menu.ID = 1
	menuDao.EXPECT().GetMenuByID(gomock.Any(), uint(1)).Return(menu, nil)
	menuDao.EXPECT().GetMenuByID(gomock.Any(), uint(2)).Return(nil, gorm.ErrRecordNotFound)
	menuDao.EXPECT().GetMenuByID(gomock.Any(), uint(3)).Return(nil, gorm.ErrInvalidDB)

	// 测试
	menuService := NewSysMenuService(menuDao)
	menu2, err := menuService.GetMenuByID(1)
	assert.Equal(t, menu, menu2)
	assert.Nil(t, err)

	menu3, err := menuService.GetMenuByID(2)
	assert.Nil(t, menu3)
	assert.Equal(t, err, e.ErrMenuNotExist)

	menu4, err := menuService.GetMenuByID(3)
	assert.Nil(t, menu4)
	assert.Equal(t, err, e.ErrMysql)
}

func TestSysMenuService_UpdateMenu(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock apiDao
	menuDao := mock.NewMockSysMenuDao(mockCtl)
	menu := &po.SysMenu{
		Name:         "menu名",
		Description:  "menu描述",
		Code:         "menuCode",
		ParentMenuID: 10,
	}
	menu.ID = 1
	menuDao.EXPECT().UpdateMenu(gomock.Any(), menu).
		DoAndReturn(func(db *gorm.DB, sysMenu *po.SysMenu) error {
			assert.NotEqual(t, sysMenu.UpdatedAt, time.Time{})
			sysMenu.UpdatedAt = time.Time{}
			assert.Equal(t, menu, sysMenu)
			return nil
		})
	menuDao.EXPECT().UpdateMenu(gomock.Any(), gomock.Any()).Return(gorm.ErrInvalidDB)

	// 测试
	menuService := NewSysMenuService(menuDao)
	err := menuService.UpdateMenu(menu)
	assert.Nil(t, err)
	err = menuService.UpdateMenu(&po.SysMenu{})
	assert.Equal(t, err, e.ErrMysql)
}

func TestSysMenuService_InsertMenu(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock apiDao
	menuDao := mock.NewMockSysMenuDao(mockCtl)
	menu := &po.SysMenu{
		Name:         "menu名",
		Description:  "menu描述",
		Code:         "menuCode",
		ParentMenuID: 10,
	}
	menuDao.EXPECT().InsertMenu(gomock.Any(), menu).
		DoAndReturn(func(db *gorm.DB, sysMenu *po.SysMenu) error {
			assert.Equal(t, menu, sysMenu)
			sysMenu.ID = 1
			return nil
		})
	menuDao.EXPECT().InsertMenu(gomock.Any(), gomock.Any()).Return(gorm.ErrInvalidDB)

	// 测试
	menuService := NewSysMenuService(menuDao)
	id, err := menuService.InsertMenu(menu)
	assert.Equal(t, id, uint(1))
	assert.Nil(t, err)
	id, err = menuService.InsertMenu(&po.SysMenu{})
	assert.Equal(t, id, uint(0))
	assert.Equal(t, err, e.ErrMysql)
}

func TestSysMenuService_GetMenuTree(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock apiDao
	menuDao := mock.NewMockSysMenuDao(mockCtl)
	menus := make([]*po.SysMenu, 4)
	for i := 0; i < 4; i++ {
		menu := &po.SysMenu{}
		menu.Name = "menu" + strconv.Itoa(i)
		menu.Description = "menu描述" + strconv.Itoa(i)
		menu.Code = "menuPath" + strconv.Itoa(i)
		menu.ID = uint(i + 1)
		menus[i] = menu
	}
	menus[1].ParentMenuID = 1
	menus[2].ParentMenuID = 1
	menus[3].ParentMenuID = 1
	menuDao.EXPECT().GetAllMenu(gomock.Any()).Return(menus, nil)
	menuDao.EXPECT().GetAllMenu(gomock.Any()).Return([]*po.SysMenu{}, gorm.ErrInvalidDB)

	// 测试
	menuService := NewSysMenuService(menuDao)

	treeDtos, err := menuService.GetMenuTree()
	treeDto := dto.NewSysMenuTreeDto(menus[0])
	for i := 1; i < 4; i++ {
		treeDto.Children = append(treeDto.Children, dto.NewSysMenuTreeDto(menus[i]))
	}
	assert.Equal(t, []*dto.SysMenuTreeDto{treeDto}, treeDtos)
	assert.Nil(t, err)

	treeDtos, err = menuService.GetMenuTree()
	assert.Nil(t, treeDtos)
	assert.Equal(t, err, e.ErrMysql)
}
