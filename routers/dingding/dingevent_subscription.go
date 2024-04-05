package dingding

import (
	"ding/model/dingding"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupDingEventSubscription(DingEvent *gin.RouterGroup) {
	DingEvent.GET("chat_update_title", func(context *gin.Context) {
		fmt.Println("群名更改成功")
	})
	DingEvent.POST("user_add_org", func(c *gin.Context) {
		var UserAddOrg dingding.UserAddOrg
		err := c.ShouldBindJSON(&UserAddOrg)
		if err != nil {
			zap.L().Error("参数有误", zap.Error(err))
			return
		}
		user := &dingding.DingUser{UserId: UserAddOrg.UserID[0]}
		err = user.Add()
		if err != nil {
			zap.L().Error("添加用户有误", zap.Error(err))
			return
		}
	})
	DingEvent.POST("user_leave_org", func(c *gin.Context) {
		var UserLeaveOrg dingding.UserLeaveOrg
		err := c.ShouldBindJSON(&UserLeaveOrg)
		if err != nil {
			zap.L().Error("参数有误", zap.Error(err))
			return
		}
		user := &dingding.DingUser{UserId: UserLeaveOrg.UserID[0]}
		err = user.Delete()
		if err != nil {
			zap.L().Error("删除用户有误", zap.Error(err))
			return
		}
	})
	DingEvent.POST("user_modify_org", func(c *gin.Context) {
		var UserLeaveOrg dingding.UserModifyOrg
		err := c.ShouldBindJSON(&UserLeaveOrg)
		if err != nil {
			zap.L().Error("参数有误", zap.Error(err))
			return
		}
		user := &dingding.DingUser{UserId: UserLeaveOrg.UserID[0]}
		err = user.GetUserDetailByUserId()
		if err != nil {
			zap.L().Error("获取用户信息失败", zap.Error(err))
			return
		}
		err = user.UpdateByDingEvent()
		if err != nil {
			zap.L().Error("更新用户有误", zap.Error(err))
			return
		}
	})
	// https://open.dingtalk.com/document/orgapp/address-book-enterprise-department-create-stream
	DingEvent.POST("org_dept_create", func(c *gin.Context) {
		var OrgDeptCreate dingding.OrgDeptCreate
		err := c.ShouldBindJSON(&OrgDeptCreate)
		if err != nil {
			zap.L().Error("参数有误", zap.Error(err))
			return
		}
		dept := &dingding.DingDept{DeptId: OrgDeptCreate.DeptId[0]}
		err = dept.Insert()
		if err != nil {
			zap.L().Error("部门插入有误", zap.Error(err))
			return
		}
	})
	DingEvent.POST("org_dept_modify", func(c *gin.Context) {
		var OrgDeptModify dingding.OrgDeptModify
		err := c.ShouldBindJSON(&OrgDeptModify)
		if err != nil {
			zap.L().Error("参数有误", zap.Error(err))
			return
		}
		dept := &dingding.DingDept{DeptId: OrgDeptModify.DeptId[0]}
		err = dept.UpdateByDingEvent()
		if err != nil {
			zap.L().Error("部门插入有误", zap.Error(err))
			return
		}
	})
	DingEvent.POST("org_dept_remove", func(c *gin.Context) {
		var OrgDeptRemove dingding.OrgDeptRemove
		err := c.ShouldBindJSON(&OrgDeptRemove)
		if err != nil {
			zap.L().Error("参数有误", zap.Error(err))
			return
		}
		dept := &dingding.DingDept{DeptId: OrgDeptRemove.DeptId[0]}
		err = dept.Delete()
		if err != nil {
			zap.L().Error("部门删除有误", zap.Error(err))
			return
		}
	})
	// https://open.dingtalk.com/document/orgapp/attendance-group-change-stream
	DingEvent.POST("attend_group_change", func(c *gin.Context) {
		var p dingding.AttendGroupChange
		err := c.ShouldBindJSON(&p)
		if err != nil {
			zap.L().Error("参数有误", zap.Error(err))
			return
		}
		attengroup := &dingding.DingAttendGroup{GroupId: p.ID}
		if p.Action == "attend_group_create" {
			err = attengroup.Insert()
		} else if p.Action == "attend_group_delete" {
			err = attengroup.Delete()
		} else if p.Action == "attend_group_update" {
			err = attengroup.UpdateByDingEvent()
		}
		if err != nil {
			zap.L().Error("考勤组变更有误", zap.Error(err))
		}
	})
}
