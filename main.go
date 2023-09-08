package main

import (
	"FanCode/global"
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
	if err := global.InitSetting(path); err != nil {
		fmt.Println("加载配置文件出错")
		return
	}

	//连接数据库
	if err := global.InitMysql(global.Conf.MySqlConfig); err != nil {
		fmt.Println("数据库连接失败")
	}

	//连接redis
	if err := global.InitRedis(global.Conf.RedisConfig); err != nil {
		fmt.Println("redis连接失败")
	}

	// 模型绑定
	global.Mysql.AutoMigrate(&po.SysUser{})
	global.Mysql.AutoMigrate(&po.SysApi{})
	global.Mysql.AutoMigrate(&po.SysMenu{})
	global.Mysql.AutoMigrate(&po.SysRole{})
	global.Mysql.AutoMigrate(&po.SysUser{})
	global.Mysql.AutoMigrate(&po.Problem{})
	global.Mysql.AutoMigrate(&po.Submission{})
	global.Mysql.AutoMigrate(&po.ProblemAttempt{})
	//注册路由
	routers.Run()
}
