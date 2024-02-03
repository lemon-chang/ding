package system

import (
	system2 "ding/controllers/system"

	"github.com/gin-gonic/gin"
)

func SetupSystem(System *gin.RouterGroup) {
	System.GET("/", system2.WelcomeHandler)

	InitDataDictionary(System)
	InitMenu(System)
	InitAuthority(System)

}
