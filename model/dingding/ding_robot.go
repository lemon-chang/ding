package dingding

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"ding/global"
	"ding/initialize/viper"
	"ding/utils"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrorSpecInvalid = errors.New("定时规则可能不合法")
)

type DingRobot struct {
	RobotId    string         `gorm:"primaryKey;foreignKey:RobotId" json:"robot_id"` //机器人的token
	Deleted    gorm.DeletedAt `json:"deleted"`                                       //软删除字段
	DingUserID string         `json:"ding_user_id"`                                  // 机器人所属用户id
	UserName   string         `json:"user_name"`                                     //机器人所属用户名
	Tasks      []Task         `gorm:"foreignKey:RobotId;references:RobotId"`         //机器人拥有多个任务
	Name       string         `json:"name"`                                          //机器人的名称
	DingToken  `json:"ding_token" gorm:"-"`
	IsShared   int `json:"is_shared"`
}

func (r *DingRobot) GetSharedRobot(p *ParamGetRobotList) (Robots []DingRobot, count int64, err error) {
	limit := p.PageSize
	offset := p.PageSize * (p.Page - 1)
	db := global.GLOAB_DB
	err = db.Model(&DingRobot{}).Where("is_shared = ?", 1).Count(&count).Error
	if err != nil {
		return
	}
	if p.Name != "" {
		db = db.Where("name = ?", p.Name)
	}
	err = global.GLOAB_DB.Limit(limit).Offset(offset).Where("is_shared = ?", 1).Find(&Robots).Error
	return

}
func (r *DingRobot) PingRobot() (err error) {
	robot, err := r.GetRobotByRobotId()
	if err != nil {
		zap.L().Error("通过robot_id获取robot失败", zap.Error(err))
		return
	}
	if robot.RobotId == "" {
		zap.L().Error("测试机器人发送消息失败，机器人id或者secret为空")
		return
	}
	p := &ParamCronTask{}
	p.MsgText.Msgtype = "text"
	p.MsgText.Text.Content = "测试"
	err = robot.SendMessage(p)
	if err != nil {
		zap.L().Error("测试机器人发送消息失败", zap.Error(err))
		return
	}
	return
}

type ResponseSendMessage struct {
	DingResponseCommon
}

func (r *DingRobot) AddDingRobot() (err error) {
	err = global.GLOAB_DB.Create(r).Error
	return
}
func (r *DingRobot) RemoveRobot() (err error) {
	err = global.GLOAB_DB.Delete(r).Error
	return
}
func (r *DingRobot) RemoveRobots(Robots []DingRobot) (err error) {
	err = global.GLOAB_DB.Delete(Robots).Error
	return
}
func (r *DingRobot) CreateOrUpdateRobot() (err error) {
	err = global.GLOAB_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "robot_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "is_shared"}),
	}).Create(r).Error
	if err != nil {
		return
	}
	var task *Task
	err = global.GLOAB_DB.Model(&task).Where("robot_id", r.RobotId).Update("robot_name", r.Name).Error
	if err != nil {
		return
	}
	return
}
func (r *DingRobot) GetRobotByRobotId() (robot *DingRobot, err error) {
	err = global.GLOAB_DB.Where("robot_id = ?", r.RobotId).First(&robot).Error
	return
}

type MySendParam struct {
	MsgParam  string   `json:"msgParam"`
	MsgKey    string   `json:"msgKey"`
	RobotCode string   `json:"robotCode"`
	UserIds   []string `json:"userIds"`
}

