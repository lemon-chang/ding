package ding

type ParamUpdateDeptToCron struct {
	DeptID            int    `json:"dept_id"`
	IsSendFirstPerson int    `json:"is_send_first_person"` // 0为不推送，1为推送
	RobotToken        string `json:"robot_token"`
	IsRobotAttendance int    `json:"is_robot_attendance"`
	IsJianshuOrBlog   int    `json:"is_jianshu_or_blog"`
}
