package mysql

// CheckRobotExist 添加机器人的时候检查机器人是否已ding/dingding经存在
//func CheckRobotExistForAddRobot(r *dingding.ParamAddRobot) (err error) {
//	//先来判断机器人是否已经存在
//	robot := &dingding.DingRobot{
//		RobotId: r.RobotId,
//	}
//	result := global.GLOAB_DB.First(robot)
//	if errors.Is(result.Error, gorm.ErrRecordNotFound) { //如果之前没有注册过
//		return nil
//	} else { //之前注册过了，或者有其他的错误
//		return ErrorRobotExist
//	}
//}
//
//func GetRobotByRobotId(robot_id string) (dingding.DingRobot, error) {
//	var robot dingding.DingRobot
//	result := global.GLOAB_DB.Where("robot_id = ?", robot_id).First(&robot)
//	if result.Error != nil {
//		return dingding.DingRobot{}, result.Error
//	} else {
//		return robot, nil
//	}
//}
//func GetRobotByID(ID uint) (dingding.DingRobot, error) {
//	var robot dingding.DingRobot
//	result := global.GLOAB_DB.Where("ID = ?", ID).First(&robot)
//	if result.Error != nil {
//		return dingding.DingRobot{}, result.Error
//	} else {
//		return robot, nil
//	}
//}
//
//// GetRobots 查询自己账号所有的机器人,同时把机器人下面的电话号码也查询到了
//func GetRobots(c *gin.Context, user model.User) (robotList []*dingding.DingRobot, err error) {
//	var robots []*dingding.DingRobot
//	//result := global.GLOAB_DB.Model(&user).Joins("Robots").Find(&robots)
//	result := global.GLOAB_DB.Model(&user).Association("Robots").Find(&robots)
//	for _, robot := range robots {
//		var teles []dingding.Tele
//		err := global.GLOAB_DB.Model(&robot).Association("Teles").Find(&teles)
//		if err != nil {
//			zap.L().Error("根据机器人查询对应的电话号码错误", zap.Error(err))
//		}
//		(robot.Teles) = teles
//	}
//
//	if result == nil {
//		return robots, nil
//	} else if errors.Is(result, gorm.ErrRecordNotFound) {
//		zap.L().Warn("there is no Robot in db")
//		return nil, nil
//	} else {
//		return robots, result
//	}
//}
