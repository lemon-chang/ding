package system

import (
	"ding/controllers/system"
	"github.com/gin-gonic/gin"
)

func InitAuthority(System *gin.RouterGroup) {
	authorityRouter := System.Group("authority")
	{
		authorityRouter.POST("createAuthority", system.CreateAuthority)       // 创建角色
		authorityRouter.POST("deleteAuthority", system.DeleteAuthority)       // 删除角色
		authorityRouter.PUT("updateAuthority", system.UpdateAuthority)        // 更新角色
		authorityRouter.POST("getAuthorityList", system.GetAuthorityList)     // 获取角色列表
		authorityRouter.POST("setUserAuthorities", system.SetUserAuthorities) // 设置用户角色
	}
}
