package dingding

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"ding/global"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrorUserExist             = errors.New("用户已经存在")
	ErrorUserNotExist          = errors.New("用户不存在")
	ErrorInvalidPassword       = errors.New("用户密码错误")
	ErrorInvalidID             = errors.New("无效的ID")
	ErrorRobotExist            = errors.New("该机器人ID已经存在或者在您的账户机器人昵称重复")
	ErrorRobotNotExist         = errors.New("在当前用户下机器人不存在,无法删除")
	ErrorTeleOrPersonNameExist = errors.New("在该机器人中此电话号码或者姓名已经存在")
	ErrorNotHasRobot           = errors.New("该用户未拥有该机器人")
	ErrorNotHasTask            = errors.New("未拥有该任务")
	ErrorSpecInvalid           = errors.New("定时规则可能不合法")
)

type DingRobot struct {
	RobotId            string         `gorm:"primaryKey;foreignKey:RobotId" json:"robot_id"` //机器人的token
	Deleted            gorm.DeletedAt `json:"deleted"`                                       //软删除字段
	Type               string         `json:"type"`                                          //机器人类型，1为企业内部机器人，2为自定义webhook机器人
	TypeDetail         string         `json:"type_detail"`                                   //具体机器人类型
	ChatBotUserId      string         `json:"chat_bot_user_id"`                              //加密的机器人id，该字段无用
	Secret             string         `json:"secret"`                                        //如果是自定义成机器人， 则存在此字段
	DingUserID         string         `json:"ding_user_id"`                                  // 机器人所属用户id
	UserName           string         `json:"user_name"`                                     //机器人所属用户名
	DingUsers          []DingUser     `json:"ding_users" gorm:"many2many:user_robot"`        //机器人@多个人，一个人可以被多个机器人@
	ChatId             string         `json:"chat_id"`                                       //机器人所在的群聊chatId
	OpenConversationID string         `json:"open_conversation_id"`                          //机器人所在的群聊openConversationID
	Tasks              []Task         `gorm:"foreignKey:RobotId;references:RobotId"`         //机器人拥有多个任务
	Name               string         `json:"name"`                                          //机器人的名称
	DingToken          `json:"ding_token" gorm:"-"`
	IsShared           int `json:"is_shared"`
}

func (r *DingRobot) GetSharedRobot() (Robots []DingRobot, err error) {
	err = global.GLOAB_DB.Where("is_shared = ?", 1).Find(&Robots).Error
	return
}
func (r *DingRobot) InsertRobot() (err error) {
	err = global.GLOAB_DB.Create(r).Error
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

func (r *DingRobot) CreateOrUpdateRobot() (err error) {
	err = global.GLOAB_DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(r).Error
	return
}
func (r *DingRobot) GetRobotByRobotId() (robot *DingRobot, err error) {
	err = global.GLOAB_DB.Where("robot_id = ?", r.RobotId).First(&robot).Error
	return
}

//钉钉机器人单聊
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
	b := struct {
		MsgParam  string   `json:"msgParam"`
		MsgKey    string   `json:"msgKey"`
		RobotCode string   `json:"robotCode"`
		UserIds   []string `json:"userIds"`
	}{MsgParam: fmt.Sprintf("{       \"content\": \"%s\"   }", p.MsgParam),
		MsgKey:    p.MsgKey,
		RobotCode: r.RobotId,
		UserIds:   p.UserIds,
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
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil
	}
	if h.Code != "" {
		return errors.New(h.Message)
	}
	// 此处举行具体的逻辑判断，然后返回即可

	return nil
}

