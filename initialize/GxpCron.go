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
		var records []dingding.Record
		err = global.GLOAB_DB.Where("created_at like ?", "%"+day+"%").Find(&records).Error
		if err != nil {
			zap.L().Error(fmt.Sprintf("查询%s的数据失败", day), zap.Error(err))
			return
		}
		//将records中的id提取出来
		notAtRobotUserIds := make([]common.AtUserId, 0)
		//通过每条数据中的userid查询同学姓名
		for _, record := range records {
			notAtRobotUserIds = append(notAtRobotUserIds, common.AtUserId{AtUserId: record.TongXinUserID})
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
	fmt.Println("关鑫鹏22：00定时任务", entryID)
	return
}
