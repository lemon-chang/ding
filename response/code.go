package response

type ResCode int64

const (
	CodeSuccess      ResCode = 200
	CodeInvalidParam         = 400
	CodeNeedLogin            = 401
	CodeUserExist            = 403 + iota
	CodeLoginEror
	CodeRobotNameNotNull
	CodeRobotIdOrSecretInvalid
	CodeUserNotExist
	CodeInvalidPassword
	CodeServerBusy

	CodeInvalidToken
	CodeRobotExist
	CodeRobotNotExist
	CodeTeleOrPersonNameExist
	CodeNotHasRobot
	CodeSpecInvalid
	CodeNotRemoveTask
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:                "success",
	CodeInvalidParam:           "请求参数错误",
	CodeUserExist:              "用户已经存在",
	CodeUserNotExist:           "用户不存在",
	CodeLoginEror:              "登录有误，请重新登录",
	CodeRobotNameNotNull:       "机器人姓名不能为空",
	CodeRobotIdOrSecretInvalid: "机器人id或者密码不合法",
	CodeInvalidPassword:        "密码或者用户名错误",
	CodeServerBusy:             "系统繁忙",
	CodeNeedLogin:              "需要登录",
	CodeInvalidToken:           "无效的token",
	CodeRobotExist:             "机器人id已经存在或者您的账户机器人昵称重复",
	CodeRobotNotExist:          "机器人不存在",
	CodeTeleOrPersonNameExist:  "电话号码或者人名已经存在",
	CodeNotHasRobot:            "该用户未拥有该机器人",
	CodeSpecInvalid:            "定时规则可能不合法，注意使用英文",
	CodeNotRemoveTask:          "未能成功移除该定时任务",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
