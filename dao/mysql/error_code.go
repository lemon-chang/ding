package mysql

import "github.com/pkg/errors"

var (
	ErrorUserExist             = errors.New("用户已经存在")
	ErrorUserNotExist          = errors.New("用户不存在")
	ErrorInvalidPassword       = errors.New("用户密码错误")
	ErrorInvalidID             = errors.New("无效的ID")
	ErrorRobotExist            = errors.New("该机器人ID已经存在或者在您的账户机器人昵称重复")
	ErrorRobotNotExist         = errors.New("在当前用户下机器人不存在,无法删除")
	ErrorTeleOrPersonNameExist = errors.New("在该机器人中此电话号码或者姓名已经存在")
	ErrorNotHasRobot           = errors.New("该用户未拥有该机器人")
	ErrorNotHasTask            = errors.New("未拥有该任务")
	ErrorSpecInvalid           = errors.New("定时规则可能不合法")
)
