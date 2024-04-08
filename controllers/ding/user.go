package ding

import (
	"crypto/tls"
	"ding/model/common/localTime"
	"ding/model/common/request"
	response2 "ding/model/common/response"
	"runtime"
	"strconv"

	"ding/global"
	"ding/initialize/jwt"
	"ding/model/dingding"
	"ding/model/params"
	"ding/response"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}

// SelectAllUsers 查询所有用户
func SelectAllUsers(c *gin.Context) {
	var pageInfo request.PageInfo
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage("分页参数有误", c)
		return
	}
	name := c.Query("name")
	mobile := c.Query("mobile")
	deptId, _ := strconv.Atoi(c.Query("deptId"))
	authorityId, _ := strconv.Atoi(c.Query("authorityId"))
	var DingUser dingding.DingUser
	list, total, err := DingUser.FindDingUsersInfo(name, mobile, deptId, authorityId, pageInfo, c)
	if err != nil {
		response.FailWithMessage("查询用户失败", c)
		return
	}
	response.OkWithDetailed(response2.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "查询所有用户成功", c)
}

func GetUserInfoDetailByToken(c *gin.Context) {
	user_id, _ := c.Get(global.CtxUserIDKey)
	DingUser := dingding.DingUser{UserId: user_id.(string)}
	err := DingUser.GetUserInfoDetailByUserId()
	if err != nil {
		response.FailWithMessage("查询用户失败", c)
		return
	}
	response.OkWithDetailed(DingUser, "查询所有用户成功", c)
}