// 钉钉机器人单聊
func (r *DingRobot) ChatSendMessage(p *ParamChat) error {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	var b MySendParam
	b.RobotCode = viper.Conf.MiniProgramConfig.RobotCode
	if p.MsgKey == "sampleText" {
		b.MsgKey = p.MsgKey
		b.RobotCode = viper.Conf.MiniProgramConfig.RobotCode
		b.UserIds = p.UserIds
		b.MsgParam = fmt.Sprintf("{       \"content\": \"%s\"   }", p.MsgParam)

	} else if strings.Contains(p.MsgKey, "sampleActionCard") {
		b.MsgKey = p.MsgKey
		b.RobotCode = viper.Conf.MiniProgramConfig.RobotCode
		b.UserIds = p.UserIds
		b.MsgParam = p.MsgParam
	}

	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return nil
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return nil
	}
	token, err := r.DingToken.GetAccessToken()
	if err != nil {
		return err
	}
	request.Header.Set("x-acs-dingtalk-access-token", token)
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return nil
	}
	h := struct {
		Code                      string   `json:"code"`
		Message                   string   `json:"message"`
		ProcessQueryKey           string   `json:"processQueryKey"`
		InvalidStaffIdList        []string `json:"invalidStaffIdList"`
		FlowControlledStaffIdList []string `json:"flowControlledStaffIdList"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &h)
	if err != nil {
		return nil
	}
	if h.Code != "" {
		return errors.New(h.Message)
	}
	// 此处举行具体的逻辑判断，然后返回即可

	return nil
}

type MySendGroupParam struct {
	MsgParam           string `json:"msgParam"`
	MsgKey             string `json:"msgKey"`
	RobotCode          string `json:"robotCode"`
	OpenConversationId string `json:"openConversationId"`
	CoolAppCode        string `json:"coolAppCode"`
}

func (r *DingRobot) ChatSendGroupMessage(p *ParamChat) (map[string]interface{}, error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	var res map[string]interface{}
	URL := "https://api.dingtalk.com/v1.0/robot/groupMessages/send"
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	var b MySendGroupParam
	b.RobotCode = viper.Conf.MiniProgramConfig.RobotCode
	b.CoolAppCode = "COOLAPP-1-102118DC0ABA212C89C7000H"
	//b.OpenConversationId = "cidNOZESlAdvOGV/s3CVZxdlQ=="
	b.OpenConversationId = p.OpenConversationId
	if p.MsgKey == "sampleText" {
		b.MsgKey = p.MsgKey
		b.RobotCode = viper.Conf.MiniProgramConfig.RobotCode
		b.MsgParam = fmt.Sprintf("{       \"content\": \"%s\"   }", p.MsgParam)
	} else if strings.Contains(p.MsgKey, "sampleActionCard") {
		b.MsgKey = p.MsgKey
		b.RobotCode = viper.Conf.MiniProgramConfig.RobotCode
		b.MsgParam = p.MsgParam
	}

	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return res, nil
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return res, nil
	}
	token, err := r.DingToken.GetAccessToken()
	if err != nil {
		return res, err
	}
	request.Header.Set("x-acs-dingtalk-access-token", token)
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request)
	if err != nil {
		return res, nil
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return res, nil
	}

	//h := struct {
	//	Code                      string   `json:"code"`
	//	Message                   string   `json:"message"`
	//	ProcessQueryKey           string   `json:"processQueryKey"`
	//	InvalidStaffIdList        []string `json:"invalidStaffIdList"`
	//	FlowControlledStaffIdList []string `json:"flowControlledStaffIdList"`
	//}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &res)
	if err != nil {
		return res, nil
	}
	return res, nil
}
func (r *DingRobot) CronSend(c *gin.Context, p *ParamCronTask) (err error, task Task) {
	spec, detailTimeForUser := "", ""
	if (p.RepeatTime) == "立即发送" {
		err = r.SendMessage(p)
		return
	} else {
		spec, detailTimeForUser, err = HandleSpec(p)
		p.Spec = spec
		if err != nil {
			return
		}
	}
	userId, _ := c.Get(global.CtxUserIDKey)
	UserName, _ := c.Get(global.CtxUserNameKey)
	//把定时任务添加到数据库中
	task = Task{
		//TaskID:            int(TaskID),
		TaskName:          p.TaskName,
		UserId:            userId.(string),
		UserName:          UserName.(string),
		RobotId:           r.RobotId,
		DetailTimeForUser: detailTimeForUser, //给用户看的
		Spec:              spec,              //cron后端定时规则
		FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
		FrontDetailTime:   p.DetailTime,
		MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
		IsSuspend:         false,
		NextTime:          time.Now(),
	}
	err = (&task).Insert(p)
	if err != nil {
		return
	}
	return
}

// SendMessage Function to send message
//
//goland:noinspection GoUnhandledErrorResult
func (t *DingRobot) SendMessage(p *ParamCronTask) error {
	b := []byte{}
	//我们需要在文本，链接，markdown三种其中的一个
	if p.MsgText.Msgtype == "text" {
		msg := map[string]interface{}{}
		atMobileStringArr := make([]string, len(p.MsgText.At.AtMobiles))
		for i, atMobile := range p.MsgText.At.AtMobiles {
			atMobileStringArr[i] = atMobile.AtMobile
		}
		atUserIdStringArr := make([]string, len(p.MsgText.At.AtUserIds))
		for i, AtuserId := range p.MsgText.At.AtUserIds {
			atUserIdStringArr[i] = AtuserId.AtUserId
		}
		msg = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": p.MsgText.Text.Content,
			},
			"at": map[string]interface{}{
				"atMobiles": atMobileStringArr, //字符串切片类型
				"atUserIds": atUserIdStringArr,
				"isAtAll":   p.MsgText.At.IsAtAll,
			},
		}
		b, _ = json.Marshal(msg)
	}

	var resp *http.Response
	var err error
	resp, err = http.Post(t.getURLV2(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	date, err := ioutil.ReadAll(resp.Body)
	r := ResponseSendMessage{}
	err = json.Unmarshal(date, &r)
	if err != nil {
		return err
	}
	if r.Errcode != 0 {
		fmt.Println(r.Errmsg)
		return errors.New(r.Errmsg)
	}

	return nil
}

func (t *DingRobot) hmacSha256(stringToSign string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (t *DingRobot) getURLV2() string {
	url := "https://oapi.dingtalk.com/robot/send?access_token=" + t.RobotId //拼接token路径
	return url
}

func (*DingRobot) SendSessionWebHook(p *ParamReveiver) (err error) {
	var msg map[string]interface{}
	//如果@机器人的消息包含考勤，且包含三期或者四期，再加上时间限制
	robot := &DingRobot{}
	if strings.Contains(p.Text.Content, "打字邀请码") {
		code, _, err := robot.GetInviteCode()
		if err != nil {
			zap.L().Error("申请新的TypingInviationCode失败", zap.Error(err))
			return err
		}
		msg = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": utils.TypingInviationSucc + ": " + code,
			},
		}
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var resp *http.Response

	resp, err = http.Post(p.SessionWebhook, "application/json", bytes.NewBuffer(b))

	defer resp.Body.Close()
	date, err := ioutil.ReadAll(resp.Body)
	fmt.Println(date)
	if err != nil {
		return err
	}
	return nil
}

func TypingInviation() (TypingInvitationCode string, expire time.Duration, err error) {
	zap.L().Info("进入到了chromedp，开始申请")
	timeCtx, cancel := context.WithTimeout(GetChromeCtx(false), 5*time.Minute)
	defer cancel()
	//opts := append(
	//	chromedp.DefaultExecAllocatorOptions[:],
	//	chromedp.NoDefaultBrowserCheck,                        //不检查默认浏览器
	//	chromedp.Flag("headless", false),                      // 禁用chrome headless（禁用无窗口模式，那就是开启窗口模式）
	//	chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
	//	chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
	//	chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
	//	chromedp.Flag("disable-extensions", true),             //开启插件支持
	//	chromedp.Flag("disable-default-apps", true),
	//	chromedp.NoFirstRun, //设置网站不是首次运行
	//	chromedp.WindowSize(1921, 1024),
	//	chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
	//)
	//
	//allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	//defer cancel()
	//
	////创建上下文实例
	//timeCtx, cancel := chromedp.NewContext(
	//	allocCtx,
	//	chromedp.WithLogf(log.Printf),
	//)
	//defer cancel()
	// 创建超时上下文
	var html string
	//ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)

	err = chromedp.Run(timeCtx,
		chromedp.Navigate("https://dazi.kukuw.com/"),
		//点击“我的打字“按钮
		chromedp.Click(`document.getElementById("globallink").getElementsByTagName("a")[5]`, chromedp.ByJSPath),
		// 锁定用户名框并填写内容
		chromedp.WaitVisible(`document.querySelector("#name")`, chromedp.ByJSPath),
		chromedp.SetValue(`document.querySelector("#name")`, "闫佳鹏", chromedp.ByJSPath),
		//锁定密码框并填写内容
		chromedp.WaitVisible(`document.querySelector("#pass")`, chromedp.ByJSPath),
		chromedp.SetValue(`document.querySelector("#pass")`, "123456", chromedp.ByJSPath),
		//点击登录按钮
		chromedp.WaitVisible(`document.querySelector(".button").firstElementChild`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector(".button").firstElementChild`, chromedp.ByJSPath),
		//点击发布竞赛
		chromedp.WaitVisible(`document.querySelector("a.groupnew")`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector("a.groupnew")`, chromedp.ByJSPath),
		chromedp.Sleep(time.Second),
		////点击所要打字的文章
		chromedp.WaitVisible(`document.querySelector("a#select_b.select_b")`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector("a#select_b.select_b")`, chromedp.ByJSPath),
		chromedp.WaitVisible(`document.querySelector("a.sys.on")`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector("a.sys.on")`, chromedp.ByJSPath),
		//设置比赛时间2分钟
		chromedp.Evaluate(`document.querySelector("#set_time").value=10`, nil),
		//选择有效期
		chromedp.Evaluate("document.querySelector(\"select#youxiaoqi\").value = document.querySelector(\"#youxiaoqi > option:nth-child(5)\").value", nil),
		//设置成为不公开
		chromedp.Click(`document.querySelectorAll("input#gongkai")[1]`, chromedp.ByJSPath),
		//点击发布按钮
		chromedp.Click(`document.querySelectorAll(".artnew table tr td input")[7]`, chromedp.ByJSPath),
		chromedp.WaitVisible(`document.querySelectorAll("#my_main .art_table td")[9].childNodes[0]`, chromedp.ByJSPath),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("打字吗出现了")
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("爬取前:", TypingInvitationCode)
			a := chromedp.OuterHTML(`document.querySelector("body")`, &html, chromedp.ByJSPath)
			err := a.Do(ctx)
			if err != nil {
				zap.L().Error("chromedp获取页面全部数据失败", zap.Error(err))
				return err
			}
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				zap.L().Error("chromedp获取页面全部数据后，转化成dom失败", zap.Error(err))
				return err
			}
			//dom.Find(`#my_main > .art_table > tbody > tr:nth-child(2) > td:nth-child(4) > span`).Each(func(i int, selection *goquery.Selection) {
			//	TypingInvitationCode = TypingInvitationCode + selection.Text()
			//	fmt.Println("爬取后:",TypingInvitationCode)
			//	selection.Next()
			//})
			TypingInvitationCode = dom.Find(`#my_main > .art_table > tbody > tr:nth-child(2) > td:nth-child(4) > span`).First().Text()
			if TypingInvitationCode == "" {
				zap.L().Error("爬取打字邀请码失败")
				return err
			}
			_, err = global.GLOBAL_REDIS.Set(context.Background(), utils.ConstTypingInvitationCode, TypingInvitationCode, time.Second*60*60*5).Result() //5小时过期时间
			if err != nil {
				zap.L().Error("爬取打字邀请码后存入redis失败", zap.Error(err))
			}
			return err
		}),
	)

	if err != nil {
		zap.L().Error("chromedp.Run有误", zap.Error(err))
		return "", time.Second * 0, err
	} else {
		zap.L().Info(fmt.Sprintf("chromedp.Run无误，成功获取打字邀请码:%v", TypingInvitationCode), zap.Error(err))
		return TypingInvitationCode, time.Second * 60 * 60 * 5, err
	}

}
func (d *DingRobot) GetInviteCode() (code string, expire time.Duration, err error) {
	//如果@机器人的消息包含考勤，且包含三期或者四期，再加上时间限制
	//去redis中取一下打字邀请码
	var TypingInviationCode string
	var expire1 int64
	fmt.Println(expire1)
	expire, err = global.GLOBAL_REDIS.TTL(context.Background(), utils.ConstTypingInvitationCode).Result()
	if err != nil {
		zap.L().Error("判断token剩余生存时间失败", zap.Error(err))
	}
	//如果redis里面没有的话
	if expire == -2 {
		zap.L().Error("redis中无打字码，去申请", zap.Error(err))
		//申请新的TypingInviationCode并已经存入redis
		TypingInviationCode, expire, err = TypingInviation()
		if err != nil || TypingInviationCode == "" {
			zap.L().Error("申请新的TypingInviationCode失败", zap.Error(err))
			return TypingInviationCode, time.Second * 0, err
		}

	} else {
		//从redis从取到邀请码
		TypingInviationCode = global.GLOBAL_REDIS.Get(context.Background(), utils.ConstTypingInvitationCode).Val()
		if len(TypingInviationCode) != 5 {
			zap.L().Error("申请新的TypingInviationCode失败", zap.Error(err))
			return TypingInviationCode, expire, errors.New("申请新的TypingInviationCode失败")
		}
	}
	return TypingInviationCode, expire, nil

}
func HandleSpec(p *ParamCronTask) (spec, detailTimeForUser string, err error) {
	spec = ""
	detailTimeForUser = ""
	n := len(p.DetailTime)
	if p.Spec != "" {
		spec = p.Spec
		detailTimeForUser = p.Spec
	}
	if p.RepeatTime == "仅发送一次" {
		second := p.DetailTime[n-2:]
		minute := p.DetailTime[n-5 : n-3]
		hour := p.DetailTime[n-8 : n-6]
		//year := p.DetailTime[:4]
		month := p.DetailTime[5:7]
		day := p.DetailTime[8:10]
		week := "?" //问号代表放弃周
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
		detailTimeForUser = "仅在" + p.DetailTime + "发送一次"
	}
	if string([]rune(p.RepeatTime)[0:3]) == "周重复" {
		M := map[string]string{"0": "周日", "1": "周一", "2": "周二", "3": "周三", "4": "周四", "5": "周五", "6": "周六"}
		detailTimeForUser = "周重复 ："
		weeks := strings.Split(p.RepeatTime, "/")[1:]
		week := ""
		for i := 0; i < len(weeks); i++ {
			detailTimeForUser += M[weeks[i]]
			week += weeks[i] + ","
		}
		week = week[0 : len(week)-1]
		HMS := strings.Split(p.DetailTime, ":")
		second := HMS[2]
		minute := HMS[1]
		hour := HMS[0]
		month := "*" //每个月的每个星期都发送
		day := "?"   //选了星期就要放弃具体的某一天
		detailTimeForUser += hour + "：" + minute + "：" + second
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
	}

	if string([]rune(p.RepeatTime)[0:3]) == "月重复" {
		var daymap map[int]string
		daymap = make(map[int]string)
		for i := 1; i <= 31; i++ {
			daymap[i] += strconv.Itoa(i) + "号"
		}
		//字符串数组
		days := strings.Split(p.RepeatTime, "/")[1:]
		detailTimeForUser = "月重复 ："
		day := ""
		for i := 0; i < len(days); i++ {
			atoi, _ := strconv.Atoi(days[i])
			detailTimeForUser += daymap[atoi]
			day += days[i] + ","
		}
		day = day[0 : len(day)-1]
		HMS := strings.Split(p.DetailTime, ":")
		second := HMS[2]
		minute := HMS[1]
		hour := HMS[0]
		month := "*" //每个月的每个星期都发送
		week := "?"
		detailTimeForUser += hour + ":" + minute + ":" + second
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
	}

	if spec == "" || detailTimeForUser == "" {
		return spec, detailTimeForUser, errors.New("cron定时规则转化错误")
	}
	return spec, detailTimeForUser, nil
}

