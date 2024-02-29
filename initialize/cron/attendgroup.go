package cron

import (
	"ding/global"
	"ding/model/common"
	"ding/model/dingding"
	"ding/model/params"
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
		if group.IsRobotAttendance {
			p := &params.ParamAllDepartAttendByRobot{GroupId: group.GroupId}
			//正常考勤
			_, taskID, err := group.AllDepartAttendByRobot(p)
			if err != nil {
				return err
			}
			//提醒没有打开的人考勤
			group.AlertAttend(p)
			err = global.GLOAB_DB.Model(&group).Update("robot_attend_task_id", int(taskID)).Error
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
