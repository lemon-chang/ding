package robot

import (
	"context"
	"ding/model/chatRobot"
	"ding/model/dingding"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"go.uber.org/zap"
	"strings"
)

func RobotOutGoing() {
	var clientId, clientSecret string
	clientId = "dinglyjekzn80ebnlyge"
	clientSecret = "zKsishrdIL3h3LvWSV2Sm0G1WBP5uVJgDpRnzUmli564HJO78W_10GIFEBrM9r-C"

	cli := client.NewStreamClient(client.WithAppCredential(client.NewAppCredentialConfig(clientId, clientSecret)))
	//robot的路由
	cli.RegisterChatBotCallbackRouter(OnChatBotMessageReceived)

	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}
}

func OnChatBotMessageReceived(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	handlerMsg := messagesHandler(data)
	//replyMsg := []byte(fmt.Sprintf("输入的内容为: [%s]", strings.TrimSpace(data.Text.Content)))
	replier := chatbot.NewChatbotReplier()
	replyMsg := []byte(handlerMsg)
	if err := replier.SimpleReplyText(ctx, data.SessionWebhook, replyMsg); err != nil {
		return nil, err
	}
	return []byte(""), nil
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
