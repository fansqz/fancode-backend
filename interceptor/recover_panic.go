package interceptor

import (
	"FanCode/models/vo"
	"github.com/gin-gonic/gin"
	"log"
)

type RecoverPanicInterceptor struct {
}

func NewRecoverPanicInterceptor() *RecoverPanicInterceptor {
	return &RecoverPanicInterceptor{}
}

func (i *RecoverPanicInterceptor) RecoverPanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			result := vo.NewResult(c)
			if err := recover(); err != nil {
				// 记录 panic 的错误信息
				log.Printf("Recovered from panic: %s\n", err)
				// 给客户端一个友好的错误提示
				result.SimpleErrorMessage("系统错误")
			}
		}()

		// 继续执行下一个中间件或者处理函数
		c.Next()
	}
}
