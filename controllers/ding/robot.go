package ding

import (
	"context"
	"ding/global"
	"ding/model/common"
	response2 "ding/model/common/response"
	"ding/model/dingding"
	"ding/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
	UserId := global.GetCurrentUserId(c)
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
	uid := global.GetCurrentUserId(c)
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

func AddTask(c *gin.Context) {
	var p *dingding.ParamCronTask
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("CronTask做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	err, task := (&dingding.DingRobot{RobotId: p.RobotId}).CronSend(c, p)
	if err != nil {
		zap.L().Error("定时任务失败", zap.Error(err))
		response.FailWithMessage("创建定时任务失败", c)
	} else {
		response.OkWithDetailed(task, "创建定时任务成功", c)
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

func GetTaskList(c *gin.Context) {
	var p dingding.ParamGetTaskList
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("GetTaskList做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	tasks, total, err := (&dingding.Task{RobotId: p.RobotId}).GetTaskList(&p, c)
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
	var p *dingding.ParamRemoveTask
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("CronTask做定时任务参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	err := (&dingding.Task{}).Remove()
	if err != nil {
		zap.L().Error(fmt.Sprintf("移除定时任务失败"), zap.Error(err))
		response.FailWithMessage("移除定时任务失败", c)
	} else {
		response.OkWithMessage("移除定时任务成功", c)
	}
}
func UpdateTask(c *gin.Context) {
	var p *dingding.UpdateTask
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("UpdateTask参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	spec, detailTimeForUser, err := dingding.HandleSpec1(p)
	if err != nil {
		return
	}
	p.Spec = spec
	p.DetailTimeForUser = detailTimeForUser

	err = (&dingding.Task{Model: gorm.Model{ID: p.ID}}).UpdateTask(p)
	if err != nil {
		zap.L().Error("更新失败", zap.Error(err))
		response.FailWithMessage("更新失败", c)
	} else {
		response.OkWithMessage("更新成功", c)
	}
}

func GetTaskDetailByID(c *gin.Context) {
	var p *dingding.ParamGetTaskDeatil
	if err := c.ShouldBindQuery(&p); err != nil {
		zap.L().Error("GetTaskDetail参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
	}
	task := &dingding.Task{Model: gorm.Model{ID: p.ID}}
	err := task.GetTaskDetailByID()
	if err != nil {
		zap.L().Error(fmt.Sprintf("ReStartTask定时任务失败"), zap.Error(err))
		response.FailWithMessage("ReStartTask定时任务失败", c)
	} else {
		response.OkWithDetailed(task, "ReStartTask定时任务成功", c)
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
