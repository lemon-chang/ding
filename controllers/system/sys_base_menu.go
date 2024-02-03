package system

import (
	"ding/global"
	"ding/model/common/response"
	"ding/model/system"
	"ding/utils"
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
func AddBaseMenu(c *gin.Context) {
	var menu system.SysBaseMenu
	err := c.ShouldBindJSON(&menu)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(menu, utils.MenuVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(menu.Meta, utils.MenuMetaVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = (&system.SysBaseMenu{}).AddBaseMenu(menu)
	if err != nil {
		zap.L().Error("添加失败!", zap.Error(err))
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

type ParamAddMenuAuthorityInfo struct {
	Menus       []system.SysBaseMenu `json:"menus"`
	AuthorityId uint                 `json:"authorityId"` // 角色ID
}

func AddMenuAuthority(c *gin.Context) {
	var authorityMenu ParamAddMenuAuthorityInfo
	err := c.ShouldBindJSON(&authorityMenu)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := utils.Verify(authorityMenu, utils.AuthorityIdVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := (&system.SysBaseMenu{}).AddMenuAuthority(authorityMenu.Menus, authorityMenu.AuthorityId); err != nil {
		zap.L().Error("添加失败!", zap.Error(err))
		response.FailWithMessage("添加失败", c)
	} else {
		response.OkWithMessage("添加成功", c)
	}
}
