package system

import (
	"ding/controllers/system"
	"github.com/gin-gonic/gin"
)

func InitMenu(System *gin.RouterGroup) {
	Menu := System.Group("Menu")
	{
		Menu.POST("addBaseMenu", system.AddBaseMenu)           // 新增菜单
		Menu.POST("addMenuAuthority", system.AddMenuAuthority) //	增加menu和角色关联关系
		//Menu.POST("deleteBaseMenu", system.DeleteBaseMenu)     // 删除菜单
		//Menu.POST("updateBaseMenu", system.UpdateBaseMenu)     // 更新菜单

		Menu.GET("getMenu", system.GetMenu)
	}
}
