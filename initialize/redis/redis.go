package redis

import (
	"context"
	"ding/global"
	"ding/initialize/viper"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	KeyDeptAveLeave = "leave:" // 根据各部门平均请假次数排序的集合
	KeyDeptAveLate  = "late:"  // 根据各部门平均迟到次数排序的集合
	Prefix          = "ding:"
	ActiveTask      = "activeTask:" //活跃任务部分
	Attendance      = "attendance:" //考勤状态部分
	User            = "user:"
	UserSign        = "sign:"
	LeetCode        = "leetcode:"
)

func Init(redisCfg *viper.RedisConfig) (err error) {

	client := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})

	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		zap.L().Error("redis connect ping failed , err :", zap.Error(err))
		return
	} else {
		zap.L().Info("redis connect ping response:", zap.String("pong", pong))
		global.GLOBAL_REDIS = client
		fmt.Println("redis连接成功")
	}
	return
}
