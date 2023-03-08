package ding

import (
	"ding/controllers"
	"ding/global"
	"ding/model/dingding"
	"ding/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

//addRobot 添加机器人
//思路如下：
//当前登录的用户添加了一个属于自己的机器人
func AddRobot(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if UserId == "" || err != nil {
		response.ResponseError(c, response.CodeLoginEror)
		return
	}
	user, err := (&dingding.DingUser{UserId: UserId}).GetUserByUserId()

	if err != nil {
		response.ResponseError(c, response.CodeLoginEror)
		return
	}
	//1.获取参数和参数校验
	var p *dingding.ParamAddRobot
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("Add Robot invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		}
		response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
		return
	}
	//说明插入的内部机器人
	if p.Type == "1" {
		//我们需要让用户扫码，添加群成员信息
	}
	// 2.逻辑处理
	dingRobot := &dingding.DingRobot{
		Type:               p.Type,
		RobotId:            p.RobotId,
		ChatBotUserId:      p.ChatBotUserId,
		Secret:             p.Secret,
		DingUsers:          p.DingUsers,
		UserName:           user.Name,
		ChatId:             p.ChatId,
		OpenConversationID: p.OpenConversationID,
		Name:               p.Name,
	}
	err = dingRobot.AddDingRobot()
	if err != nil {
		response.FailWithMessage("添加机器人失败", c)
	} else {
		response.OkWithDetailed(dingRobot, "添加机器人成功", c)
	}

}

func RemoveRobot(c *gin.Context) {
	var p dingding.ParamRemoveRobot
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("remove Robot invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		}
		response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
		return
	}
	err := (&dingding.DingRobot{}).RemoveRobot()
	if err != nil {
		response.FailWithMessage("移除机器人失败", c)
	} else {
		response.OkWithMessage("移除机器人成功", c)
	}
}

//GetRobots 获得用户自身的所有机器人
func GetRobots(c *gin.Context) {
	uid, err := global.GetCurrentUserId(c)
	if err != nil {
		return
	}
	//查询到所有的机器人
	robots, err := (&dingding.DingUser{UserId: uid}).GetRobotList()
	if err != nil {
		zap.L().Error("logic.GetRobotst() failed", zap.Error(err))
		response.ResponseError(c, response.CodeServerBusy) //不轻易把服务器的报错返回给外部
		return
	}

	response.ResponseSuccess(c, gin.H{
		"response": robots,
	})
	return
}
func UpdateRobot(c *gin.Context) {
	var p dingding.ParamUpdateRobot
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("UpdateRobot invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}
	}
	dingRobot := &dingding.DingRobot{
		RobotId:            p.RobotId,
		Type:               p.Type,
		ChatBotUserId:      p.ChatBotUserId,
		Secret:             p.Secret,
		DingUsers:          p.DingUsers,
		ChatId:             p.ChatId,
		OpenConversationID: p.OpenConversationID,
		Name:               p.Name,
	}
	err := (dingRobot).UpdateRobot()
	if err != nil {
		response.FailWithMessage("更新机器人失败", c)
	} else {
		response.OkWithDetailed(dingRobot, "更新机器人成功", c)
	}
}