// LoginHandler 处理登录请求的函数
func LoginHandler(c *gin.Context) {
	//1.获取请求参数及参数校验
	var p params.ParamLogin
	if err := c.ShouldBindJSON(&p); err != nil { //这个地方只能判断是不是json格式的数据
		zap.L().Error("Login with invalid param", zap.Error(err))
		response.FailWithMessage("参数有误", c)
		return
	}
	//2.业务逻辑处理
	//3.返回响应
	user, err := (&dingding.DingUser{Mobile: p.Mobile, Password: p.Password}).Login()
	if err != nil {
		response.FailWithMessage("用户登录失败", c)
		return
	}
	// 生成JWT
	token, err := jwt.GenToken(c, user)
	if err != nil {
		response.FailWithMessage("用户登录失败", c)
		zap.L().Error("JWT生成错误", zap.Error(err))
		return
	}
	user.AuthToken = token
	if err != nil {
		response.FailWithMessage("用户登录失败", c)
	} else {
		response.OkWithDetailed(user, "用户登录成功", c)
	}
}
func LoginByDingDing(c *gin.Context) {
	var p params.ParamLoginByDingDing
	if err := c.ShouldBindJSON(&p); err != nil { //这个地方只能判断是不是json格式的数据
		zap.L().Error("Login with invalid param", zap.Error(err))
		response.FailWithMessage("参数有误", c)
		return
	}
	err := global.GLOAB_VALIDATOR.Struct(p)
	if err != nil {
		zap.L().Error("validator with invalid param", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	user, err := (&dingding.Code{AuthCode: p.AuthCode}).SweepLogin(c)
	if err != nil {
		response.FailWithMessage("扫码登陆有误", c)
		zap.L().Error("扫码登陆有误", zap.Error(err))
		return
	}
	// 生成JWT
	token, err := jwt.GenToken(c, &user)
	if err != nil {
		response.FailWithMessage("用户登录失败", c)
		zap.L().Error("JWT生成错误", zap.Error(err))
		return
	}
	user.AuthToken = token
	if err != nil {
		response.FailWithMessage("登陆失败", c)
	} else {
		response.OkWithDetailed(user, "用户登录成功", c)
	}

}
func LoginHandlerByToken(c *gin.Context) {
	//1.获取请求参数及参数校验
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		response.FailWithMessage("需要登录", c)
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
	var buf []byte
	var ChatID, title string
	var err error
	if runtime.GOOS == "linux" {
		buf, ChatID, title, err = (&dingding.DingUser{}).GetQRCodeInLinux(c)
	} else if runtime.GOOS == "windows" {
		buf, ChatID, title, err = (&dingding.DingUser{}).GetQRCodeInWindows(c)
	}
	if err != nil {
		zap.L().Error("截取二维码和获取群聊基本错误", zap.Error(err))
		response.FailWithMessage("截取二维码和获取群聊基本错误", c)
	}
	token, err := (&dingding.DingToken{}).GetAccessToken()
	if err != nil {
		zap.L().Error("获取token失败", zap.Error(err))
		return
	}
	userIds, err := (&dingding.DingRobot{DingToken: dingding.DingToken{Token: token}}).GetGroupUserIds()
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

// 获取所有任务列表,包括已暂停的任务
func GetTasks(c *gin.Context) {
	var p params.ParamGetTasks
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("Login with invalid param", zap.Error(err))
		response.FailWithMessage("参数有误", c)
		return
	}
	user_id := global.GetCurrentUserId(c)
	//err := global.GLOAB_DB.Model(&tasks).Where("user_id", user_id).Unscoped().Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").Find(&tasks).Error
	tasks, err := (&dingding.Task{}).GetTasks(user_id, &p)
	if err != nil {
		zap.L().Error("获取所有定时任务失败", zap.Error(err))
		response.FailWithMessage("服务繁忙", c)
		return
	}
	response.OkWithDetailed(tasks, "获取成功", c)
}
func GetAllActiveTask(c *gin.Context) {
	tasks, err := (&dingding.Task{}).GetAllActiveTask()
	if err != nil {
		response.FailWithMessage("获取定时任务失败", c)
		return
	}
	response.OkWithDetailed(tasks, "获取所有获取定时任务成功", c)
}
func MakeupSign(c *gin.Context) {
	var p params.ParamMakeupSign
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage("参数错误", c)
		zap.L().Error("参数错误", zap.Error(err))
		return
	}
	WeekSignNum, _, err := (&dingding.DingUser{UserId: p.Userid}).Sign(p.Semester, p.StartWeek, p.WeekDay, p.MNE)
	if err != nil {
		zap.L().Error("为用户补签失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(WeekSignNum, "补签成功", c)
}
func GetWeekConsecutiveSignNum(c *gin.Context) {
	var p params.ParamGetWeekConsecutiveSignNum
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage("参数错误", c)
		zap.L().Error("参数错误", zap.Error(err))
		return
	}
	curTime := &localTime.MySelfTime{}
	err := curTime.GetCurTime(nil)

	weekDay, MNE := 7, 3
	if curTime.Semester == p.Semester && curTime.StartWeek == p.StartWeek {
		// 说明用户要看当前周，赋值为调用本接口时刻的周和具体的上午下午晚上
		weekDay = curTime.Week
		MNE = curTime.Duration
	}
	consecutiveSignNum, err := (&dingding.DingUser{UserId: p.Userid}).GetConsecutiveSignNum(p.Semester, p.StartWeek, weekDay, MNE)
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
	WeekSignNum, err := (&dingding.DingUser{UserId: p.Userid}).GetWeekSignNum(p.Semester, p.StartWeek)
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
	WeekSignNum, err := (&dingding.DingUser{UserId: p.Userid}).GetWeekSignDetail(p.Semester, p.StartWeek)

	if err != nil {
		zap.L().Error("获取用户一周的签到详情失败", zap.Error(err))
		response.FailWithMessage("获取用户一周的签到详情失败", c)
		return
	}
	response.OkWithDetailed(WeekSignNum, "获取用户一周的签到详情成功", c)
}

// 通过userid查询部门id
func GetDeptByUserId(c *gin.Context) {
	UserId := global.GetCurrentUserId(c)
	user := dingding.GetDeptByUserId(UserId)
	response.OkWithDetailed(user.DeptList, "该用户的部门信息列表", c)
}
