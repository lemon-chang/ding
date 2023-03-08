package initialize

import (
	"ding/global"
	"ding/model/dingding"
	"ding/model/params"
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
		}
	}
	return err
}
