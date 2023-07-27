package dao

import (
	"FanCode/models/po"
	"github.com/jinzhu/gorm"
)

// InsertUser 向数据库中插入用户信息
func InsertUser(db *gorm.DB, user *po.User) {
	db.Create(user)
}

// GetUserByUserCode
func GetUserByUserCode(db *gorm.DB, userID string) (*po.User, error) {
	//写sql语句
	sqlStr := `select id,username,password,email,sex,phone,code,role 
	from users where code = ?`
	//执行
	row := db.Raw(sqlStr, userID)
	user := &po.User{}
	row.Scan(&user)
	return user, nil
}

// UpdateUser 更新用户
func UpdateUser(db *gorm.DB, user *po.User) error {
	sqlStr := "update `users` set username = ?, password = ?, email = ?, phone = ?, sex = ? where id = ?"
	//执行
	db.Exec(sqlStr, user.Username, user.Password, user.Email, user.Phone, user.Sex, user.ID)
	return nil
}

// CheckUserCode检测用户code是否存在
func CheckUserCode(db *gorm.DB, userCode string) bool {
	//执行
	row := db.Model(&po.User{}).Select("user_code").Where("code = ?", userCode)
	user := &po.User{}
	row.Scan(&user)
	return user.Code != ""
}
