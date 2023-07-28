package main

import (
	"FanCode/global"
	"FanCode/initialize"
	"FanCode/models/po"
	"FanCode/routers"
	"fmt"
	"os"
	"strings"
)

func main() {
	//获取参数
	path, _ := os.Getwd()
	path = strings.ReplaceAll(path, "\\", "/")
	path = path + "/conf/config.ini"

	//加载配置
	if err := initialize.InitSetting(path); err != nil {
		fmt.Println("加载配置文件出错")
		return
	}

	//连接数据库
	if err := initialize.InitMysql(global.Conf.MySqlConfig); err != nil {
		fmt.Println("数据库连接失败")
	}

	// 模型绑定
	global.Mysql.AutoMigrate(&po.SysUser{})
	global.Mysql.AutoMigrate(&po.SysApi{})
	global.Mysql.AutoMigrate(&po.SysRole{})
	global.Mysql.AutoMigrate(&po.SysUser{})
	global.Mysql.AutoMigrate(&po.Problem{})
	global.Mysql.AutoMigrate(&po.Submission{})
	defer initialize.CloseMysql()

	//注册路由
	routers.Run()
}
