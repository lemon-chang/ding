package ding

import (
	"context"
	"crypto/tls"
	"ding/controllers"
	"ding/dao/redis"
	"ding/global"
	"ding/initialize/jwt"
	dingding2 "ding/model/dingding"
	"ding/model/params"
	"ding/model/params/ding"
	"ding/response"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
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
	name := c.Query("name")
	mobile := c.Query("mobile")
	var DingUser dingding2.DingUser
	us, err := DingUser.FindDingUsers(name, mobile)
	if err != nil {
		response.FailWithMessage("查询用户失败", c)
		return
	}
	response.OkWithDetailed(us, "查询所有用户成功", c)
}
func GetUserInfo(c *gin.Context) {
	user_id, _ := c.Get(global.CtxUserIDKey)
	DingUser := dingding2.DingUser{}
	DingUser.UserId = user_id.(string)
	err := DingUser.GetUserInfo()
	if err != nil {
		response.FailWithMessage("查询用户失败", c)
		return
	}
	response.OkWithDetailed(DingUser, "查询所有用户成功", c)
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

//LoginHandler 处理登录请求的函数
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
	user, err := (&dingding2.DingUser{Mobile: p.Mobile, Password: p.Password}).Login()
	// 生成JWT
	token, err := jwt.GenToken(c, user)
	if err != nil {
		zap.L().Debug("JWT生成错误")
		return
	}
	user.AuthToken = token
	if err != nil {
		response.FailWithMessage("用户登录失败", c)
	} else {
		response.OkWithDetailed(user, "用户登录成功", c)
	}
}
func LoginHandlerByToken(c *gin.Context) {
	//1.获取请求参数及参数校验
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		response.ResponseError(c, response.CodeNeedLogin)
		return
	}
	//调用oss的接口，来进行登录认证
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	//URL := "https://oapi.dingtalk.com/attendance/listRecord?access_token=" + a.DingToken.Token
	URL := "http://127.0.0.1:8890/marchsoft/getUserInfo"
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}

	//然后把结构体对象序列化一下
	//然后就可以放入具体的request中的
	request, err := http.NewRequest(http.MethodPost, URL, nil)
	request.Header.Set("Authorization", authHeader)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	zap.L().Info(fmt.Sprintf("发送请求成功，原始resp为:%v", resp))
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		Code int `json:"code"`
		Data struct {
			UserId string `json:"userid"`
			Name   string `json:"name"`
			Mobile string `json:"mobile"`
		} `json:"data"`
		Msg string `json:"msg"`
	}{}

	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	zap.L().Info(fmt.Sprintf("把请求结果序列化到结构体对象中成功%v", r))
	if r.Code != 0 {
		response.FailWithMessage("登录失败", c)
		return
	}
	response.OkWithDetailed(r, "登录成功", c)

}

