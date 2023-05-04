package initialize

import (
	"ding/global"
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
			zap.L().Info(fmt.Sprintf("考勤组：%v 开启机器人考勤", group.GroupName))
			p := &params.ParamAllDepartAttendByRobot{GroupId: group.GroupId}
			_, taskID, err := group.AllDepartAttendByRobot(p)
			if err != nil {
				return err
			}
			err = global.GLOAB_DB.Model(&group).Update("robot_attend_task_id", int(taskID)).Error
			if err != nil {
				return err
			}
		} else {
			zap.L().Warn(fmt.Sprintf("考勤组：%v 开启未机器人考勤", group.GroupName))
		}
	}
	return err
}
