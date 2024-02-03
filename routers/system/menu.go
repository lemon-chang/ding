package system

import (
	"ding/controllers/system"
	"github.com/gin-gonic/gin"
)

func InitMenu(System *gin.RouterGroup) {
	Menu := System.Group("Menu")
	{
		Menu.GET("getMenu", system.GetMenu)
	}
}