func GetQRCode(c *gin.Context) {
	buf, ChatID, title, err := (&dingding2.DingUser{}).GetQRCode1(c)
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

//获取所有任务列表,包括已暂停的任务
func GetAllTask(c *gin.Context) {
	var tasks []dingding2.Task
	err := global.GLOAB_DB.Model(&tasks).Unscoped().Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").Find(&tasks).Error
	if err != nil {
		zap.L().Error("获取所有定时任务失败", zap.Error(err))
		response.FailWithMessage("服务繁忙", c)
		return
	}
	response.ResponseSuccess(c, tasks)
}

func GetAllActiveTask(c *gin.Context) {
	//先删除所有的任务，然后再重新加载一遍
	activeTasksKeys, err := global.GLOBAL_REDIS.Keys(context.Background(), fmt.Sprintf("%s*", redis.Perfix+redis.ActiveTask)).Result()
	if err != nil {
		zap.L().Error("从redis中获取旧的活跃任务的key失败", zap.Error(err))
		return
	}
	//删除所有的key
	global.GLOBAL_REDIS.Del(context.Background(), activeTasksKeys...)

	//拿到所有的任务的id
	//entries := global.GLOAB_CORN.Entries()
	//拿到所有任务的id
	//var entriesInt = make([]int, len(entries))
	//for index, value := range entries {
	//	entriesInt[index] = int(value.ID)
	//}
	// 根据id查询数据库，拿到详细的任务信息，存放到redis中
	var tasks []dingding2.Task //拿到所有的活跃任务
	global.GLOAB_DB.Model(&tasks).Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").Where("deleted_at is null").Find(&tasks)
	//查询所有的在线任务
	//把找到的数据存储到redis中 ，现在先写成手动获取
	//应该是存放在一个集合里面，集合里面存放着此条任务的所有信息，以id作为标识
	//哈希特别适合存储对象，所以我们用哈希来存储

	for _, task := range tasks {
		taskValue, err := json.Marshal(task) //把对象序列化成为一个json字符串
		if err != nil {
			return
		}
		err = global.GLOBAL_REDIS.Set(context.Background(), redis.GetTaskKey(task.TaskID), string(taskValue), 0).Err()
		if err != nil {
			zap.L().Error(fmt.Sprintf("从mysql获取所有活跃任务存入redis失败，失败任务id：%s，任务名：%s,执行人：%s,对应机器人：%s", task.TaskID, task.TaskName, task.UserName, task.RobotName), zap.Error(err))
			return
		}
	}
	zap.L().Info("获取所有获取定时任务成功")
	response.OkWithDetailed(tasks, "获取所有获取定时任务成功", c)
}
func MakeupSign(c *gin.Context) {
	var p params.ParamMakeupSign
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage("参数错误", c)
		zap.L().Error("参数错误", zap.Error(err))
		return
	}
	consecutiveSignNum, err := (&dingding2.DingUser{UserId: p.Userid}).Sign(p.Year, p.UpOrDown, p.StartWeek, p.WeekDay, p.MNE)
	if err != nil {
		zap.L().Error("为用户补签失败", zap.Error(err))
		response.FailWithMessage("为用户补签失败", c)
		return
	}
	response.OkWithDetailed(consecutiveSignNum, "补签成功", c)
}
func GetWeekConsecutiveSignNum(c *gin.Context) {
	var p params.ParamGetWeekConsecutiveSignNum
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage("参数错误", c)
		zap.L().Error("参数错误", zap.Error(err))
		return
	}
	consecutiveSignNum, err := (&dingding2.DingUser{UserId: p.Userid}).GetConsecutiveSignNum(p.Year, p.UpOrDown, p.StartWeek, p.WeekDay, p.MNE)
	if err != nil {
		zap.L().Error("获取用户本周连续签到次数失败", zap.Error(err))
		response.FailWithMessage("获取用户本周连续签到次数失败", c)
		return
	}
	response.OkWithDetailed(consecutiveSignNum, "获取用户本周连续签到次数成功", c)
}
func GetWeekSignNum(c *gin.Context) {
	var p params.ParamGetWeekSignNum
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage("参数错误", c)
		zap.L().Error("参数错误", zap.Error(err))
		return
	}
	WeekSignNum, err := (&dingding2.DingUser{UserId: p.Userid}).GetWeekSignNum(p.Year, p.UpOrDown, p.StartWeek)
	if err != nil {
		zap.L().Error("获取用户一周的签到次数失败", zap.Error(err))
		response.FailWithMessage("获取用户一周的签到次数失败", c)
		return
	}
	response.OkWithDetailed(WeekSignNum, "获取用户一周的签到次数成功", c)
}
func GetWeekSignDetail(c *gin.Context) {
	var p params.ParamGetWeekSignDetail
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage("参数错误", c)
		zap.L().Error("参数错误", zap.Error(err))
		return
	}
	WeekSignNum, err := (&dingding2.DingUser{UserId: p.Userid}).GetWeekSignDetail(p.Year, p.UpOrDown, p.StartWeek)
	if err != nil {
		zap.L().Error("获取用户一周的签到详情失败", zap.Error(err))
		response.FailWithMessage("获取用户一周的签到详情失败", c)
		return
	}
	response.OkWithDetailed(WeekSignNum, "获取用户一周的签到详情成功", c)
}

//通过userid查询部门id
func GetDeptByUserId(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		zap.L().Error("token获取userid失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	user := dingding2.GetDeptByUserId(UserId)
	response.OkWithDetailed(user.DeptList, "该用户的部门信息列表", c)
}
