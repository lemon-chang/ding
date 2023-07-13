package ding

import (
	"ding/controllers"
	"ding/global"
	"ding/model/common"
	"ding/model/dingding"
	"ding/response"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"strings"
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

	err = (&dingding.DingRobot{}).SendSessionWebHook(&p)
	if err != nil {
		zap.L().Error("钉钉机器人回调出错", zap.Error(err))
		response.ResponseErrorWithMsg(c, response.CodeServerBusy, "钉钉机器人回调出错")
		return
	}
	response.ResponseSuccess(c, "回调成功")
}
func GxpRobot(c *gin.Context) {
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
	err = (&dingding.DingRobot{}).GxpSendSessionWebHook(&p)
	if err != nil {
		zap.L().Error("钉钉机器人回调出错", zap.Error(err))
		return
	}
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
		IsShared:   p.IsShared,
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
func GetSharedRobot(c *gin.Context) {
	robot, err := (&dingding.DingRobot{}).GetSharedRobot()
	if err != nil {

	}
	response.OkWithDetailed(robot, "获取成功", c)
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
	var err error
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("remove Robot invaild param", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}

	go func() {
		consumerMsgs, err := global.GLOBAL_Kafka_Cons.ConsumePartition("delete-topic", 1, sarama.OffsetNewest)
		if err != nil {
			fmt.Println(err)
			zap.L().Error("kafka consumer msg failed ...")
			return
		}
		for msg := range consumerMsgs.Messages() {
			id := msg.Value
			err = (&dingding.DingRobot{RobotId: string(id)}).RemoveRobot()
			if err != nil {
				break
			}
		}
		if err != nil {
			response.FailWithMessage("移除机器人失败 kafka消息消费失败", c)
		} else {
			response.OkWithMessage("移除机器人成功 kafka消息消费失败", c)
		}
	}()

	for i := 0; i < len(p.RobotIds); i++ {
		if _, _, err := global.GLOBAL_Kafka_Prod.SendMessage(global.KafMsg("delete-topic", p.RobotIds[i], 1)); err != nil {
			zap.L().Error("kafka produce msg failed ... ")
			return
		}
	}

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
		zap.L().Error("获取uid失败", zap.Error(err))
		response.FailWithMessage("身份获取失败", c)
		return
	}
	//查询到所有的机器人
	robots, err := (&dingding.DingUser{UserId: uid}).GetRobotList()
	if err != nil {
		zap.L().Error("logic.GetRobotst() failed", zap.Error(err))
		response.ResponseError(c, response.CodeServerBusy) //不轻易把服务器的报错返回给外部
		return
	}

	response.ResponseSuccess(c, robots)
	//return
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
func B(c *gin.Context) {
	var users []dingding.DingUser
	var userids []string
	//global.GLOAB_DB.Model(&dingding.DingDept{}).Preload("")
	global.GLOAB_DB.Table("user_dept").Where("is_responsible = ? and ding_dept_dept_id = ?", true, 546623914).Select("ding_user_user_id").Find(&userids)
	global.GLOAB_DB.Model(&dingding.DingUser{}).Where("user_id IN ?", userids).Find(&users)
	response.ResponseSuccess(c, users)
	//var userids []string
	//deptid := 546623914
	//global.GLOAB_DB.Table("user_dept").Where("is_responsible = ? and ding_dept_dept_id = ?", true, deptid).Select("ding_user_user_id").Find(&userids)
	//p := &dingding.ParamChat{
	//	RobotCode: "dingepndjqy7etanalhi",
	//	UserIds:   userids,
	//	MsgKey:    "sampleText",
	//	MsgParam:  "测试单聊消息的发送",
	//}
	//err := (&dingding.DingRobot{}).ChatSendMessage(p)
	//if err != nil {
	//	fmt.Println("wochucuol:", err)
	//}
}

//修改定时任务的内容
func EditTaskContent(c *gin.Context) {
	var r *dingding.EditTaskContentParam
	err := c.ShouldBindJSON(&r)
	if err != nil && r.TaskID == 0 {
		zap.L().Error("编辑定时任务内容参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	err = (&dingding.DingRobot{}).EditTaskContent(r)
	if err != nil {
		zap.L().Error("编辑失败", zap.Error(err))
		response.FailWithMessage("编辑失败", c)
		return
	}
	response.OkWithMessage("修改成功", c)
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

// 获取所有的公共机器人
func GetAllPublicRobot(c *gin.Context) {
	robots, err := dingding.GetAllPublicRobot()
	if err != nil {
		zap.L().Error("查询所有公共机器人失败", zap.Error(err))
		response.FailWithMessage("获取机器人失败", c)
	} else if len(robots) == 0 {
		response.FailWithMessage("没有公共机器人", c)
	} else {
		//返回机器人的基本信息
		response.ResponseSuccess(c, robots)
	}
}
func AlterResultByRobot(c *gin.Context) {
	var p *dingding.ParamAlterResultByRobot
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("AlterResultByRobot参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	err := dingding.AlterResultByRobot(p)
	if err != nil {
		zap.L().Error("AlterResultByRobot失败", zap.Error(err))
		response.FailWithMessage("更新失败", c)
	}
	response.ResponseSuccess(c, "更新成功")
}

// 进行单聊
func SingleChat(c *gin.Context) {
	var p dingding.ParamChat
	err := c.ShouldBindJSON(&p)
	if err != nil {
	}
	err = (&dingding.DingRobot{}).ChatSendMessage(&p)
}
func SubscribeTo(c *gin.Context) {
	// 1. 参数获取
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	zap.L().Info(fmt.Sprintf("signature: " + signature + ", timestamp: " + timestamp + ", nonce: " + nonce))
	var m map[string]interface{}
	if err := c.ShouldBindJSON(&m); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	zap.L().Info(fmt.Sprintf("encrypt: %v", m))

	// 2. 参数解密
	callbackCrypto := dingding.NewDingTalkCrypto("marchSoft", "xoN8265gQVD4YXpcAPqV4LAm6nsvipEm1QiZoqlQslj", "dingepndjqy7etanalhi")
	//解密后的数据是一个json字符串
	decryptMsg, _ := callbackCrypto.GetDecryptMsg(signature, timestamp, nonce, m["encrypt"].(string))
	// 3. 反序列化回调事件json数据
	//把取值不方便的json字符串反序列化带map中
	result := make(map[string]interface{})
	json.Unmarshal([]byte(decryptMsg), &result)
	//事件类型
	eventType := result["EventType"].(string)
	subscription := dingding.NewDingSubscribe(result)

	// 4.根据EventType分类处理
	if eventType == "check_url" {
		// 测试回调url的正确性
		zap.L().Info("测试回调url的正确性\n")
	} else if eventType == "chat_add_member" {
		// 处理通讯录用户增加事件
		zap.L().Info("发生了：" + eventType + "事件")
		subscription.UserAddOrg(c)
	} else if eventType == "chat_remove_member" {
		// 处理通讯录用户减少事件
		zap.L().Info("发生了：" + eventType + "事件")
		subscription.UserLeaveOrg(c)
	} else if eventType == "check_in" {
		// 用户签到事件
		subscription.CheckIn(c)
	} else if eventType == "bpms_instance_change" {
		title := result["title"].(string)
		s := result["type"].(string)
		if strings.Contains(title, "请假") && s == "start" {
			fmt.Println("123456")
			//c.Get(global.CtxUserIDKey) 是通过用户登录后生成的token 中取到 user_id
			//c.Query("user_id")  是取前端通过 发来的params参数中的 user_id字段
			subscription.Leave(result)
		} else {

		}
	} else {
		// 添加其他已注册的
		zap.L().Info("发生了：" + eventType + "事件")
	}

	// 5. 返回success的加密数据
	successMap, _ := callbackCrypto.GetEncryptMsg("success")
	c.JSON(http.StatusOK, successMap)
}

type ParamLeetCodeAddr struct {
	Name         string `json:"name"`
	LeetCodeAddr string `json:"leetCodeAddr"`
}

func GetLeetCode(c *gin.Context) {
	var leetcode ParamLeetCodeAddr
	err := c.ShouldBindJSON(&leetcode)
	if err != nil {
		zap.L().Error("LeetCodeResp", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	fmt.Println(leetcode)
	err = global.GLOAB_DB.Table("ding_users").Where("name = ?", leetcode.Name).Update("leet_code_addr", leetcode.LeetCodeAddr).Error
	if err != nil {
		zap.L().Error("存入数据库失败", zap.Error(err))
		response.FailWithMessage("存入数据库失败", c)
		return
	}
	response.ResponseSuccess(c, "成功")
}

func RobotAt(c *gin.Context) {
	var resp *dingding.RobotAtResp
	if err := c.ShouldBindJSON(&resp); err != nil {
		zap.L().Error("RobotAtResp", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	fmt.Println("内容为:", resp.Text)
	//userId := resp.SenderStaffId
	conversationType := resp.ConversationType               //群聊id
	str := strings.TrimSpace(resp.Text["content"].(string)) //用户发给机器人的内容,去除前后空格
	dingRobot := &dingding.DingRobot{}
	//单聊
	if conversationType == "1" {
		if str == "打字邀请码" {
			err := dingRobot.RobotSendInviteCode(resp)
			if err != nil {
				return
			}
		} else if str == "送水电话号码" {
			err := dingRobot.RobotSendWater(resp)
			if err != nil {
				return
			}
		} else if str == "获取个人信息" {
			err := dingRobot.RobotSendPrivateMessage(resp)
			if err != nil {
				return
			}
		} else if strings.Contains(str, "保存个人信息") {
			err := dingRobot.RobotSavePrivateMessage(resp)
			if err != nil {
				return
			}
		} else if strings.Contains(str, "更改个人信息") {
			err := dingRobot.RobotPutPrivateMessage(resp)
			if err != nil {
				return
			}
		} else if str == "学习资源" {

		} else {
			err := dingRobot.RobotSendHelpCard(resp)
			if err != nil {
				return
			}
		}
		//群聊
	} else if conversationType == "2" {
		if str == "打字邀请码" {
			//_ 代表res["processQueryKey"]可以查看已读状态
			_, err := dingRobot.RobotSendGroupInviteCode(resp)
			if err != nil {
				return
			}
		} else if str == "送水电话号码" {
			//_ 代表res["processQueryKey"]可以查看已读状态
			_, err := dingRobot.RobotSendGroupWater(resp)
			if err != nil {
				return
			}
		} else if str == "帮助" {
			//_ 代表res["processQueryKey"]可以查看已读状态
			_, err := dingRobot.RobotSendGroupCard(resp)
			if err != nil {
				return
			}
		}
	}
	response.ResponseSuccess(c, "成功")
}

//上传个人资源
func UpdatePersonalDate(c *gin.Context) {

	response.ResponseSuccess(c, "成功")
}

//上传部门资源
func UpdateDeptDate(c *gin.Context) {

	response.ResponseSuccess(c, "成功")
}

//上传公共资源
func UpdatePublicDate(c *gin.Context) {

	response.ResponseSuccess(c, "成功")
}