func HandleSpec1(p *UpdateTask) (spec, detailTimeForUser string, err error) {
	spec = ""
	detailTimeForUser = ""
	n := len(p.DetailTime)
	if p.Spec != "" {
		spec = p.Spec
		detailTimeForUser = p.Spec
	} else if p.RepeatTime == "仅发送一次" {
		second := p.DetailTime[n-2:]
		minute := p.DetailTime[n-5 : n-3]
		hour := p.DetailTime[n-8 : n-6]
		//year := p.DetailTime[:4]
		month := p.DetailTime[5:7]
		day := p.DetailTime[8:10]
		week := "?" //问号代表放弃周
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
		detailTimeForUser = "仅在" + p.DetailTime + "发送一次"
	} else if string([]rune(p.RepeatTime)[0:3]) == "周重复" {
		M := map[string]string{"0": "周日", "1": "周一", "2": "周二", "3": "周三", "4": "周四", "5": "周五", "6": "周六"}
		detailTimeForUser = "周重复 ："
		weeks := strings.Split(p.RepeatTime, "/")[1:]
		week := ""
		for i := 0; i < len(weeks); i++ {
			detailTimeForUser += M[weeks[i]]
			week += weeks[i] + ","
		}
		week = week[0 : len(week)-1]
		HMS := strings.Split(p.DetailTime, ":")
		second := HMS[2]
		minute := HMS[1]
		hour := HMS[0]
		month := "*" //每个月的每个星期都发送
		day := "?"   //选了星期就要放弃具体的某一天
		detailTimeForUser += hour + "：" + minute + "：" + second
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
	} else if string([]rune(p.RepeatTime)[0:3]) == "月重复" {
		var daymap map[int]string
		daymap = make(map[int]string)
		for i := 1; i <= 31; i++ {
			daymap[i] += strconv.Itoa(i) + "号"
		}
		//字符串数组
		days := strings.Split(p.RepeatTime, "/")[1:]
		detailTimeForUser = "月重复 ："
		day := ""
		for i := 0; i < len(days); i++ {
			atoi, _ := strconv.Atoi(days[i])
			detailTimeForUser += daymap[atoi]
			day += days[i] + ","
		}
		day = day[0 : len(day)-1]
		HMS := strings.Split(p.DetailTime, ":")
		second := HMS[2]
		minute := HMS[1]
		hour := HMS[0]
		month := "*" //每个月的每个星期都发送
		week := "?"
		detailTimeForUser += hour + ":" + minute + ":" + second
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
	}

	if spec == "" || detailTimeForUser == "" {
		return spec, detailTimeForUser, errors.New("cron定时规则转化错误")
	}
	return spec, detailTimeForUser, nil
}

