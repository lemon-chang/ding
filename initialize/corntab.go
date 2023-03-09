package initialize

import (
	"ding/global"
	"github.com/robfig/cron/v3"
)

//func newWithSeconds() *cron.Cron {
//	secondParser := cron.NewParser(cron.Second | cron.Minute |
//		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
//	return cron.New(cron.WithParser(secondParser), cron.WithChain())
//}
func InitCorn() {
	var Gcontab = cron.New(cron.WithSeconds()) //精确到秒
	global.GLOAB_CORN = Gcontab
	global.GLOAB_CORN.Start()
}
