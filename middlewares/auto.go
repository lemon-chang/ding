package middlewares

import (
	"ding/global"
	"ding/initialize/jwt"
	"ding/response"
	"github.com/gin-gonic/gin"
	"strings"
)

//当在中间件或 handler 中启动新的 Goroutine 时，不能使用原始的上下文，必须使用只读副本。
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
		// 这里假设Token放在Header的Authorization中，并使用Bearer开头
		// 这里的具体实现方式要依据你的实际业务情况决定
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			response.ResponseError(c, response.CodeNeedLogin)
			c.Abort()
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.ResponseError(c, response.CodeInvalidToken)
			c.Abort()
			return
		}
		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := (&jwt.MyClaims{}).ParseToken(parts[1])
		if err != nil {
			response.ResponseError(c, response.CodeInvalidToken)
			c.Abort()
			return
		}
		// 将当前请求的user的ID信息保存到请求的上下文c上
		c.Set(global.CtxUserIDKey, mc.UserId)
		c.Set(global.CtxUserNameKey, mc.Username)
		c.Set(global.CtxUserAuthorityIDKey, mc.AuthorityID)
		c.Next() // 后续的处理函数可以用过c.Get("username")来获取当前请求的用户信息
	}
}