// 获取机器人所在的群聊的userIdList ，前提是获取到OpenConversationId，获取到OpenConverstaionId的前提是获取到二维码
func (r *DingRobot) GetGroupUserIds() (userIds []string, _err error) {
	//所需参数access_token, OpenConversationId string
	olduserIds := []*string{}
	client, _err := createClient()
	if _err != nil {
		return
	}

	batchQueryGroupMemberHeaders := &dingtalkim_1_0.BatchQueryGroupMemberHeaders{}
	batchQueryGroupMemberHeaders.XAcsDingtalkAccessToken = tea.String(r.DingToken.Token)
	batchQueryGroupMemberRequest := &dingtalkim_1_0.BatchQueryGroupMemberRequest{

		CoolAppCode: tea.String("COOLAPP-1-102118DC0ABA212C89C7000H"),
		MaxResults:  tea.Int64(300),
		NextToken:   tea.String("XXXXX"),
	}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.BatchQueryGroupMemberWithOptions(batchQueryGroupMemberRequest, batchQueryGroupMemberHeaders, &util.RuntimeOptions{})
		if _err != nil {
			return _err
		}
		olduserIds = result.Body.MemberUserIds
		return
	}()

	if tryErr != nil {
		var err = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			err = _t
		} else {
			err.Message = tea.String(tryErr.Error())
		}
		if !tea.BoolValue(util.Empty(err.Code)) && !tea.BoolValue(util.Empty(err.Message)) {
			// err 中含有 code 和 message 属性，可帮助开发定位问题
		}

	}
	userIds = make([]string, len(olduserIds))
	for i, id := range olduserIds {
		userIds[i] = *id
	}
	return
}

