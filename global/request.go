package global

import (
	"errors"
	"github.com/gin-gonic/gin"
)

var ErrorUserNotLogin = errors.New("用户登录状态有误，请重新登录")
var ErrorRobotNotLogin = errors.New("机器人未登录")
var ErrorCornTabNotGet = errors.New("定时任务获取失败")

const CtxUserIDKey = "user_id"
const CtxIDKey = "id"
const CtxUserNameKey = "userName"
const CtxUserAuthorityIDKey = "authority_id"

const CtxRobotIDKey = "robotID"
const CtxCornTab = "task"

// GetCurrentUser 获取当前登录用户的ID
func GetCurrentUserId(c *gin.Context) (UserID string) {
	uid, _ := c.Get(CtxUserIDKey)
	UserID = uid.(string) // 进行类型断言
	return UserID
}
