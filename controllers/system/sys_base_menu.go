package system

import (
	"ding/global"
	"ding/model/common/response"
	"ding/model/system"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetMenu(c *gin.Context) {
	AuthorityId, _ := c.Get(global.CtxUserAuthorityIDKey)
	menus, err := (&system.SysBaseMenu{}).GetMenuTree(AuthorityId.(uint))
	if err != nil {
		zap.L().Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
	}
	if menus == nil {
		menus = []system.SysMenu{}
	}
	response.OkWithDetailed(menus, "获取成功", c)

}