func createClient() (_result *dingtalkim_1_0.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result, _err = dingtalkim_1_0.NewClient(config)
	return _result, _err
}

func GetImage(c *gin.Context) { //显示图片的方法
	imageName := c.Query("imageName")     //截取get请求参数，也就是图片的路径，可是使用绝对路径，也可使用相对路径
	file, _ := ioutil.ReadFile(imageName) //把要显示的图片读取到变量中
	c.Writer.WriteString(string(file))    //关键一步，写给前端
}

// 获取所有的公共机器人
func GetAllPublicRobot() (robot []DingRobot, err error) {
	//IsShare值为1为公共机器人
	err = global.GLOAB_DB.Where("is_shared=?", 1).Find(&robot).Error
	if err != nil {
		zap.L().Error("服务繁忙", zap.Error(err))
		return nil, err
	}
	return robot, err
}

func AlterResultByRobot(p *ParamAlterResultByRobot) (err error) {
	err = global.GLOAB_DB.Table("ding_depts").Where("dept_id", p.DeptId).Update("robot_token", p.Token).Error
	return
}

type result struct {
	ChatId string `json:"chatId"`
	Title  string `json:"title"`
}
type data struct {
	Result result `json:"result"`
}

func (t *DingRobot) RobotSendInviteCode(resp *RobotAtResp) error {
	code, expire, err := t.GetInviteCode()
	if err != nil {
		zap.L().Error("获取邀请码失败", zap.Error(err))
	}
	content := fmt.Sprintf(
		"欢迎加入闫佳鹏的打字邀请比赛\n网站: https://dazi.kukuw.com/\n邀请码: %v\n比赛剩余时间: %v",
		code, expire)
	if err != nil {
		content = "获取失败！"
	}
	param := &ParamChat{
		MsgKey:    "sampleText",
		MsgParam:  content,
		RobotCode: viper.Conf.MiniProgramConfig.RobotCode,
		UserIds:   []string{resp.SenderStaffId},
	}
	err = t.ChatSendMessage(param)
	if err != nil {
		zap.L().Error("单聊中发送打字邀请码错误" + err.Error())
		return err
	}
	return nil
}

