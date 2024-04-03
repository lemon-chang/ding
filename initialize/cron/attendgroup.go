package cron

import (
	"ding/global"
	"ding/initialize/viper"
	"ding/model/common"
	"ding/model/dingding"
	"fmt"
	"go.uber.org/zap"
)

func AttendanceByRobot() (err error) {
	var groupList []dingding.DingAttendGroup
	err = global.GLOAB_DB.Find(&groupList).Error
	if err != nil {
		return
	}
	for _, group := range groupList {
		//根据考勤组id获取成员信息

		//提醒没有打卡的人考勤
		_, AlertSpec, err := group.AlertAttendByRobot(group.GroupId)
		if err != nil {
			return err
		}

		//正常考勤
		// 此处记录一下问题，如果使用不传递group.GroupId 的话，会一直考勤最后一个部门的考勤，暂时没有解决方法
		_, AttendSpec, err := group.AllDepartAttendByRobot(group.GroupId)
		if err != nil {
			return err
		}

		d := &dingding.ParamCronTask{
			MsgText: &common.MsgText{
				Msgtype: "text",
				At:      common.At{AtMobiles: []common.AtMobile{{AtMobile: "18737480171"}}},
				Text: common.Text{
					Content: fmt.Sprintf("考勤组：%v 开启机器人考勤\n提醒考勤定时规则：%v\n考勤定时规则:%v\n ", group.GroupName, AlertSpec, AttendSpec),
				},
			},
		}
		zap.L().Info(fmt.Sprintf("考勤组：%v 开启机器人考勤", group.GroupName))
		err = (&dingding.DingRobot{RobotId: viper.Conf.MiniProgramConfig.RobotToken}).SendMessage(d)
		if err != nil {
			zap.L().Error("发送错误", zap.Error(err))
		}

	}
	return err
}
