package cron

import (
	"ding/global"
	"ding/model/dingding"
	"go.uber.org/zap"
)

func weeklyCheckByRobot() (err error) {
	reprotDescribe := &dingding.ReportDescribe{}
	//每周日晚上10点
	reportSpec := "0 0 22 * * 0"
	weekFun := reprotDescribe.WeekCheckByRobot
	//reprotDescribe.WeekCheckByRobot()
	_, err = global.GLOAB_CORN.AddFunc(reportSpec, weekFun)
	if err != nil {
		zap.L().Error("启动周报检测失败", zap.Error(err))
		return
	}

	//指定群发消息确定
	//d := &dingding.ParamCronTask{
	//	MsgText: &common.MsgText{
	//		Msgtype: "text",
	//		At:      common.At{AtMobiles: []common.AtMobile{{AtMobile: "18737480171"}}},
	//		Text: common.Text{
	//			Content: fmt.Sprintf("考勤组：%v 开启机器人考勤\n提醒考勤定时规则：%v\n考勤定时规则:%v\n ", group.GroupName, AlertSpec, AttendSpec),
	//		},
	//	},
	//}
	//zap.L().Info(fmt.Sprintf("考勤组：%v 开启机器人考勤", group.GroupName))
	//err = (&dingding.DingRobot{RobotId: "b14ef369d04a9bbfc10f3092d58f7214819b9daa93f3998121661ea0f9a80db3"}).SendMessage(d)
	//if err != nil {
	//	zap.L().Error("发送错误", zap.Error(err))
	//}

	return err
}