func (t *DingRobot) RobotSendGroupInviteCode(resp *RobotAtResp) (res map[string]interface{}, err error) {
	code, expire, err := t.GetInviteCode()
	if err != nil {
		zap.L().Error("获取邀请码失败", zap.Error(err))
	}
	content := fmt.Sprintf(
		"欢迎加入闫佳鹏的打字邀请比赛\n网站: https://dazi.kukuw.com/\n邀请码: %v\n比赛剩余时间: %v",
		code, expire)
	if err != nil {
		content = "获取失败！"
	}
	param := &ParamChat{
		MsgKey:             "sampleText",
		MsgParam:           content,
		RobotCode:          viper.Conf.MiniProgramConfig.RobotCode,
		UserIds:            []string{resp.SenderStaffId},
		OpenConversationId: resp.ConversationId,
	}
	res, err = t.ChatSendGroupMessage(param)
	if err != nil {
		zap.L().Error("单聊中发送打字邀请码错误" + err.Error())
	}
	return res, nil
}
func (t *DingRobot) RobotSendGroupWater(resp *RobotAtResp) (res map[string]interface{}, err error) {
	param := &ParamChat{
		MsgKey:             "sampleText",
		MsgParam:           "送水师傅电话: 15236463964",
		RobotCode:          viper.Conf.MiniProgramConfig.RobotCode,
		UserIds:            []string{resp.SenderStaffId},
		OpenConversationId: resp.ConversationId,
	}
	res, err = t.ChatSendGroupMessage(param)
	if err != nil {
		zap.L().Error("发送送水师傅电话失败" + err.Error())
	}
	return res, nil
}
func (t *DingRobot) RobotSendGroupCard(resp *RobotAtResp) (res map[string]interface{}, err error) {
	param := &ParamChat{
		MsgKey: "sampleActionCard2",
		MsgParam: "{\n" +
			"        \"title\": \"帮助\",\n" +
			"        \"text\": \"请问你是否在查找以下功能\",\n" +
			"        \"actionTitle1\": \"送水电话号码\",\n" +
			fmt.Sprintf("'actionURL1':'dtmd://dingtalkclient/sendMessage?content=%s',\n", url.QueryEscape("送水电话号码")) +
			"        \"actionTitle2\": \"打字邀请码\",\n" +
			fmt.Sprintf("'actionURL2':'dtmd://dingtalkclient/sendMessage?content=%s',\n", url.QueryEscape("打字邀请码")) +
			"    }",
		RobotCode:          viper.Conf.MiniProgramConfig.RobotCode,
		UserIds:            []string{resp.SenderStaffId},
		OpenConversationId: resp.ConversationId,
	}
	res, err = t.ChatSendGroupMessage(param)
	if err != nil {
		zap.L().Error("发送chatSendMessage错误" + err.Error())
	}
	return res, nil
}

