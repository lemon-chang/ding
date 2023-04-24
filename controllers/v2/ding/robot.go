package ding

import (
	"ding/controllers"
	"ding/global"
	"ding/model/common"
	"ding/model/dingding"
	"ding/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"log"
)

func OutGoing(c *gin.Context) {
	var p dingding.ParamReveiver
	err := c.ShouldBindJSON(&p)
	err = c.ShouldBindHeader(&p)
	if err != nil {
		zap.L().Error("OutGoing invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}
	}

	err = dingding.SendSessionWebHook(&p)
	if err != nil {
		zap.L().Error("钉钉机器人回调出错", zap.Error(err))
		response.ResponseErrorWithMsg(c, response.CodeServerBusy, "钉钉机器人回调出错")
		return
	}
	response.ResponseSuccess(c, "回调成功")

}

// addRobot 添加机器人
// 思路如下：
// 当前登录的用户添加了一个属于自己的机器人
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
	err = c.ShouldBindJSON(&p)
	if err != nil {
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
	dingRobot := &dingding.DingRobot{
		Type:       p.Type,
		RobotId:    p.RobotId,
		Secret:     p.Secret,
		DingUserID: UserId,
		UserName:   user.Name,
		Name:       p.Name,
	}
	if p.Type == "" {
		zap.L().Info("前端没有告知机器人的类型,我们前往数据库进行查询")
		err = global.GLOAB_DB.Where("robot_id = ?", p.RobotId).First(&dingRobot).Error
		p.Type = dingRobot.Type
		zap.L().Info(fmt.Sprintf("%v对应的机器人姓名为：%v,type = %v", p.RobotId, p.Name, p.Type))
	}
	if p.Type == "1" {
		//我们需要让用户扫码，添加群成员信息
		//此处展示二维码
		//_, ChatID, title, err := (&dingding.DingUser{}).GetQRCode(c)
		if err != nil {
			zap.L().Error("截取二维码和获取群聊基本错误", zap.Error(err))
		}
		//dingRobot.ChatId = ChatID
		dingRobot.Name = p.Name
		token, err1 := (&dingding.DingToken{}).GetAccessToken()
		if err1 != nil {
			zap.L().Error("获取token失败", zap.Error(err))
			return
		}
		openConversationID := (&dingding.DingGroup{Token: dingding.DingToken{Token: token}}).GetOpenConversationID()
		dingRobot.OpenConversationID = openConversationID
		userIds, err := (&dingding.DingRobot{DingToken: dingding.DingToken{Token: token}, OpenConversationID: openConversationID}).GetGroupUserIds()
		var users []dingding.DingUser

		err2 := global.GLOAB_DB.Where("user_id in ?", userIds).Find(&users).Error
		if err2 != nil {
			zap.L().Error(fmt.Sprintf("根据userids查询users失败"), zap.Error(err2))
		}
		global.GLOAB_DB.Model(&dingRobot)
		//dingRobot.DingUsers = users
		err = global.GLOAB_DB.Model(dingRobot).Association("DingUsers").Replace(users)
		if err != nil {
			zap.L().Error("global.GLOAB_DB.Model(dingRobot).Association(\"DingUsers\").Replace(users)有误", zap.Error(err))
		}
	} else if p.Type == "2" {
		//直接更新即可
	}
	// 2.逻辑处理
	err = dingRobot.CreateOrUpdateRobot()
	if err != nil {
		response.FailWithMessage("添加机器人失败", c)
	} else {
		response.OkWithDetailed(dingRobot, "添加机器人成功", c)
	}

}
func GetRobotDetailByRobotId(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if UserId == "" || err != nil {
		response.ResponseError(c, response.CodeLoginEror)
		return
	}
	//1.获取参数和参数校验
	var p *dingding.ParamGetRobotBase
	err = c.ShouldBindQuery(&p)
	if err != nil {
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
	dingRobot := &dingding.DingRobot{
		RobotId: p.RobotId,
	}
	err = global.GLOAB_DB.Where("robot_id = ? and ding_user_id = ?", p.RobotId, UserId).Preload("DingUsers").Preload("Tasks").First(dingRobot).Error
	if err != nil {
		zap.L().Error("通过机器人id和所属用户id查询机器人基本信息失败", zap.Error(err))
		response.FailWithMessage("获取机器人信息失败", c)
	} else {
		response.OkWithDetailed(dingRobot, "获取机器人信息成功", c)
	}
}
func GetRobotBaseList(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if UserId == "" || err != nil {
		response.ResponseError(c, response.CodeLoginEror)
		return
	}
	//1.获取参数和参数校验
	var p *dingding.ParamGetRobotListBase
	err = c.ShouldBindQuery(&p)
	if err != nil {
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
	var dingRobotList []dingding.DingRobot
	err = global.GLOAB_DB.Where("ding_user_id = ?", UserId).Find(&dingRobotList).Error
	if err != nil {
		zap.L().Error(fmt.Sprintf("获取用户%v拥有的所有机器人列表基本信息失败", UserId), zap.Error(err))
		response.FailWithMessage("获取机器人列表基本信息失败", c)
	} else {
		response.OkWithDetailed(dingRobotList, "获取机器人列表基本信息失败", c)
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
	err := (&dingding.DingRobot{RobotId: p.RobotId}).RemoveRobot()
	if err != nil {
		response.FailWithMessage("移除机器人失败", c)
	} else {
		response.OkWithMessage("移除机器人成功", c)
	}
}

// GetRobots 获得用户自身的所有机器人
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
	err := (dingRobot).CreateOrUpdateRobot()
	if err != nil {
		response.FailWithMessage("更新机器人失败", c)
	} else {
		response.OkWithDetailed(dingRobot, "更新机器人成功", c)
	}
}

func ChatHandler(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		UserId = "453562553921462447"
	}
	CurrentUser, err := (&dingding.DingUser{UserId: UserId}).GetUserByUserId()
	if err != nil {
		CurrentUser = dingding.DingUser{}
	}
	var p *dingding.ParamChat
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("ChatHandler做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	err = (&dingding.DingRobot{RobotId: "dingepndjqy7etanalhi"}).ChatSendMessage(p)

	if err != nil {
		zap.L().Error(fmt.Sprintf("使用机器人发送定时任务失败，发送人：%v,发送人id:%v", CurrentUser.Name, CurrentUser.UserId), zap.Error(err))
		response.FailWithDetailed(err, "发送定时任务失败", c)
	} else {
		response.OkWithMessage("发送定时任务成功", c)
	}
}
func CronTask(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		UserId = ""
	}
	CurrentUser, err := (&dingding.DingUser{UserId: UserId}).GetUserByUserId()
	if err != nil {
		CurrentUser = dingding.DingUser{}
	}
	var p *dingding.ParamCronTask
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("CronTask做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	err, task := (&dingding.DingRobot{RobotId: p.RobotId}).CronSend(c, p)

	if err != nil {
		zap.L().Error(fmt.Sprintf("使用机器人发送定时任务失败，发送人：%v,发送人id:%v", CurrentUser.Name, CurrentUser.UserId), zap.Error(err))
		response.FailWithMessage("发送定时任务失败", c)
	} else {
		response.OkWithDetailed(task, "发送定时任务成功", c)
	}
}
func PingRobot(c *gin.Context) {
	var p *dingding.ParamCronTask
	p = &dingding.ParamCronTask{
		MsgText:    &common.MsgText{Text: common.Text{Content: "机器人测试成功"}, At: common.At{}, Msgtype: "text"},
		RepeatTime: "立即发送",
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("CronTask做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		UserId = ""
	}
	CurrentUser, err := (&dingding.DingUser{UserId: UserId}).GetUserByUserId()
	if err != nil {
		CurrentUser = dingding.DingUser{}
	}

	err, task := (&dingding.DingRobot{RobotId: p.RobotId}).CronSend(c, p)

	r := struct {
		taskName string `json:"task_name"` //任务名字
		taskId   int    `json:"task_id"`
	}{
		taskName: task.TaskName,
	}

	if err != nil {
		zap.L().Error(fmt.Sprintf("测试机器人失败，发送人：%v,发送人id:%v", CurrentUser.Name, CurrentUser.UserId), zap.Error(err))
		response.FailWithMessage("发送定时任务失败", c)
	} else {
		response.OkWithDetailed(r, "发送定时任务成功", c)
	}
}

func StopTask(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		UserId = ""
	}
	CurrentUser, err := (&dingding.DingUser{UserId: UserId}).GetUserByUserId()
	if err != nil {
		CurrentUser = dingding.DingUser{}
	}
	var p *dingding.ParamStopTask
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("暂停定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	err = (&dingding.DingRobot{}).StopTask(p.TaskID)

	if err != nil {
		zap.L().Error(fmt.Sprintf("暂停定时任务失败，发送人：%v,发送人id:%v", CurrentUser.Name, CurrentUser.UserId), zap.Error(err))
		response.FailWithMessage("暂停定时任务失败", c)
	} else {
		response.OkWithMessage("暂停定时任务成功", c)
	}
}
func GetTaskList(c *gin.Context) {
	//UserId, err := global.GetCurrentUserId(c)
	//if err != nil {
	//	UserId = ""
	//}
	//CurrentUser, err := (&dingding.DingUser{UserId: UserId}).GetUserByUserId()
	//if err != nil {
	//	CurrentUser = dingding.DingUser{}
	//}
	var p *dingding.ParamGetTaskList
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("GetTaskList做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	tasks, err := (&dingding.DingRobot{}).GetTaskList(p.RobotId)
	if err != nil {
		zap.L().Error(fmt.Sprintf("获取定时任务列表失败"), zap.Error(err))
		response.FailWithMessage("获取定时任务列表失败", c)
	} else {
		response.OkWithDetailed(tasks, "获取定时任务列表成功", c)
	}

}
func RemoveTask(c *gin.Context) {

	var p *dingding.ParamStopTask
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("CronTask做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	err := (&dingding.DingRobot{}).RemoveTask(p.TaskID)
	if err != nil {
		zap.L().Error(fmt.Sprintf("移除定时任务失败"), zap.Error(err))
		response.FailWithMessage("移除定时任务失败", c)
	} else {
		response.OkWithMessage("移除定时任务成功", c)
	}
}
func ReStartTask(c *gin.Context) {
	var p *dingding.ParamRestartTask
	err := c.ShouldBindJSON(&p)
	if err != nil || p.ID == "" {
		zap.L().Error("CronTask做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
	}
	_, err = (&dingding.DingRobot{}).ReStartTask(p.ID)
	if err != nil {
		zap.L().Error(fmt.Sprintf("ReStartTask定时任务失败"), zap.Error(err))
		response.FailWithMessage(err.Error(), c)
	} else {
		response.OkWithMessage("ReStartTask定时任务成功", c)
	}

}
func GetTaskDetail(c *gin.Context) {
	var p *dingding.ParamGetTaskDeatil
	if err := c.ShouldBindQuery(&p); err != nil {
		zap.L().Error("GetTaskDetail参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	task, err := (&dingding.DingRobot{}).GetUnscopedTaskByID(p.TaskID)
	if err != nil {
		zap.L().Error(fmt.Sprintf("ReStartTask定时任务失败"), zap.Error(err))
		response.FailWithMessage("ReStartTask定时任务失败", c)
	} else {
		response.OkWithDetailed(task, "ReStartTask定时任务成功", c)
	}
}

//	func SubscribeTo(c *gin.Context) {
//		p := dingding.ParamCronTask{
//			MsgText: &common.MsgText{
//				At: common.At{
//					IsAtAll: true,
//				},
//				Text: common.Text{
//					Content: "subscription start",
//				},
//				Msgtype: "text",
//			},
//			RepeatTime: "立即发送",
//			TaskName:   "事件订阅",
//		}
//
//		err, task := (&dingding.DingRobot{RobotId: "2e36bf946609cd77206a01825273b2f3f33aed05eebe39c9cc9b6f84e3f30675"}).CronSend(c, &p)
//		if err != nil {
//			response.FailWithMessage("获取消息订阅信息失败，详情联系后端", c)
//			return
//		}
//		fmt.Println(task)
//		response.OkWithMessage("获取消息订阅成功", c)
//	}
func SubscribeTo(c *gin.Context) {
	var ding = dingding.NewDingTalkCrypto("KLkA8WdUV1fJfBN3KxEh6FNxPinwGdC6s7FIPro8LvxYRe37yvgl", "MyOhDfHxAlrzLjBLY6LVR26w8NrPEopY5U8GPDLntp2", "dingepndjqy7etanalhi")
	msg, _ := ding.GetEncryptMsg("success")
	log.Printf("msg: %v\n", msg)
	//success, _ := ding.GetDecryptMsg("111108bb8e6dbc2xxxx", "1783610513", "380320111", "rlmRqtlLfm7tTAM8fTim3WNSwyWbd-KM3wTZ8wBtwKX8Pw6M4ZzEiIQVrCqKgCwu")
	//log.Printf("success: %v\n", success)
}
