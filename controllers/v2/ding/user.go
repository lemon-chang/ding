package ding

import (
	"ding/controllers"
	dingding2 "ding/model/dingding"
	"ding/model/params"
	"ding/model/params/ding"
	"ding/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"sync"
)

var wg = sync.WaitGroup{}

func ImportDingUserData(c *gin.Context) {
	var DingUser dingding2.DingUser
	t := dingding2.DingToken{}
	token, err := t.GetAccessToken()
	DingUser.DingToken.Token = token
	err = DingUser.ImportUserToMysql()
	if err != nil {
		response.ResponseErrorWithMsg(c, 0, gin.H{
			"message": err.Error(),
		})
		return
	}
	response.OkWithMessage("导入组织架构用户数据成功", c)

}

// SelectAllUsers 查询所有用户
func SelectAllUsers(c *gin.Context) {
	var DingUser dingding2.DingUser
	us, err := DingUser.FindDingUsers()
	if err != nil {
		response.FailWithMessage("查询用户失败", c)
		return
	}
	response.OkWithDetailed(us, "查询所有用户成功", c)
}

// UpdateDingUserAddr 更新用户博客&简书地址
func UpdateDingUserAddr(c *gin.Context) {
	var DingUser dingding2.DingUser
	var userParam ding.UserAndAddrParam
	if err := c.ShouldBindJSON(&userParam); err != nil {
		response.FailWithMessage("参数错误", c)
		zap.L().Error("参数错误", zap.Error(err))
		return
	}
	if err := DingUser.UpdateDingUserAddr(userParam); err != nil {
		zap.L().Error("更新用户博客和简书地址失败", zap.Error(err))
		response.FailWithMessage("更新用户博客&简书地址失败", c)
	}
	response.OkWithMessage("更新用户博客&简书地址成功", c)
}

func FindAllJinAndBlog(c *gin.Context) {
	var DingDept dingding2.DingDept
	list, err := DingDept.GetAllJinAndBlog()
	if err != nil {
		response.FailWithMessage("查询简书或者博客失败", c)
		return
	}
	response.OkWithDetailed(list, "查询简书或者博客成功", c)
}

// LoginHandler 处理登录请求的函数
func LoginHandler(c *gin.Context) {
	//1.获取请求参数及参数校验
	var p params.ParamLogin
	if err := c.ShouldBindJSON(&p); err != nil { //这个地方只能判断是不是json格式的数据
		zap.L().Error("Login with invalid param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}
	}
	//2.业务逻辑处理
	//3.返回响应
	user, err := (&dingding2.DingUser{Name: "闫佳鹏", Password: p.Password}).Login()
	if err != nil {
		response.FailWithMessage("用户登录失败", c)
	} else {
		response.OkWithDetailed(user, "用户登录成功", c)
	}
}

func GetQRCode(c *gin.Context) {
	buf, ChatID, title, err := (&dingding2.DingUser{}).GetQRCode(c)
	if err != nil {
		zap.L().Error("截取二维码和获取群聊基本错误", zap.Error(err))
		response.FailWithMessage("截取二维码和获取群聊基本错误", c)
	}
	token, err := (&dingding2.DingToken{}).GetAccessToken()
	if err != nil {
		zap.L().Error("获取token失败", zap.Error(err))
		return
	}
	openConversationID := (&dingding2.DingGroup{Token: dingding2.DingToken{Token: token}, ChatID: ChatID}).GetOpenConversationID()
	userIds, err := (&dingding2.DingRobot{DingToken: dingding2.DingToken{Token: token}, OpenConversationID: openConversationID}).GetGroupUserIds()
	result := struct {
		buf     []byte
		ChatId  string
		Title   string
		UserIds []string
	}{
		buf:     buf,
		ChatId:  ChatID,
		Title:   title,
		UserIds: userIds,
	}
	response.OkWithDetailed(result, "获取二维码成功", c)
}