type Result struct {
	Name     string `json:"name"`
	DataName string `json:"data_name"`
	DataLink string `json:"data_link"`
}

// 机器人问答发送卡片给个人https://open.dingtalk.com/document/isvapp/the-internal-outgoing-of-the-enterprise-realizes-the-interaction-in
func (t *DingRobot) RobotSendCardToPerson(resp *RobotAtResp, dataByStr []Result) (err error) {
	cardLen := len(dataByStr)
	if cardLen <= 5 {
	} else {
		cardLen = 5
	}
	action := ""
	var param *ParamChat
	if cardLen == 1 {
		//for i, data := range dataByStr {
		//	action += "        \"actionTitle" + strconv.Itoa(i+1) + "\": \"" + data.DataName + "\",\n" +
		//		fmt.Sprintf("'actionURL%d':'dtmd://dingtalkclient/sendMessage?content=%s',\n", i+1, url.QueryEscape(data.DataName))
		//}

		//action = fmt.Sprintf("\"singleTitle\": \"%s\",\n     \"singleURL\": \"%s\"", dataByStr[0].DataName, dataByStr[0].DataLink)
		for _, data := range dataByStr {
			action += "        \"singleTitle" + "\": \"" + data.DataName + "\",\n" +
				fmt.Sprintf("'singleURL':'dtmd://dingtalkclient/sendMessage?content=%s',\n", url.QueryEscape(data.DataName))
		}
		param = &ParamChat{
			MsgKey: "sampleActionCard",
			MsgParam: "{\n" +
				"        \"title\": \"资料\",\n" +
				"        \"text\": \"请问你是否在查找以下资料\",\n" +
				action +
				"    }",
			RobotCode: viper.Conf.MiniProgramConfig.RobotCode,
			UserIds:   []string{resp.SenderStaffId},
		}

	} else {
		for i, data := range dataByStr {
			action += "        \"actionTitle" + strconv.Itoa(i+1) + "\": \"" + data.DataName + "\",\n" +
				fmt.Sprintf("'actionURL%d':'dtmd://dingtalkclient/sendMessage?content=%s',\n", i+1, url.QueryEscape(data.DataName))
		}
		param = &ParamChat{
			MsgKey: "sampleActionCard" + strconv.Itoa(cardLen),
			MsgParam: "{\n" +
				"        \"title\": \"资料\",\n" +
				"        \"text\": \"请问你是否在查找以下资料\",\n" +
				action +
				"    }",
			RobotCode: viper.Conf.MiniProgramConfig.RobotCode,
			UserIds:   []string{resp.SenderStaffId},
		}
	}

	fmt.Println(action)

	err = t.ChatSendMessage(param)
	if err != nil {
		zap.L().Error("发送chatSendCardToPerson错误" + err.Error())
	}
	return
}

