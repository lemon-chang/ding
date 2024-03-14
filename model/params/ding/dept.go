package ding

type ParamUpdateDept struct {
	DeptID             int      `json:"dept_id" validate:"required"`
	IsSendFirstPerson  int      `json:"is_send_first_person"` // 0为不推送，1为推送
	RobotToken         string   `json:"robot_token"`
	IsRobotAttendance  int      `json:"is_robot_attendance"`
	IsJianshuOrBlog    int      `json:"is_jianshu_or_blog"`
	IsLeetCode         int      `json:"is_leet_code"`
	ResponsibleUserIds []string `json:"ResponsibleUserIds"`
}
type ParamSendFrequencyLeaveDept struct {
	DeptID int `json:"dept_id"`
}
type ParamSendFrequencyLeaveUser struct {
	UserID int `json:"user_id"`
}
type ParameIsInSchool struct {
	GroupId    int  `json:"group_id"`
	IsInSchool bool `json:"is_in_school"` //是否在学校，如果在学校，开启判断是否有课
}