func (r *DingRobot) CronSend(c *gin.Context, p *ParamCronTask) (err error, task Task) {
	robotId := r.RobotId
	spec, detailTimeForUser, err := HandleSpec(p)
	if p.Spec != "" {
		spec = p.Spec
	}
	tid := "0"
	UserId := ""
	if c != nil {
		UserId, err = global.GetCurrentUserId(c)
	}

	if err != nil {
		UserId = ""
	}
	CurrentUser, err := (&DingUser{UserId: UserId}).GetUserByUserId()
	if err != nil {
		CurrentUser = DingUser{}
	}
	r, err = (&DingRobot{RobotId: r.RobotId}).GetRobotByRobotId()
	if err != nil {
		zap.L().Error("通过机器人的robot_id获取机器人失败，是一个没有注册的机器人", zap.Error(err))
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r = &DingRobot{RobotId: robotId}
	}

	//到了这里就说明这个用户有这个小机器人
	//crontab := cron.New(cron.WithSeconds()) //精确到秒
	//spec := "* 30 22 * * ?" //cron表达式，每五秒一次
	if p.MsgText == nil && p.MsgLink == nil && p.MsgMarkDown == nil {
		p.MsgText.Msgtype = "text"
		p.RepeatTime = "立即发送"
	}
	if p.MsgText.Msgtype == "text" {
		if (p.RepeatTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
			zap.L().Info("进入即时发送消息模式")
			err = r.SendMessage(p)
			if err != nil {
				return err, Task{}
			} else {
				zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
			}

			return err, task
		} else { //我要做定时任务
			tasker := func() {}
			zap.L().Info("进入定时任务模式")
			tasker = func() {
				err := r.SendMessage(p)
				if err != nil {
					return
				} else {
					zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
				}
			}
			TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
			tid = strconv.Itoa(int(TaskID))
			if err != nil {
				zap.L().Error("定时任务启动失败", zap.Error(err))
				err = ErrorSpecInvalid

				return err, Task{}
			}
			//把定时任务添加到数据库中
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				//MsgLink:           p.MsgLink,
				//MsgMarkDown:       p.MsgMarkDown,
			}
			err = (&task).InsertTask()
			if err != nil {
				zap.L().Info(fmt.Sprintf("定时任务插入数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", CurrentUser.Name, r.Name, p.DetailTime, zap.Error(err)))
				return err, Task{}
			}
			zap.L().Info(fmt.Sprintf("定时任务插入数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", CurrentUser.Name, r.Name, p.DetailTime))
		}
	} else if p.MsgLink.Msgtype == "link" {
		if (p.RepeatTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
			zap.L().Info("进入即时发送消息模式")
			err := r.SendMessage(p)
			if err != nil {
				return err, Task{}
			} else {
				zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
			}
			//定时任务
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				//MsgLink:           p.MsgLink,
				//MsgMarkDown:       p.MsgMarkDown,
			}
			return err, task
		} else { //我要做定时任务
			tasker := func() {}
			zap.L().Info("进入定时任务模式")
			tasker = func() {
				err := r.SendMessage(p)
				if err != nil {
					return
				} else {
					zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
				}
			}
			TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
			tid = strconv.Itoa(int(TaskID))
			if err != nil {
				err = ErrorSpecInvalid
				return err, Task{}
			}
			//把定时任务添加到数据库中
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				//MsgLink:           p.MsgLink,
				//MsgMarkDown:       p.MsgMarkDown,
			}
			err = (&task).InsertTask()
			if err != nil {
				zap.L().Info(fmt.Sprintf("定时任务插入数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", CurrentUser.Name, r.Name, p.DetailTime, zap.Error(err)))
				return err, Task{}
			}
			zap.L().Info(fmt.Sprintf("定时任务插入数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", CurrentUser.Name, r.Name, p.DetailTime))
		}
	} else if p.MsgMarkDown.Msgtype == "markdown" {

		if err != nil {
			zap.L().Error("通过人名查询电话号码失败", zap.Error(err))
			return
		}
		if (p.RepeatTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
			zap.L().Info("进入即时发送消息模式")
			err := r.SendMessage(p)
			if err != nil {
				return err, Task{}
			} else {
				zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
			}
			//定时任务
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				//MsgLink:           p.MsgLink,
				//MsgMarkDown:       p.MsgMarkDown,
			}
			return err, task
		} else { //我要做定时任务
			tasker := func() {}
			zap.L().Info("进入定时任务模式")
			tasker = func() {
				err := r.SendMessage(p)
				if err != nil {
					return
				} else {
					zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
				}
			}
			TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
			tid = strconv.Itoa(int(TaskID))
			if err != nil {
				err = ErrorSpecInvalid
				return err, Task{}
			}
			//把定时任务添加到数据库中
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				//MsgLink:           p.MsgLink,
				//MsgMarkDown:       p.MsgMarkDown,
			}
			err = (&task).InsertTask()
			if err != nil {
				zap.L().Info(fmt.Sprintf("定时任务插入数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", CurrentUser.Name, r.Name, p.DetailTime, zap.Error(err)))
				return err, Task{}
			}
			zap.L().Info(fmt.Sprintf("定时任务插入数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", CurrentUser.Name, r.Name, p.DetailTime))
		}
	}

	global.GLOAB_CORN.Start()

	return err, task

}

