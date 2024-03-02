package cron

import (
	"ding/global"
	"ding/model/common"
	"ding/model/dingding"
	"ding/model/params"
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
			p := &params.ParamAllDepartAttendByRobot{GroupId: group.GroupId}
			//正常考勤
			AttendTaskID, err := group.AllDepartAttendByRobot()
			if err != nil {
				return err
			}
			var AlertTaskID cron.EntryID
			if group.IsAlert {
				//提醒没有打卡的人考勤
				AlertTaskID, err = group.AlertAttendByRobot(p)
				if err != nil {
					return err
				}
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
						Content: fmt.Sprintf("考勤组：%v 开启机器人考勤", group.GroupName),
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
