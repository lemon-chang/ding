package leave

// DingLeaveStatus 请假状态
type DingLeaveStatus struct {
	StartTime       int64  `json:"start_time"`
	DurationPercent int    `json:"duration_percent"`
	EndTime         int64  `json:"end_time"`
	LeaveCode       string `json:"leave_code"` // 请假类型 个人事假：d4edf257-e581-45f9-b9b9-35755b598952  非个人事假：baf811bc-3daa-4988-9604-d68ec1edaf50  病假：a7ffa2e6-872a-498d-aca7-4554c56fbb52
	DurationUnit    string `json:"duration_unit"`
	UserID          string `json:"userid"`
}

// DingLeaveResult 请假列表
type DingLeaveResult struct {
	LeaveStatus *[]DingLeaveStatus `json:"leave_status"`
	HasMore     bool               `json:"has_more"` // 是否有更多数据
}

// DingResponse 钉钉响应
type DingResponse struct {
	ErrCode   int             `json:"errcode"`
	Result    DingLeaveResult `json:"result"`
	Success   bool            `json:"success"`
	ErrMsg    string          `json:"errmsg"`
	RequestID string          `json:"request_id"`
}

type DingUser struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Type         map[string]int
	AverageLeave float64 // 平均请假次数
}
