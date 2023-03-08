package redis

//import (
//	"context"
//	"ding/model/dingding"
//	"errors"
//	"gorm.io/gorm"
//
//	//"ding/dao/mysql"
//	"ding/global"
//	"ding/model"
//	"encoding/json"
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"go.uber.org/zap"
//)
//
//type redis_mysql struct {
//}
//
//func AddRobotToRedis(robot *dingding.Robot, c *gin.Context) (err error) {
//	redisValue, err := json.Marshal(&robot)
//	if err != nil {
//		return err
//	}
//	uid, err := global.GetCurrentUserID(c) //通过token在gin框架里面设置上下文取出来当前用户登录的id
//	//var mysql Mysql
//	//var redis_mysql *redis_mysql
//	//mysql = redis_mysql
//	//user, err := mysql.GetUserByID(uid)
//	user, err := GetUserByID(uid)
//	//我们需要获取到前缀
//	err = SetRedisKey_Set(GetRobotInfoKey(user.Username, robot.RobotName), string(redisValue))
//	return err
//}
//
////go-redis对Set集合进行操作，传入 key ,value
//func SetRedisKey_Set(key string, value string) (err error) {
//	ret, err := global.GLOBAL_REDIS.SAdd(context.Background(), key, value).Result()
//	fmt.Printf("插入redis数据库%i条数据", ret)
//	if err != nil {
//		zap.L().Error("key设置失败;"+"key: "+key+" value:"+value+" ", zap.Error(err))
//	}
//	return err
//}
//func GetRobotInfoKey(username, robotname string) string {
//	return Prefix + KeyRobot + robotname
//}
//
//func GetUserByID(id int64) (model.User, error) {
//	//这个地方是有一个需要注意的地方的，当我们通过主键检索单个对象的时候，我们使用内联条件，不使用where
//	var user model.User
//	resultDB := global.GLOAB_DB.Table("users").Where("user_id = ?", id).First(&user)
//
//	if errors.Is(resultDB.Error, gorm.ErrRecordNotFound) {
//		return model.User{}, resultDB.Error
//	}
//	return user, resultDB.Error
//}
