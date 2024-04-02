package ding

import (
	"ding/model/chatRobot"
	"ding/model/dingding"
	"ding/response"
	"github.com/gin-gonic/gin"
)

func CreateStreamMsg(c *gin.Context) {
	r := chatRobot.NewRobotStream()
	err := c.ShouldBindJSON(&r)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	user := dingding.DingUser{UserId: r.UserId}
	dept, _ := user.GetUserDeptIdByUserId()
	r.DeptId = dept.DeptId
	err = r.CreateRobotStream()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

//查询信息

//修改信息

//删除信息
