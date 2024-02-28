package system

import (
	"ding/controllers/system"
	"github.com/gin-gonic/gin"
)

func InitMenu(System *gin.RouterGroup) {
	Menu := System.Group("Menu")
	{
		Menu.POST("addMenu", system.AddBaseMenu)               // 新增菜单
		Menu.POST("deleteMenu", system.DeleteBaseMenu)         // 删除菜单
		Menu.POST("updateMenu", system.UpdateBaseMenu)         // 更新菜单
		Menu.GET("getMenuByToken", system.GetMenuByToken)      // 登陆后获取动态路由
		Menu.POST("getMenuById", system.GetMenuById)           // 根据id获取菜单具体信息
		Menu.POST("addMenuAuthority", system.AddMenuAuthority) // 增加menu和角色关联关系
		Menu.POST("getMenuAuthority", system.GetMenuAuthority) // 获取指定角色menu
		Menu.POST("getMenuList", system.GetMenuList)           // 分页获取基础menu列表

	}
}