// SendMessage Function to send message
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
		}
		if p.MsgText.At.IsAtAll {
			msg["at"] = map[string]interface{}{
				"isAtAll": p.MsgText.At.IsAtAll,
			}
		} else {
			msg["at"] = map[string]interface{}{
				"atMobiles": atMobileStringArr, //字符串切片类型
				"atUserIds": atUserIdStringArr,
				"isAtAll":   p.MsgText.At.IsAtAll,
			}
		}
		b, _ = json.Marshal(msg)

	} else if p.MsgLink.Msgtype == "link" {
		//直接序列化
		b, _ = json.Marshal(p.MsgLink)
	} else if p.MsgMarkDown.Msgtype == "markdown" {
		msg := map[string]interface{}{}
		atMobileStringArr := make([]string, len(p.MsgMarkDown.At.AtMobiles))
		for i, atMobile := range p.MsgMarkDown.At.AtMobiles {
			atMobileStringArr[i] = atMobile.AtMobile
		}
		msg = map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": p.MsgMarkDown.MarkDown.Title,
				"text":  p.MsgMarkDown.MarkDown.Text,
			},
		}
		if p.MsgText.At.IsAtAll {
			msg["at"] = map[string]interface{}{
				"isAtAll": p.MsgText.At.IsAtAll,
			}
		} else {
			msg["at"] = map[string]interface{}{
				"atMobiles": atMobileStringArr, //字符串切片类型
				"isAtAll":   p.MsgText.At.IsAtAll,
			}
		}
		b, _ = json.Marshal(msg)
	} else {
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
		}
		if p.MsgText.At.IsAtAll {
			msg["at"] = map[string]interface{}{
				"isAtAll": p.MsgText.At.IsAtAll,
			}
		} else {
			msg["at"] = map[string]interface{}{
				"atMobiles": atMobileStringArr, //字符串切片类型
				"atUserIds": atUserIdStringArr,
				"isAtAll":   p.MsgText.At.IsAtAll,
			}
		}
		b, _ = json.Marshal(msg)
	}

	var resp *http.Response
	var err error
	if t.Type == "1" || t.Secret == "" {
		resp, err = http.Post(t.getURLV2(), "application/json", bytes.NewBuffer(b))
	} else {
		resp, err = http.Post(t.getURL(), "application/json", bytes.NewBuffer(b))
	}
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

func (t *DingRobot) getURL() string {
	url := "https://oapi.dingtalk.com/robot/send?access_token=" + t.RobotId //拼接token路径
	timestamp := time.Now().UnixNano() / 1e6                                //以毫秒为单位
	formatTimeStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
	zap.L().Info(fmt.Sprintf("当时时间戳对应的时间是：%s", formatTimeStr))
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, t.Secret)
	sign := t.hmacSha256(stringToSign, t.Secret)
	url = fmt.Sprintf("%s&timestamp=%d&sign=%s", url, timestamp, sign) //把timestamp和sign也拼接在一起
	return url
}
func (t *DingRobot) getURLV2() string {
	url := "https://oapi.dingtalk.com/robot/send?access_token=" + t.RobotId //拼接token路径
	return url
}

