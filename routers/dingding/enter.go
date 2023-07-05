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
		Dept.GET("getDeptListFromMysql", ding.GetDeptListFromMysql)
		Dept.PUT("updateDept", ding.UpdateDept)     // 更新部门信息，用来设置机器人token，各种开关
		Dept.PUT("updateSchool", ding.UpdateSchool) //更新部门是否在校信息
	}

	AttendanceGroup := System.Group("attendanceGroup")
	{
		AttendanceGroup.GET("ImportAttendanceGroupData", ding.ImportAttendanceGroupData) //将考勤组信息导入到数据库中
		AttendanceGroup.PUT("updateAttendanceGroup", ding.UpdateAttendanceGroup)         //考勤组开关

		AttendanceGroup.GET("GetAttendanceGroupList", ding.GetAttendanceGroupListFromMysql) //批量获取考勤组
	}
	LeaveGroup := System.Group("leave")
	{
		LeaveGroup.POST("SubscribeToSomeone", ding.SubscribeToSomeone) //订阅某人考勤情况
		LeaveGroup.DELETE("Unsubscribe", ding.Unsubscribe)             //取消订阅
	}
	User := System.Group("user")
	{
		User.GET("getUserInfo", ding.GetUserInfo)
		//User.POST("ImportDingUserData", ding.ImportDingUserData) //将钉钉用户导入到数据库中
		User.POST("UpdateDingUserAddr", ding.UpdateDingUserAddr) // 更新用户的博客和简书地址
		User.GET("GetAllUsers", ding.SelectAllUsers)             // 查询所有用户信息
		User.GET("GetAllJinAndBlog", ding.FindAllJinAndBlog)

		User.GET("showQRCode", func(c *gin.Context) {
			username, _ := c.Get(global.CtxUserNameKey)
			c.File(fmt.Sprintf("Screenshot_%s.png", username))
		})
		//User.GET("getQRCode", ding.GetQRCode)             //获取群聊基本信息已经群成员id
		User.GET("/getActiveTask", ding.GetAllActiveTask) //查看所有的活跃任务,也就是手动更新，后续可以加入casbin，然后就是管理员权限
	}
	Robot := System.Group("robot")
	{

		Robot.POST("/pingRobot", ding.PingRobot)
		Robot.POST("/addRobot", ding.AddRobot)
		Robot.DELETE("/removeRobot", ding.RemoveRobot)
		Robot.PUT("/updateRobot", ding.AddRobot) //更新机器人直接使用
		Robot.GET("getSharedRobot", ding.GetSharedRobot)
		Robot.GET("getRobotDetailByRobotId", ding.GetRobotDetailByRobotId)
		//Robot.GET("getRobotBaseList", ding.GetRobotBaseList)
		Robot.GET("/getRobotBaseList", ding.GetRobots)          //获取所有及重庆人
		Robot.POST("/cronTask", ding.CronTask)                  //发送定时任务
		Robot.POST("getTaskList", ding.GetTaskList)             //加载定时任务
		Robot.POST("stopTask", ding.StopTask)                   //暂停定时任务
		Robot.DELETE("removeTask", ding.RemoveTask)             //移除定时任务
		Robot.POST("reStartTask", ding.ReStartTask)             //重启定时任务
		Robot.PUT("editTaskContent", ding.EditTaskContent)      //编辑定时任务的内容
		Robot.GET("/getTaskDetail", ding.GetTaskDetail)         //获取定时任务详情
		Robot.GET("/getAllPublicRobot", ding.GetAllPublicRobot) //获取所有的公共机器人

		Robot.POST("singleChat", ding.SingleChat)
	}
}
