package ding

type ParamDeptWeekPaperCheck struct {
	DeptId      int    `json:"dept_id"  form:"dept_id"` //部门id
	Name        string `json:"name"`                    //部门名称
	IsWeekPaper int    `json:"is_week_paper"`           //部门是否参与周报检测
}

type ParamUserWeekPaperCheck struct {
	UserId      string ` json:"user_id" form:"user_id"`            //用户id
	Name        string `json:"name"`                               //用户名称
	IsWeekPaper int    `json:"is_week_paper" form:"is_week_paper"` //用户是否参与周报检测
}
