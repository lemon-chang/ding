package system

import (
	"ding/global"
	"ding/model/common/request"
	"ding/model/common/response"
	"ding/model/system"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetMenuByToken(c *gin.Context) {
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
	if err := (&system.SysBaseMenu{}).AddMenuAuthority(authorityMenu.Menus, authorityMenu.AuthorityId); err != nil {
		zap.L().Error("添加失败!", zap.Error(err))
		response.FailWithMessage("添加失败", c)
	} else {
		response.OkWithMessage("添加成功", c)
	}
}
func DeleteBaseMenu(c *gin.Context) {
	var menu request.GetById
	err := c.ShouldBindJSON(&menu)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = (&system.SysBaseMenu{}).DeleteBaseMenu(menu.ID)
	if err != nil {
		zap.L().Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func UpdateBaseMenu(c *gin.Context) {
	var menu system.SysBaseMenu
	err := c.ShouldBindJSON(&menu)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = (&system.SysBaseMenu{}).UpdateBaseMenu(menu)
	if err != nil {
		zap.L().Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}
func GetMenuById(c *gin.Context) {
	var idInfo request.GetById
	err := c.ShouldBindJSON(&idInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	menu, err := (&system.SysBaseMenu{}).GetBaseMenuById(idInfo.ID)
	if err != nil {
		zap.L().Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(menu, "获取成功", c)
}
func GetMenuAuthority(c *gin.Context) {
	var param request.GetAuthorityId
	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	menus, err := (&system.SysBaseMenu{}).GetMenuAuthority(&param)
	if err != nil {
		zap.L().Error("获取失败!", zap.Error(err))
		response.FailWithDetailed(menus, "获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"menus": menus}, "获取成功", c)
}
func GetMenuList(c *gin.Context) {
	var pageInfo request.PageInfo
	err := c.ShouldBindJSON(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	menuList, total, err := (&system.SysBaseMenu{}).GetInfoList()
	if err != nil {
		zap.L().Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     menuList,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}
