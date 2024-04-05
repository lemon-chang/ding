package outgoing

import (
	"context"
	"crypto/tls"
	"ding/initialize/viper"
	"ding/model/chatRobot"
	"ding/model/dingding"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/event"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func Init() (err error) {
	clientId := viper.Conf.MiniProgramConfig.AppKey
	clientSecret := viper.Conf.MiniProgramConfig.AppSecret
	cli := client.NewStreamClient(client.WithAppCredential(client.NewAppCredentialConfig(clientId, clientSecret)))
	// 注册机器人回调robot的路由
	cli.RegisterChatBotCallbackRouter(OnChatBotMessageReceived)
	// 注册事件回调函数
	cli.RegisterAllEventRouter(OnEventReceived)
	err = cli.Start(context.Background())
	return
}

func OnChatBotMessageReceived(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	//handlerMsg := messagesHandler(data)
	replyMsg := []byte(fmt.Sprintf("输入的内容为: [%s]", strings.TrimSpace(data.Text.Content)))
	replier := chatbot.NewChatbotReplier()
	//replyMsg := []byte(handlerMsg)
	if err := replier.SimpleReplyText(ctx, data.SessionWebhook, replyMsg); err != nil {
		return nil, err
	}
	return []byte(""), nil
}

type MyEngine struct {
	Engine *gin.Engine
}

func OnEventReceived(_ context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error) {
	eventHeader := event.NewEventHeaderFromDataFrame(df)
	if eventHeader.EventType == "chat_update_title" {
		fmt.Println("chat_update_title")
		// ignore events not equals `chat_update_title`; 忽略`chat_update_title`之外的其他事件；
		// 该示例仅演示 chat_update_title 类型的事件订阅；
		return event.NewSuccessResponse()
	}

	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte

	//`这里请注意，使用 InsecureSkipVerify: true 来跳过证书验证`
	client = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(100) * time.Millisecond}
	//marshal, _ := json.Marshal(df.Data)
	url := fmt.Sprintf("http://localhost:8889/dingEvent/%s", eventHeader.EventType)
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(df.Data))
	if err != nil {

	}
	fmt.Println()
	//发送请求
	resp, err = client.Do(request)
	if err != nil {

	}
	defer resp.Body.Close()
	//读取返回的结果
	body, err = ioutil.ReadAll(resp.Body)
	//把结果绑定到到对象上面并反序列化为json类型
	fmt.Println(body)
	defer resp.Body.Close()
	return event.NewSuccessResponse()
	//global.GLOBAL_GIN_Engine.ServeDINGEVENT("/" + eventHeader.EventType) // 此处需要添加一个 / ，和路由树的匹配规则一样

	//fmt.Printf("received event, delay=%s, eventType=%s, eventId=%s, eventBornTime=%d, eventCorpId=%s, eventUnifiedAppId=%s, data=%s",
	//	time.Duration(time.Now().UnixMilli()-eventHeader.EventBornTime)*time.Millisecond,
	//	eventHeader.EventType,
	//	eventHeader.EventId,
	//	eventHeader.EventBornTime,
	//	eventHeader.EventCorpId,
	//	eventHeader.EventUnifiedAppId,
	//	df.Data)
	//// put your code here; 可以在这里添加你的业务代码，处理事件订阅的业务逻辑；
	//// 构造一颗前缀树，用来捕捉事件
	//
	//return event.NewSuccessResponse()
}

// id 用户id 部门id 个人/部门/全体 内容
func messagesHandler(data *chatbot.BotCallbackDataModel) string {
	s := ""
	//个人消息
	robot := chatRobot.RobotStream{UserId: data.SenderStaffId}
	personMsgs, err := robot.GetPersonMsg(strings.TrimSpace(data.Text.Content))
	if err != nil {
		zap.L().Error("robotStream获取个人存储信息失败", zap.Error(err))
		panic("robotStream获取个人存储信息失败")
	}
	for k, v := range personMsgs {
		s += k + ": " + v
		s += "\n"
	}
	//部门信息
	user := dingding.DingUser{UserId: data.SenderStaffId}
	dept, _ := user.GetUserDeptIdByUserId()
	robot.DeptId = dept.DeptId
	deptMsg, err := robot.GetDeptMsg(strings.TrimSpace(data.Text.Content))
	if err != nil {
		zap.L().Error("robotStream获取用户存储信息失败", zap.Error(err))
		panic("robotStream获取用户存储信息失败")
	}
	for k, v := range deptMsg {
		s += k + ": " + v
		s += "\n"
	}
	//全体部门信息
	allUserMsg, err := robot.GetAllUserMsg(strings.TrimSpace(data.Text.Content))
	for k, v := range allUserMsg {
		s += k + ": " + v
		s += "\n"
	}
	return s
}
