package routers

import (
	v1 "ding/controllers/v1"
	"ding/controllers/v2/ding"
	"ding/global"
	"ding/initialize/logger"
	"ding/routers/dingding"
	"ding/routers/system"

	"ding/middlewares"

	"fmt"

	"github.com/gin-gonic/gin"

	"net/http"

	"go.uber.org/zap"
)

func Setup(mode string) *gin.Engine {

	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) //设置为发布模式
	}
	//con参数检验 server逻辑处理 dao数据操作
	r := gin.New()
	//r.Use(cors.Default()) //第三方库
	r.Use(middlewares.Cors())
	fmt.Println(middlewares.Cors())

	zap.L().Info("跨域配置完成")

	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	V3 := r.Group("/api/v3")
	V3.POST("/jk", v1.Jk)
	V3.POST("/zjq", v1.Zjq)
	V3.POST("/lxy", v1.Lxy)

	V3.POST("/outgoing", ding.OutGoing) //outgoing接口是让官方
	V3.POST("/robotAt", ding.RobotAt)

	V3.POST("/gxpRobot", ding.GxpRobot)
	System := r.Group("/api/system")
	System.Use(middlewares.JWTAuthMiddleware())
	system.SetupSystem(System)
	Ding := r.Group("/api/ding")
	{
		Ding.POST("login", ding.LoginHandler)
		Ding.POST("singleChat", ding.ChatHandler)
		//放给钉钉用的接口
		Ding.POST("subscribeTo", ding.SubscribeTo)
		//获取力扣地址
		Ding.POST("getLeetCode", ding.GetLeetCode)
	}

	Ding.Use(middlewares.JWTAuthMiddleware())
	dingding.SetupDing(Ding)
	V3.GET("upload", func(c *gin.Context) {
		username, _ := c.Get(global.CtxUserNameKey)
		c.File(fmt.Sprintf("Screenshot_%s.png", username))
	})
	{
		V3.GET("/", v1.WelcomeHandler)
	}
	//注册业务路由
	V3.Use(middlewares.JWTAuthMiddleware())
	{ //中间件是会把路由改变的

		//我在send里面对全局的Gcontab进行操作
		V3.POST("/getTasks", v1.GetTasks) //获取到定时任务
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