// 机器人问答发送信息给个人
func (t *DingRobot) RobotSendMessageToPerson(resp *RobotAtResp, dataByStr []Result) (err error) {
	msg := ""
	if len(dataByStr) == 0 {
		msg = "您所查询的资源里没有此类资源"
	} else if len(dataByStr) == 1 {
		msg = dataByStr[0].DataLink
	} else {
		msg = "查询结果如下：\n"
		for _, data := range dataByStr {
			msg += "上传资料人员：" + data.Name + "\n" + "资源名称：" + data.DataName + "\n" + "资源内容：" + data.DataLink + "\n"
		}
	}
	param := &ParamChat{
		MsgKey:    "sampleText",
		MsgParam:  msg,
		RobotCode: viper.Conf.MiniProgramConfig.RobotCode,
		UserIds:   []string{resp.SenderStaffId},
	}
	err = t.ChatSendMessage(param)
	if err != nil {
		zap.L().Error("发送chatSendMessageToPerson错误" + err.Error())
	}
	return
}
func (u *DingRobot) GetOpenConversationId(access_token, chatId string) (openConversationId string, _err error) {
	client, _err := createClient()
	if _err != nil {
		return
	}

	chatIdToOpenConversationIdHeaders := &dingtalkim_1_0.ChatIdToOpenConversationIdHeaders{}
	chatIdToOpenConversationIdHeaders.XAcsDingtalkAccessToken = tea.String(access_token)
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.ChatIdToOpenConversationIdWithOptions(tea.String(chatId), chatIdToOpenConversationIdHeaders, &util.RuntimeOptions{})
		if _err != nil {
			return _err
		}
		openConversationId = *(result.Body.OpenConversationId)
		return nil
	}()

	if tryErr != nil {
		var err = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			err = _t
		} else {
			err.Message = tea.String(tryErr.Error())
		}
		if !tea.BoolValue(util.Empty(err.Code)) && !tea.BoolValue(util.Empty(err.Message)) {
			// err 中含有 code 和 message 属性，可帮助开发定位问题
		}

	}
	return
}
