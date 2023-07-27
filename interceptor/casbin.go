package interceptor

import (
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func Casbin() *casbin.Enforcer {
	a, _ := gormadapter.NewAdapter("mysql", "root:miconvert*.*@tcp(49.234.56:3306)/") // Your driver and data source.
	e, err := casbin.NewEnforcer("./model.conf", a)
	if err != nil {
		log.Fatal("载入casbin配置出错")
	}

	e.LoadPolicy() // 从数据库载入配置
	return e
}
