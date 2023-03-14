package dingding

import (
	"ding/controllers/v2/ding"
	"ding/global"
	"fmt"

	"github.com/gin-gonic/gin"
)

func SetupDing(System *gin.RouterGroup) {
	Dept := System.Group("dept")
	{
		Dept.GET("ImportDeptData", ding.ImportDeptData)                       // 递归获取部门列表存储到数据库
		Dept.GET("getSubDepartmentListById", ding.GetSubDepartmentListByID)   // 官方接口获取子部门
		Dept.GET("getSubDepartmentListById2", ding.GetSubDepartmentListByID2) // 从数据库中一层一层的取出部门
		Dept.PUT("updateDept", ding.UpdateDept)                               // 更新部门信息，用来设置机器人token，各种开关
	}
	AttendanceGroup := System.Group("attendanceGroup")
	{
		AttendanceGroup.GET("ImportAttendanceGroupData", ding.ImportAttendanceGroupData)    //将考勤组信息导入到数据库中
		AttendanceGroup.PUT("updateAttendanceGroup", ding.UpdateAttendanceGroup)            //更新部门信息，用来设置机器人token，各种开关
		AttendanceGroup.GET("GetAttendanceGroupList", ding.GetAttendanceGroupListFromMysql) //批量获取考勤组

	}
	User := System.Group("user")
	{
		//User.POST("ImportDingUserData", ding.ImportDingUserData) //将钉钉用户导入到数据库中
		User.POST("UpdateDingUserAddr", ding.UpdateDingUserAddr) // 更新用户的博客和简书地址
		User.GET("GetAllUsers", ding.SelectAllUsers)             // 查询所有用户信息
		User.GET("GetAllJinAndBlog", ding.FindAllJinAndBlog)
		User.POST("login", ding.LoginHandler)
		User.GET("showQRCode", func(c *gin.Context) {
			username, _ := c.Get(global.CtxUserNameKey)
			c.File(fmt.Sprintf("Screenshot_%s.png", username))
		})
		User.GET("getQRCode", ding.GetQRCode) //获取群聊基本信息已经群成员id
	}
	Robot := System.Group("robot")
	{
		Robot.POST("/pingRobot", ding.PingRobot)
		Robot.POST("/addRobot", ding.AddRobot)
		Robot.DELETE("/removeRobot", ding.RemoveRobot)
		Robot.PUT("/updateRobot", ding.AddRobot) //更新机器人直接使用
		Robot.GET("getRobotDetailByRobotId", ding.GetRobotDetailByRobotId)
		Robot.GET("getRobotBaseList", ding.GetRobotBaseList)
		Robot.GET("/getRobots", ding.GetRobots)
		Robot.POST("/cronTask", ding.CronTask)      //发送定时任务
		Robot.POST("getTaskList", ding.GetTaskList) //获取定时任务列表
		Robot.POST("stopTask", ding.StopTask)       //暂停定时任务
		Robot.DELETE("removeTask", ding.RemoveTask)
		Robot.POST("reStartTask", ding.ReStartTask)
	}

}
