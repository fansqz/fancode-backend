package main

import (
	"FanCode/initialize"
	"FanCode/initialize/setting"
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
	if err := setting.Init(path); err != nil {
		fmt.Println("加载配置文件出错")
		return
	}

	//连接数据库
	if err := initialize.InitMysql(setting.Conf.MySqlConfig); err != nil {
		fmt.Println("数据库连接失败")
	}

	// 模型绑定
	initialize.DB.AutoMigrate(&po.User{})
	initialize.DB.AutoMigrate(&po.Problem{})
	initialize.DB.AutoMigrate(&po.Submission{})
	defer initialize.CloseMysql()

	//注册路由
	routers.Run()
}
