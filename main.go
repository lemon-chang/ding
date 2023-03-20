package main

import (
	"ding/dao/mysql"
	"ding/dao/redis"
	"ding/initialize"
	"ding/initialize/logger"
	"ding/routers"
	"ding/settings"
	"fmt"
	"go.uber.org/zap"
)

func main() {
	//初始化viper
	err := settings.Init()
	if err != nil {
		fmt.Printf("init settings failed ,err:%v\n", err)
		zap.L().Error(fmt.Sprintf("init settings failed ,err:%v\n", err))
		return
	}
	zap.L().Debug("viper init success...")
	//初始化Zap
	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		fmt.Printf("init logger failed ,err:%v\n", err)
		zap.L().Error(fmt.Sprintf("init logger failed ,err:%v\n", err))
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("zap init success...")
	//初始化连接飞书
	//global.InitFeishu()
	//初始化corn定时器
	initialize.InitCorn()
	//初始化链接mysql,刚好使用一下gorm，没有用到连表查询，所以比较简单
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed ,err:%v\n", err)
		zap.L().Error(fmt.Sprintf("init mysql failed ,err:%v\n", err))
		return
	}

	//err = initialize.RegisterTables(global.GLOAB_DB)
	//if err != nil {
	//	return
	//}

	//初始化连接redis
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed ,err:%v\n", err)
		zap.L().Error(fmt.Sprintf("init redis failed ,err:%v\n", err))
		return
	}
	zap.L().Debug("mysql init success...")
	//if err := snowflake.Init(settings.Conf.App.StartTime, settings.Conf.App.MachineID); err != nil {
	//	fmt.Printf("init snowflake failed,err:%v\n", err)
	//	zap.L().Error(fmt.Sprintf("init snowflake failed,err:%v\n", err))
	//	return
	//}
	//go utils.Timing(&utils.localTime)
	//初始化路由

	err = initialize.Reboot()
	if err != nil {
		fmt.Printf("重启定时任务失败,err:%v\n", err)
		zap.L().Error(fmt.Sprintf("重启定时任务失败:%v\n", err))

	} else {
		zap.L().Debug("重启定时任务成功...")
	}

	//err = initialize.AttendanceByRobot()
	//if err != nil {
	//	zap.L().Error("AttendanceByRobot init fail...")
	//	return
	//
	//}
	zap.L().Debug("AttendanceByRobot init success...")
	//err = initialize.JianBlogByRobot()
	//if err != nil {
	//	zap.L().Error("启动爬虫爬取定时任务失败", zap.Error(err))
	//	return
	//}
	r := routers.Setup(settings.Conf.Mode)
	s := fmt.Sprintf(":%d", settings.Conf.App.Port)
	r.Run(s)

}
