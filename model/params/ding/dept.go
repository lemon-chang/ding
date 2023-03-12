package ding

type ParamUpdateDept struct {
	DeptID            int    `json:"dept_id"`
	Name              string `json:"name"`
	IsSendFirstPerson int    `json:"is_send_first_person"` // 0为不推送，1为推送
	RobotToken        string `json:"robot_token"`
	IsRobotAttendance bool   `json:"is_robot_attendance"`
}
