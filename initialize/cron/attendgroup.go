package cron

import (
	"ding/global"
	"ding/model/common"
	"ding/model/dingding"
	"fmt"
	"github.com/robfig/cron/v3"
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
		if group.IsRobotAttendance {
			var AlertTaskID cron.EntryID
			var AlertSpec, AttendSpec string
			if group.AlertTime != 0 {
				//提醒没有打卡的人考勤
				AlertTaskID, AlertSpec, err = group.AlertAttendByRobot(group.GroupId)
				if err != nil {
					return err
				}
			}
			//正常考勤
			// 此处记录一下问题，如果使用不传递group.GroupId 的话，会一直考勤最后一个部门的考勤，暂时没有解决方法
			AttendTaskID, AttendSpec, err := group.AllDepartAttendByRobot(group.GroupId)
			if err != nil {
				return err
			}

			err = global.GLOAB_DB.Model(&group).Updates(dingding.DingAttendGroup{RobotAttendTaskID: int(AttendTaskID), RobotAlterTaskID: int(AlertTaskID)}).Error
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
			err = (&dingding.DingRobot{RobotId: "b14ef369d04a9bbfc10f3092d58f7214819b9daa93f3998121661ea0f9a80db3"}).SendMessage(d)
			if err != nil {
				zap.L().Error("发送错误", zap.Error(err))
			}
		} else {
			zap.L().Warn(fmt.Sprintf("考勤组：%v 未开启机器人考勤", group.GroupName))
		}
	}
	return err
}
