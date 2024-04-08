package dingding

type ChatUpdateTitle struct {
	TimeStamp          int64  `json:"timeStamp"`
	EventID            string `json:"eventId"`
	ChatID             string `json:"chatId"`
	OperatorUnionID    string `json:"operatorUnionId"`
	Title              string `json:"title"`
	OpenConversationID string `json:"openConversationId"`
	Operator           string `json:"operator"` // 操作人userid
}
type UserAddOrg struct {
	TimeStamp  string   `json:"timeStamp"`
	EventID    string   `json:"eventId"`
	OptStaffID string   `json:"optStaffId"`
	UserID     []string `json:"userId"`
}
type UserLeaveOrg struct {
	TimeStamp  string   `json:"timeStamp"`
	EventID    string   `json:"eventId"`
	OptStaffID string   `json:"optStaffId"`
	UserID     []string `json:"userId"`
}
type OrgDeptCreate struct {
	DeptId []int `json:"deptId"`
}
type OrgDeptModify struct {
	DeptId []int `json:"deptId"`
}
type OrgDeptRemove struct {
	DeptId []int `json:"deptId"`
}

type UserModifyOrg struct {
	TimeStamp string `json:"timeStamp"`
	DiffInfo  []struct {
		Prev struct {
			ExtFields string `json:"extFields"`
			Name      string `json:"name"`
			WorkPlace string `json:"workPlace"`
		} `json:"prev"`
		Curr struct {
			ExtFields string `json:"extFields"`
			Name      string `json:"name"`
			WorkPlace string `json:"workPlace"`
		} `json:"curr"`
		Userid string `json:"userid"`
	} `json:"diffInfo"`
	EventID string   `json:"eventId"`
	UserID  []string `json:"userId"`
}
type AttendGroupChange struct {
	EventID string `json:"eventId"`
	Corpid  string `json:"corpid"`
	Name    string `json:"name"`
	Action  string `json:"action"`
	ID      int    `json:"id"` // 考勤组id
}
