package initialize

import (
	"ding/global"
	"ding/model/common"
	"ding/model/dingding"
	"fmt"
	"go.uber.org/zap"
	_ "strconv"
	"time"
)

const RobotToken = "11e07612181c7b596e49e80d26cb368318a2662c0f6affd453ccfd3d906c2431"

func CronSendOne() (err error) {
	spec := "0 0 22 ? * * "
	//开启定时器，定时每晚10：00(cron定时任务的创建)
	entryID, err := global.GLOAB_CORN.AddFunc(spec, func() {
		message := "已经十点了，请大家未回到宿舍的同学及时返回宿舍休息，回到宿舍的同学在群中@我发送 已到宿舍 ，请假同学在群中@我发送  已请假，谢谢大家配合"
		fmt.Println(message)

		zap.L().Info("message编辑完成，开始封装发送信息参数")
		p := &dingding.ParamCronTask{
			MsgText: &common.MsgText{
				At: common.At{
					IsAtAll: true,
				},
				Text: common.Text{
					Content: message,
				},
				Msgtype: "text",
			},
			RobotId: RobotToken,
		}
		err := (&dingding.DingRobot{
			RobotId: RobotToken,
		}).SendMessage(p)
		if err != nil {
			zap.L().Error("发送关鑫鹏22：00定时任务失败", zap.Error(err))
			return
		}
	})
	fmt.Println("关鑫鹏22：00定时任务", entryID)
	return
}

func CronSendTwo() (err error) {
	spec := "0 20 22 ? * * "
	//开启定时器，定时22：20提醒未到宿舍人员(cron定时任务的创建)
	entryID, err := global.GLOAB_CORN.AddFunc(spec, func() {
		day := time.Now().Format("2006-01-02")
		var AllUsers []dingding.TongXinUser
		var atRobotUsers []dingding.TongXinUser
		var notAtRobotUserIds []common.AtUserId
		global.GLOAB_DB1.Preload("Records", "created_at like ?", "%"+day+"%").Find(&AllUsers)
		for _, user := range AllUsers {
			if user.Records == nil || len(user.Records) == 0 {
				notAtRobotUserIds = append(notAtRobotUserIds, common.AtUserId{
					AtUserId: user.ID,
				})
			} else {
				atRobotUsers = append(atRobotUsers, user)
			}
		}

		message := "截至目前还有以下同学未报备是否到达宿舍：\n"

		zap.L().Info("message编辑完成，开始封装发送信息参数")
		p := &dingding.ParamCronTask{
			MsgText: &common.MsgText{
				At: common.At{
					AtUserIds: notAtRobotUserIds,
				},
				Text: common.Text{
					Content: message,
				},
				Msgtype: "text",
			},
			RobotId: RobotToken,
		}
		err := (&dingding.DingRobot{
			RobotId: RobotToken,
		}).SendMessage(p)
		if err != nil {
			zap.L().Error("发送关鑫鹏22：20定时任务失败", zap.Error(err))
			return
		}
	})
	fmt.Println("关鑫鹏22：20定时任务", entryID)
	return
}

// CronSendThree 晚上10：35统计结果发给gxp
func CronSendThree() (err error) {
	spec := "0 35 22 ? * * "
	//开启定时器，定时22：20提醒未到宿舍人员(cron定时任务的创建)
	entryID, err := global.GLOAB_CORN.AddFunc(spec, func() {
		day := time.Now().Format("2006-01-02")
		var AllUsers []dingding.TongXinUser
		var atRobotUsers []dingding.TongXinUser
		var notAtRobotUsers []dingding.TongXinUser
		global.GLOAB_DB1.Preload("Records", "created_at like ?", "%"+day+"%").Find(&AllUsers)
		for _, user := range AllUsers {
			if user.Records == nil || len(user.Records) == 0 {
				notAtRobotUsers = append(notAtRobotUsers, user)
			} else {
				atRobotUsers = append(atRobotUsers, user)
			}
		}

		message := "截至目前还有以下同学未报备是否到达宿舍：\n"
		for _, notAtRobotUser := range notAtRobotUsers {
			message += notAtRobotUser.Name + "  "
		}
		message += "\n" + "已发送消息的同学：\n"
		for i, atRobotUser := range atRobotUsers {
			message += atRobotUser.Name + ",发送消息内容：" + atRobotUser.Records[i].Content + "\n"
		}

		zap.L().Info("message编辑完成，开始封装发送信息参数")
		//关鑫鹏个人的userid
		var userId = []string{"01144160064621256183"}
		p := &dingding.ParamChat{
			RobotCode: RobotToken,
			UserIds:   userId,
			MsgKey:    "sampleText",
			MsgParam:  message,
		}
		err := (&dingding.DingRobot{
			RobotId: RobotToken,
		}).CommonSingleChat(p)
		//p := &dingding.ParamCronTask{
		//	MsgText: &common.MsgText{
		//		At: common.At{
		//			AtUserIds: notAtRobotUserIds,
		//		},
		//		Text: common.Text{
		//			Content: message,
		//		},
		//		Msgtype: "text",
		//	},
		//	RobotId: RobotToken,
		//}
		//err := (&dingding.DingRobot{
		//	RobotId: RobotToken,
		//}).SendMessage(p)
		if err != nil {
			zap.L().Error("发送关鑫鹏22：35定时任务失败", zap.Error(err))
			return
		}
	})
	fmt.Println("关鑫鹏22：35定时任务", entryID)
	return
}
