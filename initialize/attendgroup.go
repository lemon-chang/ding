package initialize

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
		if group.IsRobotAttendance {
			p := &params.ParamAllDepartAttendByRobot{GroupId: group.GroupId}
			_, taskID, err := group.AllDepartAttendByRobot(p)
			if err != nil {
				return err
			}
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
			(&dingding.DingRobot{RobotId: "aba857cf3ba132581d1a99f3f5c9c5fe2754ffd57a3e7929b6781367b9325e40"}).
				SendMessage(d)
		} else {
			zap.L().Warn(fmt.Sprintf("考勤组：%v 开启未机器人考勤", group.GroupName))
		}
	}
	return err
}
