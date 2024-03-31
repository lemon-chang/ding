package ding

import (
	"context"
	"ding/global"
	"ding/initialize/viper"
	"ding/model/common"
	response2 "ding/model/common/response"
	"ding/model/dingding"
	"ding/response"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

func OutGoing(c *gin.Context) {
	var p dingding.ParamReveiver
	err := c.ShouldBindJSON(&p)
	err = c.ShouldBindHeader(&p)
	if err != nil {
		zap.L().Error("OutGoing invaild param", zap.Error(err))
		response.FailWithMessage("参数有误", c)
		return
	}
	err = (&dingding.DingRobot{}).SendSessionWebHook(&p)
	if err != nil {
		zap.L().Error("钉钉机器人回调出错", zap.Error(err))
		response.FailWithMessage("回调出错", c)
		return
	}
	response.OkWithMessage("回调成功", c)
}

func AddRobot(c *gin.Context) {
	//1.获取参数和参数校验
	var p dingding.ParamAddRobot
	err := c.ShouldBindJSON(&p)
	if err != nil {
		zap.L().Error("Add Robot invalid param", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = global.GLOAB_VALIDATOR.Struct(p)
	if err != nil {
		zap.L().Error("Add Robot invalid param", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	UserId, _ := global.GetCurrentUserId(c)
	user, _ := (&dingding.DingUser{UserId: UserId}).GetUserInfo()
	//说明插入的内部机器人
	dingRobot := &dingding.DingRobot{
		RobotId:    p.RobotId,
		DingUserID: UserId,
		UserName:   user.Name,
		Name:       p.Name,
		IsShared:   p.IsShared,
	}
	// 2.逻辑处理
	err = dingRobot.CreateOrUpdateRobot()
	//更新完之后，去修改定时任务里面的机器人名字
	if err != nil {
		response.FailWithMessage("添加机器人失败", c)
	} else {
		response.OkWithDetailed(dingRobot, "添加机器人成功", c)
	}
}
func GetSharedRobot(c *gin.Context) {
	var p dingding.ParamGetRobotList
	err := c.ShouldBindQuery(&p)
	if err != nil {
		zap.L().Error("Add Robot invaild param", zap.Error(err))
		response.FailWithMessage("参数有误", c)
		return
	}
	robot, total, err := (&dingding.DingRobot{}).GetSharedRobot(&p)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response2.PageResult{
		List:     robot,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, "获取成功", c)
}
func RemoveRobot(c *gin.Context) {
	var p dingding.ParamRemoveRobot
	var err error
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("remove Robot invaild param", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	robots := make([]dingding.DingRobot, len(p.RobotIds))
	for i := 0; i < len(robots); i++ {
		robots[i].RobotId = p.RobotIds[i]
	}
	err = (&dingding.DingRobot{}).RemoveRobots(robots)
	if err != nil {
		response.FailWithMessage("移除机器人失败", c)
	} else {
		response.OkWithMessage("移除机器人成功", c)
	}

}

// GetRobotList 获得用户自身的所有机器人
func GetRobotList(c *gin.Context) {
	var p dingding.ParamGetRobotList
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithMessage("参数有误", c)
		return
	}
	uid, err := global.GetCurrentUserId(c)
	//查询到所有的机器人
	robots, count, err := (&dingding.DingUser{UserId: uid}).GetRobotList(&p)
	if err != nil {
		zap.L().Error("logic.GetRobotst() failed", zap.Error(err))
		response.FailWithMessage("获取失败", c) //不轻易把服务器的报错返回给外部
		return
	}
	response.OkWithDetailed(response2.PageResult{
		List:     robots,
		Total:    count,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, "获取成功", c)
}

func CronTask(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		UserId = ""
	}
	CurrentUser, err := (&dingding.DingUser{UserId: UserId}).GetUserInfo()
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
	p := &dingding.ParamCronTask{}
	p.TaskName = "ping"
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage("参数有误", c)
		return
	}
	p = &dingding.ParamCronTask{
		MsgText:    &common.MsgText{Text: common.Text{Content: "机器人测试成功"}, At: common.At{}, Msgtype: "text"},
		RepeatTime: "立即发送",
		RobotId:    p.RobotId,
	}
	err, _ := (&dingding.DingRobot{RobotId: p.RobotId}).CronSend(c, p)

	if err != nil {
		zap.L().Error(fmt.Sprintf("测试机器人失败"), zap.Error(err))
		response.FailWithMessage("发送定时任务失败", c)
	} else {
		response.OkWithMessage("发送定时任务成功", c)
	}
}
func StopTask(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		UserId = ""
	}
	CurrentUser, err := (&dingding.DingUser{UserId: UserId}).GetUserInfo()
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
func GetRobotTaskList(c *gin.Context) {
	var p dingding.ParamGetTaskList
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("GetTaskList做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	tasks, total, err := (&dingding.DingRobot{RobotId: p.RobotId}).GetTaskList(&p)
	if err != nil {
		zap.L().Error(fmt.Sprintf("获取定时任务列表失败"), zap.Error(err))
		response.FailWithMessage("获取定时任务列表失败", c)
	} else {
		response.OkWithDetailed(response2.PageResult{
			List:     tasks,
			Total:    total,
			Page:     p.Page,
			PageSize: p.PageSize,
		}, "获取定时任务列表成功", c)
	}
}
func GetUserTaskList(c *gin.Context) {
	var p dingding.ParamGetTaskList
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("GetTaskList做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	UserId, _ := global.GetCurrentUserId(c)

	tasks, total, err := (&dingding.DingUser{UserId: UserId}).GetTaskList(&p)
	if err != nil {
		zap.L().Error(fmt.Sprintf("获取定时任务列表失败"), zap.Error(err))
		response.FailWithMessage("获取定时任务列表失败", c)
	} else {
		response.OkWithDetailed(response2.PageResult{
			List:     tasks,
			Total:    total,
			Page:     p.Page,
			PageSize: p.PageSize,
		}, "获取定时任务列表成功", c)
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
		response.FailWithMessage("CronTask做定时任务参数绑定失败", c)
		return
	}
	_, err = (&dingding.DingRobot{}).ReStartTask(p.ID)
	if err != nil {
		zap.L().Error(fmt.Sprintf("ReStartTask定时任务失败"), zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	} else {
		response.OkWithMessage("ReStartTask定时任务成功", c)
	}
}

// EditTaskContent 修改定时任务的内容
func EditTaskContent(c *gin.Context) {
	var r *dingding.EditTaskContentParam
	err := c.ShouldBindJSON(&r)
	if err != nil && r.TaskID == "" {
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
	//测试回调的时候使用
	callbackCrypto := dingding.NewDingTalkCrypto("marchSoft", "xoN8265gQVD4YXpcAPqV4LAm6nsvipEm1QiZoqlQslj", viper.Conf.MiniProgramConfig.RobotCode)
	//解密后的数据是一个json字符串
	decryptMsg, _ := callbackCrypto.GetDecryptMsg(signature, timestamp, nonce, m["encrypt"].(string))
	// 3. 反序列化回调事件json数据
	//把取值不方便的json字符串反序列化带map中
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(decryptMsg), &result)
	if err != nil {
		return
	}
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
	UserId       string `json:"user_id"`
	LeetCodeAddr string `json:"leetCodeAddr"`
}

func RobotAt(c *gin.Context) {
	var resp *dingding.RobotAtResp
	if err := c.ShouldBindJSON(&resp); err != nil {
		zap.L().Error("RobotAtResp", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	fmt.Println("内容为:", resp.Text)
	//userId := resp.SenderStaffId
	conversationType := resp.ConversationType               //聊天类型
	str := strings.TrimSpace(resp.Text["content"].(string)) //用户发给机器人的内容,去除前后空格
	dingRobot := &dingding.DingRobot{}

	//获得的字符串状态
	var staus int
	//单聊
	if conversationType == "1" {
		dataByStr, err := GetAllDataByStr(str, resp.SenderStaffId)
		if err != nil {
			zap.L().Error("GetAllDataByStr", zap.Error(err))
			response.FailWithMessage("通过字符串获取资源失败", c)
		}
		fmt.Println(dataByStr)
		for _, data := range dataByStr {
			if data.DataName == str {
				staus++
			}
		}
		//if staus != len(dataByStr) && len(dataByStr) != 1 {
		if staus != len(dataByStr) {
			//发送卡片信息
			fmt.Println("卡片")
			err := dingRobot.RobotSendCardToPerson(resp, dataByStr)
			if err != nil {
				zap.L().Error("RobotSendCardToPerson", zap.Error(err))
			}
		} else {
			//发送直接数据
			err := dingRobot.RobotSendMessageToPerson(resp, dataByStr)
			if err != nil {
				zap.L().Error("RobotSendMessageToPerson", zap.Error(err))
			}
		}

	} else if conversationType == "2" {
		if str == "打字邀请码" {
			//_ 代表res["processQueryKey"]可以查看已读状态
			_, err := dingRobot.RobotSendGroupInviteCode(resp)
			if err != nil {
				return
			}
		}
	}
}
func GetAllDataByStr(str string, userId string) (DatasByStr []dingding.Result, err error) {
	var AllDatas []dingding.Result
	//查询所有个人资源
	redisRoad := "learningData:personal:" + userId + ":"
	AllPersonalData, err := global.GLOBAL_REDIS.HGetAll(context.Background(), redisRoad).Result()
	if err != nil {
		zap.L().Error("从redis读取失败", zap.Error(err))
	}
	user, err := (&dingding.DingUser{UserId: userId}).GetUserInfo()
	if err != nil {
		zap.L().Error("userid查询用户信息失败", zap.Error(err))
	}
	for dataName, dataLink := range AllPersonalData {
		r := dingding.Result{
			Name:     user.Name,
			DataName: dataName,
			DataLink: dataLink,
		}
		AllDatas = append(AllDatas, r)
	}
	//查询所有公共资源
	redisRoad = "learningData:public*"
	allRedisRoad, err := global.GLOBAL_REDIS.Keys(context.Background(), redisRoad).Result()
	if err != nil {
		zap.L().Error("从redis读取公共数据失败", zap.Error(err))
		return
	}
	for _, s := range allRedisRoad {
		split := strings.Split(s, ":")
		userId := split[len(split)-1-1]
		user, err := (&dingding.DingUser{UserId: userId}).GetUserInfo()
		AllPublicData, err := global.GLOBAL_REDIS.HGetAll(context.Background(), s).Result()
		if err != nil {
			zap.L().Error("从redis读取失败", zap.Error(err))
		}
		for dataName, dataLink := range AllPublicData {
			r := dingding.Result{
				Name:     user.Name,
				DataName: dataName,
				DataLink: dataLink,
			}
			AllDatas = append(AllDatas, r)
		}
	}

	//查询此人所有部门内的所有资源
	//deptList := dingding.GetDeptByUserId(userId).DeptList
	token, _ := (&dingding.DingToken{}).GetAccessToken()
	DetailUser := &dingding.DingUser{UserId: userId, DingToken: dingding.DingToken{Token: token}}
	err = DetailUser.GetUserDetailByUserId()
	if err != nil {
		return
	}
	if DetailUser.Admin {
		var deptids []int
		global.GLOAB_DB.Model(dingding.DingDept{}).Select("dept_id").Scan(&deptids)
		DetailUser.DeptIdList = deptids
	}
	for _, dept := range DetailUser.DeptIdList {
		redisRoad = "learningData:dept:" + strconv.Itoa(dept) + ":*"
		allRedisRoad, err := global.GLOBAL_REDIS.Keys(context.Background(), redisRoad).Result()
		if err != nil {
			zap.L().Error("从redis读取公共数据失败", zap.Error(err))
		}
		for _, s := range allRedisRoad {
			split := strings.Split(s, ":")
			userId := split[len(split)-1-1]
			user, err := (&dingding.DingUser{UserId: userId}).GetUserInfo()
			AllPublicData, err := global.GLOBAL_REDIS.HGetAll(context.Background(), s).Result()
			if err != nil {
				zap.L().Error("从redis读取失败", zap.Error(err))
			}
			for dataName, dataLink := range AllPublicData {
				r := dingding.Result{
					Name:     user.Name,
					DataName: dataName,
					DataLink: dataLink,
				}
				AllDatas = append(AllDatas, r)
			}
		}
	}

	for _, data := range AllDatas {
		if str == data.DataName {
			DatasByStr = append(DatasByStr, data)
		} else if strings.Contains(data.DataName, str) {
			DatasByStr = append(DatasByStr, data)
		}
	}
	return
}

type Data struct {
	Type        int    `json:"type"` //1公共 2部门 3个人
	DeptId      int    `json:"dept_id"`
	OldDataName string `json:"old_data_name"`
	DataName    string `json:"data_name"`
	DataLink    string `json:"data_link"`
	UserName    string `json:"user_name"`
	Url         string `json:"url"`
}
