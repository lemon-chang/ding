package ding

import (
	dingding2 "ding/model/dingding"
	"ding/model/params/ding"
	"ding/response"
	"github.com/gin-gonic/gin"
)

// 更新部门周报检测状态
func UpdateDeptWeekCheckStatus(c *gin.Context) {
	var p ding.ParamDeptWeekPaperCheck
	err := c.ShouldBindJSON(&p)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dep := dingding2.DingDept{DeptId: p.DeptId, IsStudyWeekPaper: p.IsStudyWeekPaper}
	err = dep.UpdateDeptWeekCheckStatus()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// 获取全部部门需要考勤状态
func GetDeptWeekCheckStatus(c *gin.Context) {
	var pDept dingding2.DingDept
	depts, err := pDept.GetDeptWeekCheckStatus()
	if err != nil {
		response.FailWithMessage("获取部门信息失败", c)
		return
	}
	ps := []ding.ParamDeptWeekPaperCheck{}
	for _, dept := range depts {
		p := ding.ParamDeptWeekPaperCheck{
			DeptId:           dept.DeptId,
			Name:             dept.Name,
			IsStudyWeekPaper: dept.IsStudyWeekPaper,
		}
		ps = append(ps, p)
	}
	response.OkWithDetailed(ps, "获取部门周报检测状态成功！", c)
}

// 更新成员周报检测
func UpdateUserWeekCheckStatus(c *gin.Context) {
	var p ding.ParamUserWeekPaperCheck
	err := c.ShouldBindJSON(&p)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	user := dingding2.DingUser{UserId: p.UserId, IsStudyWeekPaper: p.IsWeekPaper}
	err = user.UpdateUserWeekCheckStatus()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("用户周报状态更新成功", c)
}

// 获取部门中成员需要周报检测的状态
func GetUserWeekCheckStatus(c *gin.Context) {
	var pDept ding.ParamDeptWeekPaperCheck
	err := c.ShouldBindQuery(&pDept)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	paramUserWeekPaperChecks := make([]ding.ParamUserWeekPaperCheck, 0)
	var pUser dingding2.DingUser
	users, err := pUser.GetWeekPaperUsersStatusByDeptId(pDept.DeptId)
	for _, user := range users {
		var p ding.ParamUserWeekPaperCheck
		p.Name = user.Name
		p.UserId = user.UserId
		p.IsWeekPaper = user.IsStudyWeekPaper
		paramUserWeekPaperChecks = append(paramUserWeekPaperChecks, p)
	}
	if err != nil {
		response.FailWithMessage("获取部门用户周报检测状态失败", c)
		return
	}
	response.OkWithDetailed(paramUserWeekPaperChecks, "获取部门用户周报检测状态成功！", c)
}
