package v1

import (
	"ding/response"
	"github.com/gin-gonic/gin"
)

//SingUpHandler 处理注册的请求函数

func WelcomeHandler(c *gin.Context) {
	response.ResponseSuccess(c, gin.H{
		"message": "hello",
	})
}
