package dingding

import (
	"ding/model/common"
	"ding/model/common/request"
)

type ParamAddRobot struct {
	RobotId  string `json:"robot_id" validate:"required,len=64"` //机器人的token //这个后面的json标签可以改变我们返回此结构体数据的字段,同时此字段也需要和前端保持一致
	Name     string `json:"name" validate:"required"`
	IsShared int    `json:"is_shared" validate:"oneof=0 1"`
}
type ParamGetRobotList struct {
	request.PageInfo
	Name string `json:"name" form:"name" `
}
type ParamGetRobotBase struct {
	RobotId string `json:"robot_id" form:"robot_id"` //机器人的token //这个后面的json标签可以改变我们返回此结构体数据的字段,同时此字段也需要和前端保持一致
}
type ParamGetRobotListBase struct {
	request.PageInfo
	UserId string `json:"user_id" form:"user_id"` //机器人的token //这个后面的json标签可以改变我们返回此结构体数据的字段,同时此字段也需要和前端保持一致
}
type ParamRemoveRobot struct {
	RobotIds []string `json:"robot_id" binding:"required"`
}
type ParamPingRobot struct {
	Version string
	RobotId string `binding:"required" json:"robot_id"`
}
type ParamRemoveTask struct {
	ID uint `json:"id"`
}
type UpdateTask struct {
	ID                uint            `json:"id"`
	IsSuspend         bool            `json:"is_suspend"`   //是否暂停
	TaskName          string          `json:"task_name"`    //任务名字
	MsgText           *common.MsgText `json:"msg_text"`     // content和要@的人
	RepeatTime        string          `json:"repeat_time" ` //前端给的重复频率，仅重复一次，周重复，月重复
	DetailTime        string          `json:"detail_time"`  //在给定的重复频率下的具体执行时间
	Spec              string          `json:"spec"`         // 给懂cron定时库的人用
	DetailTimeForUser string          `json:"detail_time_for_user"`
	RobotId           string          `json:"robot_id" binding:"required"` //使用机器人的robot_id来确定机器人
}

type ParamGetTaskDeatil struct {
	ID uint `json:"id" form:"id"`
}
type ParamGetTaskList struct {
	request.PageInfo
	RobotId   string `json:"robot_id"`
	IsSuspend bool   `json:"is_suspend"` //是否暂停
	TaskName  string `json:"task_name"`
}
type ParamChat struct {
	RobotCode          string   `json:"robotCode"`
	UserIds            []string `json:"userIds"`
	MsgKey             string   `json:"msgKey"`             //SampleText
	MsgParam           string   `json:"msgParam"`           //具体内容
	OpenConversationId string   `json:"openConversationId"` //群发时会用
}
type ParamCronTask struct {
	MsgText     *common.MsgText     `json:"msg_text"`
	MsgLink     *common.MsgLink     `json:"msg_link"`
	MsgMarkDown *common.MsgMarkDown `json:"msg_mark_down"`
	RobotId     string              `json:"robot_id" binding:"required"`  //使用机器人的robot_id来确定机器人
	RepeatTime  string              `json:"repeat_time" `                 //前端给的重复频率，仅重复一次，周重复，月重复
	DetailTime  string              `json:"detail_time"`                  //在给定的重复频率下的具体执行时间
	TaskName    string              `json:"task_name" binding:"required"` //给这个任务起一个名字
	Spec        string              `json:"spec"`                         //通过spec进行调用
	ID          uint                `json:"id"`
}

type ParamUpdateRobot struct {
	ID                 uint       `json:"id"`
	Type               int        `json:"type"`     //机器人类型
	RobotId            string     `json:"robot_id"` //机器人的token //这个后面的json标签可以改变我们返回此结构体数据的字段,同时此字段也需要和前端保持一致
	ChatBotUserId      string     `json:"chat_bot_user_id"`
	Secret             string     `json:"secret"`
	DingUsers          []DingUser `json:"ding_users"`
	UserName           string     `json:"user_name"`
	ChatId             string     `json:"chat_id"`
	OpenConversationID string     `json:"open_conversation_id"`
	Name               string     `json:"title"`
}
type ParamReveiver struct {
	Header
	Body
}
type Header struct {
	Content_Type string `header:"Content-Type"`
	Timestamp    string `header:"Timestamp"`
	Sign         string `header:"Sign"`
}
type Body struct {
	SenderId                  string    `json:"senderId"`                  //加密的发送者ID。
	ConversationId            string    `json:"conversationId"`            //加密的会话ID。
	AtUsers                   []atUsers `json:"atUsers"`                   //被@人的信息。
	ChatbotCorpId             string    `json:"chatbotCorpId"`             //加密的机器人所在的企业corpId。
	ChatbotUserId             string    `json:"chatbotUserId"`             //加密的机器人id
	MsgId                     string    `json:"msgId"`                     //加密的消息ID。
	SenderNick                string    `json:"senderNick"`                //发送者昵称。
	IsAdmin                   bool      `json:"isAdmin"`                   //是否是管理员
	SenderStaffId             string    `json:"senderStaffId"`             //企业内部群中@该机器人的成员userid。
	SessionWebhookExpiredTime int64     `json:"sessionWebhookExpiredTime"` //当前会话的Webhook地址过期时间。
	CreateAt                  int64     `json:"createAt"`                  //消息的时间戳，单位ms。
	SenderCorpId              string    `json:"senderCorpId"`              //企业内部群有的发送者当前群的企业corpId。
	ConversationTitle         string    `json:"conversationTitle"`         //群聊时才有的会话标题。
	IsInAtList                bool      `json:"isInAtList"`                //是否在@列表中。
	SessionWebhook            string    `json:"sessionWebhook"`            //当前会话的Webhook地址
	Text                      text      `json:"text"`                      //机器人收到的信息
	Msgtype                   string    `json:"msgtype"`                   //目前只支持text
	ConversationType          string    `json:"conversationType"`
}
type atUsers struct {
	DingtalkId string `json:"dingtalkId"`
	StaffId    string `json:"staffId"`
}
type text struct {
	Content string `json:"content"`
}

type ParamAlterResultByRobot struct {
	DeptId int    `json:"dept_id"`
	Token  string `json:"token"`
}
