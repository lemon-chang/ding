package ding

import "ding/model/common/request"

type ParamUpdateUpdateAttendanceGroup struct {
	GroupId           int    `gorm:"primaryKey" json:"group_id"` //考勤组id
	IsRobotAttendance bool   `json:"is_robot_attendance"`        //该考勤组是否开启机器人查考勤 （相当于是总开关）
	IsSendFirstPerson int    `json:"is_send_first_person"`       //该考勤组是否开启推送每个部门第一位打卡人员 （总开关）
	IsInSchool        bool   `json:"is_in_school"`               //是否在学校，如果在学校，开启判断是否有课
	AlertTime         int    `json:"alert_time"`                 //如果预备了，提前几分钟
	DelayTime         int    `json:"delay_time"`                 //推迟多少分钟
	NextTime          string `json:"next_time"`                  //下次执行时间
	IsWeekPaper       bool   `json:"is_week_paper"`              // 是否开启周报提醒
}
type ParamGetAttendGroup struct {
	request.PageInfo
	Name string `json:"name" form:"name"`
}
