package FanCode

import (
	"FanCode/db"
	"FanCode/models"
	"FanCode/setting"
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
	if err := db.InitMysql(setting.Conf.MySqlConfig); err != nil {
		fmt.Println("数据库连接失败")
	}

	// 模型绑定
	db.DB.AutoMigrate(&models.User{})
	defer db.CloseMysql()

}