//func (t *DingRobot) StopTask(id string) (err error) {
//	task := Task{
//		TaskID: id,
//	}
//	taskID, err := mysql.StopTask(task)
//
//	if errors.Is(err, mysql.ErrorNotHasTask) {
//		return mysql.ErrorNotHasTask
//	}
//	global.GLOAB_CORN.Remove(cron.EntryID(taskID))
//	return err
//}
func SendSessionWebHook(p *ParamReveiver) (err error) {
	//currentTime := time.Now().Format("15:04:05")         //15:04:05固定写法，可以获取到当前时间的时分秒
	//formatTime, _ := time.Parse("15:04:05", currentTime) //把时间字符串转化成时间格式，时间格式可以直接比较
	//var msg map[string]interface{}
	////获取redis中的考勤记录的键
	//attendanceKey := redis.GetAttendanceKey(p.SenderStaffId, p.ConversationId)
	////判断键的情况
	//ttlAttendanceKey, err := redis.TTLAttendanceKey(attendanceKey)
	////如果@机器人的消息包含考勤，且包含三期或者四期，再加上时间限制
	//if ttlAttendanceKey == -2 && strings.Contains(p.Text.Content, "考勤") && (strings.Contains(p.Text.Content, "三期") || strings.Contains(p.Text.Content, "四期")) &&
	//	((utils.MorningStartTime.Before(formatTime) && utils.MorningEndTime.After(formatTime)) ||
	//		(utils.AfternoonStartTime.Before(formatTime) && utils.AfternoonEndtTime.Before(formatTime)) ||
	//		(utils.EveningStartTime.Before(formatTime) && utils.EveningEndTime.After(formatTime))) {
	//	attend := Attendance{
	//		Content:           p.Text.Content,
	//		ChatbotUserId:     p.ChatbotUserId,
	//		SenderNick:        p.SenderNick,
	//		ConversationName: p.ConversationName,
	//		SenderStaffId:     p.SenderStaffId,
	//		ChatBotUserId:     p.ChatbotUserId, //加密的机器人id也存入到数据库中
	//	}
	//	err := global.GLOAB_DB.Create(&attend).Error
	//	if err != nil {
	//		zap.L().Error("考勤记录插入失败", zap.Error(err))
	//		msg = map[string]interface{}{
	//			"msgtype": "text",
	//			"text": map[string]string{
	//				"content": utils.AttendanceFail,
	//			},
	//		}
	//		msg["at"] = map[string][]string{
	//			"atUserIds": []string{p.SenderStaffId},
	//		}
	//	} else {
	//		msg = map[string]interface{}{
	//			"msgtype": "text",
	//			"text": map[string]string{
	//				"content": utils.AttendanceSucc,
	//			},
	//		}
	//		msg["at"] = map[string][]string{
	//			"atUserIds": []string{p.SenderStaffId},
	//		}
	//		//在redis里面创建一个键，用来记录考勤是第一次记录还是更新
	//		err = redis.SetAttendanceState(p.SenderStaffId, p.ConversationId)
	//		if err != nil {
	//			zap.L().Error("在redis中存储考勤状态失败", zap.Error(err))
	//		}
	//	}
	//} else if ttlAttendanceKey != -2 && strings.Contains(p.Text.Content, "考勤") && (strings.Contains(p.Text.Content, "三期") || strings.Contains(p.Text.Content, "四期")) &&
	//	((utils.MorningStartTime.Before(formatTime) && utils.MorningEndTime.After(formatTime)) ||
	//		(utils.AfternoonStartTime.Before(formatTime) && utils.AfternoonEndtTime.Before(formatTime)) ||
	//		(utils.EveningStartTime.Before(formatTime) && utils.EveningEndTime.After(formatTime))) {
	//	attend := model.Attendance{
	//		Content:           p.Text.Content,
	//		ChatbotUserId:     p.ChatbotUserId,
	//		SenderNick:        p.SenderNick,
	//		ConversationName: p.ConversationName,
	//		SenderStaffId:     p.SenderStaffId,
	//		ChatBotUserId:     p.ChatbotUserId, //加密的机器人id也存入到数据库中
	//	}
	//	err := global.GLOAB_DB.Create(&attend).Error
	//	if err != nil {
	//		zap.L().Error("考勤记录插入失败", zap.Error(err))
	//		msg = map[string]interface{}{
	//			"msgtype": "text",
	//			"text": map[string]string{
	//				"content": utils.AttendanceFail,
	//			},
	//		}
	//		msg["at"] = map[string][]string{
	//			"atUserIds": []string{p.SenderStaffId},
	//		}
	//	} else {
	//		msg = map[string]interface{}{
	//			"msgtype": "text",
	//			"text": map[string]string{
	//				"content": utils.AttendanceUpdateSucc,
	//			},
	//		}
	//		msg["at"] = map[string][]string{
	//			"atUserIds": []string{p.SenderStaffId},
	//		}
	//	}
	//} else if strings.Contains(p.Text.Content, "打字邀请码") {
	//	//去redis中取一下打字邀请码
	//	var TypingInviationCode string
	//	var expire1 int64
	//	fmt.Println(expire1)
	//	expire, err := global.GLOBAL_REDIS.TTL(context.Background(), utils.ConstTypingInvitationCode).Result()
	//	if err != nil {
	//		zap.L().Error("判断token剩余生存时间失败", zap.Error(err))
	//	}
	//	//如果redis里面没有的话
	//	if expire == -2 {
	//		//申请新的TypingInviationCode并已经存入redis
	//		TypingInviationCode, err = utils.TypingInviation()
	//		if err != nil || TypingInviationCode == "" {
	//			zap.L().Error("申请新的TypingInviationCode失败", zap.Error(err))
	//			msg = map[string]interface{}{
	//				"msgtype": "text",
	//				"text": map[string]string{
	//					"content": utils.TypingInviationFail,
	//				},
	//			}
	//			msg["at"] = map[string][]string{
	//				"atUserIds": []string{p.SenderStaffId},
	//			}
	//		}
	//
	//	} else {
	//		//从redis从取到邀请码
	//		TypingInviationCode = global.GLOBAL_REDIS.Get(context.Background(), utils.ConstTypingInvitationCode).Val()
	//		if len(TypingInviationCode) != 5 {
	//			zap.L().Error("申请新的TypingInviationCode失败", zap.Error(err))
	//			msg = map[string]interface{}{
	//				"msgtype": "text",
	//				"text": map[string]string{
	//					"content": utils.TypingInviationFail,
	//				},
	//			}
	//			msg["at"] = map[string][]string{
	//				"atUserIds": []string{p.SenderStaffId},
	//			}
	//		} else {
	//			msg = map[string]interface{}{
	//				"msgtype": "text",
	//				"text": map[string]string{
	//					"content": utils.TypingInviationSucc + ":" + TypingInviationCode,
	//				},
	//			}
	//			msg["at"] = map[string][]string{
	//				"atUserIds": []string{p.SenderStaffId},
	//			}
	//		}
	//
	//	}
	//} else if strings.Contains(p.Text.Content, "加密机器人ID") {
	//	msg = map[string]interface{}{
	//		"msgtype": "text",
	//		"text": map[string]string{
	//			"content": "获取成功：" + p.ChatbotUserId + "\n" + "登录机器人后台，更新机器人填写此字段后即可查看该机器人考勤记录",
	//		},
	//	}
	//	msg["at"] = map[string][]string{
	//		"atUserIds": []string{p.SenderStaffId},
	//	}
	//}
	//
	//b, err := json.Marshal(msg)
	//if err != nil {
	//	return err
	//}
	//var resp *http.Response
	//
	//resp, err = http.Post(p.SessionWebhook, "application/json", bytes.NewBuffer(b))
	//
	//defer resp.Body.Close()
	//date, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(date)
	//if err != nil {
	//	return err
	//}
	return nil
}
func HandleSpec(p *ParamCronTask) (spec, detailTimeForUser string, err error) {
	spec = ""
	detailTimeForUser = ""
	n := len(p.DetailTime)
	if p.RepeatTime == "1" {
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

//获取机器人所在的群聊的userIdList ，前提是获取到OpenConversationId，获取到OpenConverstaionId的前提是获取到二维码
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
		OpenConversationId: tea.String(r.OpenConversationID),
		CoolAppCode:        tea.String("COOLAPP-1-102118DC0ABA212C89C7000H"),
		MaxResults:         tea.Int64(300),
		NextToken:          tea.String("XXXXX"),
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
	_result = &dingtalkim_1_0.Client{}
	_result, _err = dingtalkim_1_0.NewClient(config)
	return _result, _err
}

func GetImage(c *gin.Context) { //显示图片的方法
	imageName := c.Query("imageName")     //截取get请求参数，也就是图片的路径，可是使用绝对路径，也可使用相对路径
	file, _ := ioutil.ReadFile(imageName) //把要显示的图片读取到变量中
	c.Writer.WriteString(string(file))    //关键一步，写给前端
}
func (t *DingRobot) StopTask(taskId string) (err error) {
	//先来判断一下是否拥有这个定时任务
	var task Task
	err = global.GLOAB_DB.Where("task_id = ?", taskId).First(&task).Error
	if err != nil {
		zap.L().Info("通过taskId查找定时任务失败", zap.Error(err))
		return err
	}
	taskID, err := strconv.Atoi(task.TaskID)
	if err != nil {
		return err
	}
	//到了这里就说明我有这个定时任务，我要移除这个定时任务
	err = global.GLOAB_DB.Delete(&task).Error
	if err != nil {
		zap.L().Error("删除定时任务失败", zap.Error(err))
		return err
	}
	global.GLOAB_CORN.Remove(cron.EntryID(taskID))
	return err
}
func (t *DingRobot) GetTaskList(RobotId string) (tasks []Task, err error) {
	err = global.GLOAB_DB.Model(&DingRobot{RobotId: RobotId}).Unscoped().Association("Tasks").Find(&tasks) //通过机器人的id拿到机器人，拿到机器人后，我们就可以拿到所有的任务
	if err != nil {
		zap.L().Error("通过机器人robot_id拿到该机器人的所有定时任务失败", zap.Error(err))
		return
	}
	return
}
func (t *DingRobot) RemoveTask(taskId string) (err error) {
	//先来判断一下是否拥有这个定时任务
	var task Task
	err = global.GLOAB_DB.Where("task_id = ?", taskId).First(&task).Error
	if err != nil {
		zap.L().Info("通过taskId查找定时任务失败", zap.Error(err))
		return err
	}
	taskID, err := strconv.Atoi(task.TaskID)
	if err != nil {
		return err
	}
	//到了这里就说明我有这个定时任务，我要移除这个定时任务
	err = global.GLOAB_DB.Unscoped().Delete(&task).Error
	if err != nil {
		zap.L().Error("删除定时任务失败", zap.Error(err))
		return err
	}
	global.GLOAB_CORN.Remove(cron.EntryID(taskID))
	return err
}
func (t *DingRobot) GetUnscopedTaskByID(id string) (task Task, err error) {
	err = global.GLOAB_DB.Unscoped().Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").First(&task, id).Error
	if err != nil {
		zap.L().Error("通过主键id查询定时任务失败", zap.Error(err))
		return
	}
	return
}
func (t *DingRobot) ReStartTask(id string) (task Task, err error) {
	err = global.GLOAB_DB.Model(&Task{}).First(&task, id).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Task{}, errors.New("该定时任务没有暂停，所以无法重启")
	}
	task, err = t.GetUnscopedTaskByID(id)
	//根据这个id主键查询到被删除的数据
	err = global.GLOAB_DB.Unscoped().Model(&task).Update("deleted_at", nil).Error //这个地方必须加上Unscoped()，否则不报错，但是却无法更新
	p := ParamCronTask{
		MsgText:     task.MsgText,
		MsgLink:     task.MsgLink,
		MsgMarkDown: task.MsgMarkDown,
		RobotId:     task.RobotId,
	}
	d := DingRobot{
		RobotId: task.RobotId,
		Secret:  task.Secret,
	}
	tasker := func() {
		err := d.SendMessage(&p)
		if err != nil {
			//zap.L().Error(fmt.Sprintf("恢复任务失败！发送人:%s,对应机器人:%s", username, robotname), zap.Error(err))
			return
		} else {
			//zap.L().Info(fmt.Sprintf("恢复任务成功！发送人:%s,对应机器人:%s", username, robotname))
		}
	}
	//	// 添加定时任务
	TaskID, err := global.GLOAB_CORN.AddFunc(task.Spec, tasker)
	if err != nil {
		//zap.L().Error("项目重启后恢复定时任务失败,失败原因：", zap.Error(err))
		//zap.L().Error(fmt.Sprintf("该任务所属人：%s,所属机器人：%s,"+
		//"人物名：%s,任务具体消息:%s,任务具体定时规则：%s", username, robotname, message, detailTimeForUser))
		return
	}
	tid := int(TaskID)
	oldId := task.TaskID
	err = global.GLOAB_DB.Table("tasks").Where("task_id = ? ", oldId).Update("task_id", tid).Error
	if err != nil {
		//zap.L().Error("重启项目后更新任务id失败", zap.Error(err))
		return
	}
	return
}

type result struct {
	ChatId string `json:"chatId"`
	Title  string `json:"title"`
}
type data struct {
	Result result `json:"result"`
}
