package routers

import (
	ding2 "ding/controllers/ding"
	"ding/global"
	"ding/initialize/logger"
	"ding/middlewares"
	"ding/routers/dingding"
	"ding/routers/personal"
	"ding/routers/system"
	"fmt"

	"github.com/gin-gonic/gin"

	"net/http"

	"go.uber.org/zap"
)

type MyRouterGroup struct {
	RouterGroup *gin.RouterGroup
}

func (group *MyRouterGroup) DingEvent(event_url string, handlerFunc gin.HandlerFunc) {
	group.RouterGroup.Handle("DINGEVENT", event_url, handlerFunc)
}
func Setup(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) //设置为发布模式
	}
	r := gin.New()
	r.GET("chat_update_title", func(context *gin.Context) {
		fmt.Println("测试chat_update_title")
	})
	group := MyRouterGroup{RouterGroup: &r.RouterGroup}

	group.DingEvent("chat_update_title", func(c *gin.Context) {
		fmt.Println("测试chat_update_title")
	})
	global.GLOBAL_GIN_Engine = r
	//r.Use(cors.Default()) //第三方库
	r.Use(middlewares.Cors())
	zap.L().Info("跨域配置完成")
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	/*=========系统路由==========*/
	System := r.Group("/api/system") // 此处engine可以直接调用RouterGroup的方法，原因不详
	System.Use(middlewares.JWTAuthMiddleware())
	system.SetupSystem(System)
	/*=========私人个性化路由==========*/
	Personal := r.Group("/api/personal")
	personal.SetupPersonal(Personal)
	/*=========钉钉回调、无需token验证路由==========*/
	V3 := r.Group("/api/v3")
	//V3.POST("/outgoing", ding2.OutGoing) //outgoing接口是让官方
	//V3.POST("/robotAt", ding2.RobotAt)

	V3.GET("GetAllUsers", ding2.SelectAllUsers) // 查询所有用户信息
	V3.GET("upload", func(c *gin.Context) {
		username, _ := c.Get(global.CtxUserNameKey)
		c.File(fmt.Sprintf("Screenshot_%s.png", username))
	})
	/*=========具体业务路由==========*/
	Ding := r.Group("/api/ding")
	{
		//无需token验证
		Ding.POST("login", ding2.LoginHandler)
		Ding.POST("loginByDingDing", ding2.LoginByDingDing) // 判断钉钉扫码登陆
		Ding.POST("subscribeTo", ding2.SubscribeTo)         //钉钉订阅事件路由
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
