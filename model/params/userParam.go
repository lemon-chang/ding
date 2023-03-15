package params

// ParamSignUp 定义请求的结构体参数
type ParamSignUp struct {
	Username   string `json:"username"  binding:"required" from:"username"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
}

// ParamLogin 登录时请求参数
type ParamLogin struct {
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ParamOutGoing struct {
}

type ParamSearchUser struct {
	RobotId    string `json:"robot_id" binding:"required"`
	PersonName string `json:"person_name" binding:"required"`
}
