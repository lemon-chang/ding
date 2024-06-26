package cron

import (
	"ding/global"
	"fmt"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

//	func newWithSeconds() *cron.Cron {
//		secondParser := cron.NewParser(cron.Second | cron.Minute |
//			cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
//		return cron.New(cron.WithParser(secondParser), cron.WithChain())
//	}
func InitCorn() {
	global.GLOAB_CORN = cron.New(cron.WithSeconds()) //精确到秒
	global.GLOAB_CORN.Start()
	// ======重启定时任务=======
	if err := Reboot(); err != nil {
		zap.L().Error(fmt.Sprintf("重启定时任务失败:%v\n", err))
	} else {
		zap.L().Debug("重启定时任务成功...")
	}
	//======重启考勤=======
	//上班考勤
	if err := AttendanceByRobot(); err != nil {
		zap.L().Error("AttendanceByRobot init fail", zap.Error(err))
	} else {
		zap.L().Debug("AttendanceByRobot init success...")
	}
	//======周报考勤，采用企业日志检测=======
	if err := weeklyCheckByRobot(); err != nil {
		zap.L().Error("weeklyCheckByRobot init fail", zap.Error(err))
	} else {
		zap.L().Debug("weeklyCheckByRobot init success...")
	}
	//======发送爬取力扣的题目数=======
	err := SendLeetCode()
	if err != nil {
		zap.L().Error("SendLeetCode init fail", zap.Error(err))
	}
}
