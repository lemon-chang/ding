package redis

import (
	"context"
	"ding/global"
	"go.uber.org/zap"

	"time"
)

const (
	Prefix   = "dingding:"
	KeyUser  = "user:"
	KeyRobot = "robot:"
	KeyTele  = "tele:"
)

//穿入key,value，过期时间
func SetRedisKey(key string, value string, expirationSecond int) error {
	err := global.GLOBAL_REDIS.Set(context.Background(), key, value, GetSeondByNs(expirationSecond)).Err()
	if err != nil {
		zap.L().Error("key设置失败;"+"key: "+key+" value:"+value+" ", zap.Error(err))
	}
	return err
}
func GetRobotNumberInfoKey(username, roborname, personname string) string {
	//dingding:robot:考勤机器人
	return Prefix + KeyTele + roborname + ":" + personname
}
func GetSeondByNs(second int) time.Duration {
	return time.Duration(second * 1e9)
}
func GetRedisKey(key string) string {
	return Prefix + key
}
func GetRedisKeys(Prikey string) (keys []string, err error) {
	keys, err = global.GLOBAL_REDIS.Keys(context.Background(), Prikey+"*").Result()
	if err != nil {
		return nil, err
	}
	return keys, err
}
