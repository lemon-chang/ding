package ding

type ParamUpdateUpdateAttendanceGroup struct {
	GroupId           int    `gorm:"primaryKey" json:"group_id"` //考勤组id
	GroupName         string `json:"group_name"`                 //考勤组名称
	MemberCount       int    `json:"member_count"`               //参与考勤人员总数
	IsRobotAttendance bool   `json:"is_robot_attendance"`        //该考勤组是否开启机器人查考勤 （相当于是总开关）
	IsSendFirstPerson int    `json:"is_send_first_person"`       //该考勤组是否开启推送每个部门第一位打卡人员 （总开关）
	TaskID            int    `json:"task_id"`
}
type ParamGetAttendGroup struct {
}
