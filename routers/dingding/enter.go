package dingding

import (
	ding2 "ding/controllers/ding"
	"ding/global"
	"fmt"
	"github.com/gin-gonic/gin"
)

func SetupDing(System *gin.RouterGroup) {
	Dept := System.Group("/dept")
	{
		Dept.POST("/ImportDeptData", ding2.ImportDeptData)                      // 递归获取部门列表存储到数据库 (增删改)
		Dept.GET("/getDeptListDetail", ding2.GetDeptListFromMysql)              //从数据库中取出部门信息 (详细查)（Mysql）
		Dept.POST("/updateDept", ding2.UpdateDept)                              // 更新部门信息，用来设置机器人token，各种开关
		Dept.GET("/getUserByDeptid", ding2.GetUserByDeptId)                     //根据部门id查询用户信息
		Dept.GET("/getSubDepartmentListById2", ding2.GetSubDepartmentListByID2) // 从数据库中一层一层的取出部门
		Dept.GET("/getDepartmentRecursively", ding2.GetDepartmentRecursively)   // Mysql递归嵌套获取部门信息
		Dept.GET("/getSubDepartmentListById", ding2.GetSubDepartmentListByID)   // 官方接口获取子部门
	}

	AttendanceGroup := System.Group("/attendanceGroup")
	{
		AttendanceGroup.POST("/ImportAttendanceGroupData", ding2.ImportAttendanceGroupData)   //将考勤组信息同步到数据库中
		AttendanceGroup.PUT("/updateAttendanceGroup", ding2.UpdateAttendanceGroup)            //更新考勤组
		AttendanceGroup.GET("/GetAttendanceGroupList", ding2.GetAttendanceGroupListFromMysql) //批量获取考勤组
	}

	User := System.Group("/user")
	{
		User.GET("/getUserInfoDetailByToken", ding2.GetUserInfoDetailByToken)
		User.GET("/FindDingUsersInfoBase", ding2.SelectAllUsers)   // 查询所有用户基本信息
		User.GET("/FindDingUsersInfoDetail", ding2.SelectAllUsers) //查询用户详细信息

		User.GET("/getTasks", ding2.GetTasks) //获取所有定时任务，包括暂停的任务

		User.POST("/getWeekConsecutiveSignNum", ding2.GetWeekConsecutiveSignNum) //获取用户周连续签到次数
		User.POST("/getWeekSignNum", ding2.GetWeekSignNum)                       //根据第几星期获取用户签到次数（使用redis的bitCount函数）
		User.POST("/getWeekSignDetail", ding2.GetWeekSignDetail)                 //获取用户某个星期签到情况，默认是当前所处的星期，构建成为一个有序的HashMap
		User.GET("/getDeptIdByUserId", ding2.GetDeptByUserId)                    //通过userid查询部门id
		User.POST("/makeupSign", ding2.MakeupSign)                               //为用户补签到并返回用户联系签到次数
		User.GET("/showQRCode", func(c *gin.Context) {
			username, _ := c.Get(global.CtxUserNameKey)
			c.File(fmt.Sprintf("Screenshot_%s.png", username))
		})
		User.GET("/getQRCode", ding2.GetQRCode) //获取群聊基本信息已经群成员id
	}
	Robot := System.Group("robot")
	{
		Robot.POST("/pingRobot", ding2.PingRobot)
		Robot.POST("/addRobot", ding2.AddRobot)            // 添加机器人(增)
		Robot.PUT("/updateRobot", ding2.AddRobot)          //更新机器人直接使用（改）
		Robot.DELETE("/removeRobot", ding2.RemoveRobot)    // 可以批量 （删除）
		Robot.GET("/getSharedRobot", ding2.GetSharedRobot) // （查）
		Robot.GET("/getRobotList", ding2.GetRobotList)     //获取所有机器人（查）

		Robot.POST("/cronTask", ding2.CronTask)                 //发送定时任务（增）
		Robot.DELETE("/removeTask", ding2.RemoveTask)           //移除定时任务 （删）
		Robot.POST("/stopTask", ding2.StopTask)                 //暂停定时任务 (改)
		Robot.PUT("/editTaskContent", ding2.EditTaskContent)    //编辑定时任务的内容 (改)
		Robot.POST("/reStartTask", ding2.ReStartTask)           //重启定时任务 （改）
		Robot.GET("/getTaskDetail", ding2.GetTaskDetail)        //获取定时任务详情 （详细查）
		Robot.POST("/getRobotTaskList", ding2.GetRobotTaskList) //加载机器人定时任务（查）
		Robot.POST("/getUserTaskList", ding2.GetUserTaskList)   //加载机器人定时任务（查）
	}
	CronTask := System.Group("cronTask")
	{
		CronTask.POST("/cronTask", ding2.CronTask)                 //发送定时任务（增）
		CronTask.DELETE("/removeTask", ding2.RemoveTask)           //移除定时任务 （删）
		CronTask.POST("/stopTask", ding2.StopTask)                 //暂停定时任务 (改)
		CronTask.PUT("/editTaskContent", ding2.EditTaskContent)    //编辑定时任务的内容 (改)
		CronTask.POST("/reStartTask", ding2.ReStartTask)           //重启定时任务 （改）
		CronTask.GET("/getTaskDetail", ding2.GetTaskDetail)        //获取定时任务详情 （详细查）
		CronTask.POST("/getRobotTaskList", ding2.GetRobotTaskList) //加载机器人定时任务（查）
		CronTask.POST("/getUserTaskList", ding2.GetUserTaskList)   //加载机器人定时任务（查）
	}
	//机器人问答模块
	QuAndAn := System.Group("/quAndAn")
	{
		QuAndAn.POST("/updateData", ding2.UpdateData)   //上传资源
		QuAndAn.DELETE("/deleteData", ding2.DeleteData) //删除资源
		QuAndAn.PUT("/putData", ding2.PutData)          //修改资源
		QuAndAn.POST("/getData", ding2.GetData)         //查询资源
	}
	LeaveGroup := System.Group("/response")
	{
		LeaveGroup.POST("/SubscribeToSomeone", ding2.SubscribeToSomeone) //订阅某人考勤情况
		LeaveGroup.DELETE("/Unsubscribe", ding2.Unsubscribe)             //取消订阅
	}
}
