package dao

import (
	"FanCode/db"
	"FanCode/models"
)

// InsertUser 向数据库中插入用户信息
func InsertUser(user *models.User) {
	db.DB.Create(user)
}

// GetUserByName
func GetUserByName(username string) (*models.User, error) {
	//写sql语句
	sqlStr := `select id,username,password,email,sex,phone,user_id,role 
	from users where username = ?`
	//执行
	row := db.DB.Raw(sqlStr, username)
	user := &models.User{}
	row.Scan(&user)
	return user, nil
}

// GetUserByUserID
func GetUserByUserID(userID string) (*models.User, error) {
	//写sql语句
	sqlStr := `select id,username,password,email,sex,phone,user_id,role 
	from users where user_id = ?`
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
