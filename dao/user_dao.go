package dao

import (
	"FanCode/db"
	"FanCode/models"
)

// InsertUser 向数据库中插入用户信息
func InsertUser(user *models.User) {
	db.DB.Create(user)
}

// GetUserByUserNumber
func GetUserByUserNumber(userID string) (*models.User, error) {
	//写sql语句
	sqlStr := `select id,username,password,email,sex,phone,number,role 
	from users where number = ?`
	//执行
	row := db.DB.Raw(sqlStr, userID)
	user := &models.User{}
	row.Scan(&user)
	return user, nil
}

// UpdateUser 更新用户
func UpdateUser(user *models.User) error {
	sqlStr := "update `users` set username = ?, password = ?, email = ?, phone = ?, sex = ? where id = ?"
	//执行
	db.DB.Exec(sqlStr, user.Username, user.Password, user.Email, user.Phone, user.Sex, user.ID)
	return nil
}

// CheckUserNumber检测用户number是否存在
func CheckUserNumber(userNumber string) bool {
	//执行
	row := db.DB.Model(&models.User{}).Select("user_number").Where("user_number = ?", userNumber)
	user := &models.User{}
	row.Scan(&user)
	return user.Number != ""
}
