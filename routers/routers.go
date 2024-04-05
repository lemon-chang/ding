package routers

import (
	ding2 "ding/controllers/ding"
	"ding/global"
	"ding/initialize/logger"
	"ding/middlewares"
	"ding/routers/dingding"
	"ding/routers/system"
	"github.com/gin-gonic/gin"

	"net/http"
)

func Setup(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) //设置为发布模式
	}
	r := gin.New()
	global.GLOBAL_GIN_Engine = r
	r.Use(middlewares.Cors(), logger.GinLogger(), logger.GinRecovery(true))
	/*=========系统路由==========*/
	System := r.Group("/api/system") // 此处engine可以直接调用RouterGroup的方法，原因不详
	System.Use(middlewares.JWTAuthMiddleware())
	system.SetupSystem(System)
	/*=========钉钉回调、无需token验证路由==========*/
	V3 := r.Group("/api/v3")
	//V3.POST("/outgoing", ding2.OutGoing) //outgoing接口是让官方
	//V3.POST("/robotAt", ding2.RobotAt)
	V3.GET("GetAllUsers", ding2.SelectAllUsers) // 查询所有用户信息
	/*=========dingEvent事件回调路由==========*/
	DingEvent := r.Group("dingEvent")
	dingding.SetupDingEventSubscription(DingEvent)

	/*=========具体业务路由==========*/
	Ding := r.Group("/api/ding")
	{
		//无需token验证
		Ding.POST("login", ding2.LoginHandler)
		Ding.POST("loginByDingDing", ding2.LoginByDingDing) // 判断钉钉扫码登陆

	}
	Ding.Use(middlewares.JWTAuthMiddleware())
	Ding.POST("loginByToken", ding2.LoginHandlerByToken) //单点登录后续要用
	dingding.SetupDing(Ding)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
