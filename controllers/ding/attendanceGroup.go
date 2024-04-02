package ding

import (
	response2 "ding/model/common/response"
	dingding2 "ding/model/dingding"
	"ding/model/params/ding"
	"ding/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ImportAttendanceGroupData 使用官方接口导入考勤组数据到数据库中
func ImportAttendanceGroupData(c *gin.Context) {
	token, err := (&dingding2.DingToken{}).GetAccessToken()
	err = (&dingding2.DingAttendGroup{DingToken: dingding2.DingToken{Token: token}}).ImportAttendGroups()
	if err != nil {
		response.FailWithMessage("入到考勤组数据失败", c)
		return
	}
	response.OkWithMessage("导入考勤组数据成功", c)
}

// UpdateAttendanceGroup 更新考勤组
func UpdateAttendanceGroup(c *gin.Context) {
	var p ding.ParamUpdateUpdateAttendanceGroup
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("参数错误", zap.Error(err))
		response.FailWithMessage("参数有误", c)
	}
	if p.GroupId == 0 {
		response.FailWithMessage("考勤名称或者id不能为空", c)
		return
	}
	token, err := (&dingding2.DingToken{}).GetAccessToken()
	if err != nil {
		response.FailWithMessage("钉钉token获取失败！", c)
		return
	}
	err = (&dingding2.DingAttendGroup{GroupId: p.GroupId, DingToken: dingding2.DingToken{Token: token}}).UpdateAttendGroup()
	if err != nil {
		response.FailWithMessage("更新考勤组信息失败！", c)
		return
	}
	response.OkWithMessage("更新考勤组信息成功！", c)
}

// GetAttendanceGroupListFromMysql 获取考勤组列表
func GetAttendanceGroupListFromMysql(c *gin.Context) {
	var p ding.ParamGetAttendGroup
	err := c.ShouldBindQuery(&p)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	token, err := (&dingding2.DingToken{}).GetAccessToken()
	if err != nil {
		response.FailWithMessage("钉钉token获取失败！", c)
		return
	}
	AttendanceGroupList, total, err := (&dingding2.DingAttendGroup{DingToken: dingding2.DingToken{Token: token}}).GetAttendanceGroupListFromMysql(&p)
	if err != nil {
		response.FailWithMessage("获取考勤组数据成功！", c)
		return
	}
	response.OkWithDetailed(response2.PageResult{
		List:     AttendanceGroupList,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, "获取考勤组数据成功！", c)
}
