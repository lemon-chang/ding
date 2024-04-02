package enter

import (
	"ding/initialize/cron"
	"ding/initialize/logger"
	"ding/initialize/mysql"
	"ding/initialize/outgoing"
	"ding/initialize/redis"
	"ding/initialize/validator"
	"ding/initialize/viper"
	"fmt"
	"go.uber.org/zap"
)

func Init() {
	//初始化viper
	err := viper.Init()
	if err != nil {
		zap.L().Error(fmt.Sprintf("init settings failed ,err:%v\n", err))
		return
	}
	zap.L().Debug("viper init success...")
	validator.Init()
	zap.L().Debug("validator init success...")

	//初始化Zap
	if err = logger.Init(viper.Conf.LogConfig, viper.Conf.Mode); err != nil {
		zap.L().Error(fmt.Sprintf("init logger failed ,err:%v\n", err))
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("zap init success...")

	//初始化robat的outgoing功能
	err = outgoing.Init()
	if err != nil {
		return
	}
	//初始化连接飞书
	//global.InitFeishu()
	//初始化链接mysql,刚好使用一下gorm，没有用到连表查询，所以比较简单
	if err = mysql.Init(viper.Conf.MySQLConfig); err != nil {
		zap.L().Error(fmt.Sprintf("init mysql failed ,err:%v\n", err))
		return
	}
	//初始化连接redis
	if err = redis.Init(viper.Conf.RedisConfig); err != nil {
		zap.L().Error(fmt.Sprintf("init redis failed ,err:%v\n", err))
		return
	}
	zap.L().Debug("mysql init success...")
	//初始化corn定时器
	cron.InitCorn()
}
