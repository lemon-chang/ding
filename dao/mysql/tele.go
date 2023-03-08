package mysql

// CheckTeleExist 检查此电话号码是否已经添加过了
//func CheckTeleExist(p *params.ParamAddTele, robot *dingding.DingRobot) (IsExist bool, err error) {
//	//检查此电话号码要检查两点，一点是电话号码是否出现，二是出现之后是否属于当前机器人，如果不属于当前机器人的话，即是出现了，也是可以添加的
//	teles := []dingding.Tele{}
//	global.GLOAB_DB.Model(robot).Association("Teles").Find(&teles) //查到该机器人所关联的所有机器人
//	for _, tele := range teles {
//		if p.PersonName == tele.Personname || p.Number == tele.Number { //说明此时已经存在了
//			return true, ErrorTeleOrPersonNameExist
//		}
//	}
//
//	return false, nil
//
//}
//
//func InsertTele(tele *dingding.Tele) (err error) {
//	result := global.GLOAB_DB.Create(&tele)
//	if result.Error != nil {
//		return result.Error
//	} else {
//		fmt.Println(result.RowsAffected)
//		return
//	}
//}
//func InsertTeleBatch(ID string, teles []dingding.Tele) (err error) {
//	//批量插入之前应该先把之前的全部删除掉
//	r := dingding.DingRobot{
//		RobotId: ID,
//	}
//	oldTele := []dingding.Tele{}
//	err = global.GLOAB_DB.Model(&r).Association("Teles").Find(&oldTele)
//	if err != nil {
//
//	}
//
//	err = global.GLOAB_DB.Delete(&oldTele).Error
//	if err != nil {
//
//	}
//	oldTele = []dingding.Tele{}
//	err = global.GLOAB_DB.Create(&teles).Error
//
//	if err != nil {
//		return
//	}
//	return
//}
//func UpdateTele(c *gin.Context, p *params.ParamUpdateTele) (err error) {
//	//先通过ID查出来这个电话号码
//	oldTele := dingding.Tele{
//		Model: gorm.Model{
//			ID: p.Id,
//		},
//	}
//	err = global.GLOAB_DB.First(&oldTele).Error
//	if err != nil {
//		zap.L().Error("通过电话号码的id获取到电话号码失败", zap.Error(err))
//		return
//	}
//
//	err = global.GLOAB_DB.Model(&oldTele).Updates(dingding.Tele{Number: p.NewNumber, Personname: p.NewPersonName}).Error
//	if err != nil {
//		return err
//	}
//	username, _ := c.Get(global.CtxUserNameKey)
//
//	if username != "" {
//		zap.L().Error("通过上下文获取用户名失败")
//	}
//	if err != nil {
//		zap.L().Error(fmt.Sprintf("更新电话号码失败，对应用户：%s,对应机器人：%s,所更新姓名%s，所更新电话号码%s", username, oldTele.Robots[0].RobotName, oldTele.Personname, oldTele.Number))
//		return
//	}
//	return
//}
//func RemoveTele(c *gin.Context, p *params.ParamRemoveTele) (err error) {
//	tele := dingding.Tele{
//		Model: gorm.Model{
//			ID: p.Id,
//		},
//	}
//	err = global.GLOAB_DB.Delete(&tele).Error
//	if err != nil {
//		zap.L().Error("通过电话号码id删除电话号码失败", zap.Error(err))
//		return
//	}
//	return
//}
//
////GetTeles 得到该机器人的所有电话号码
//func GetTelesByRobotId(robot *dingding.DingRobot) (teles []dingding.Tele, err error) {
//	result := global.GLOAB_DB.Debug().Model(&robot).Association("Teles").Find(&teles) //查找所有匹配的关联记录
//	//GLOAB_DB.Where("name = ?",robotName).Find(&teles)
//	if result == nil {
//		return teles, nil
//	} else {
//		return nil, result
//	}
//}
//func SearchTele(p *params.ParamSearchUser) (teles []dingding.Tele, err error) {
//	PersonName := "%" + p.PersonName + "%"
//
//	robot, err := GetRobotByRobotId(p.RobotId)
//	err = global.GLOAB_DB.Model(robot).Association("Teles").Find(&teles, "personname like ?", PersonName)
//	return
//}
//func PostUserDetailByUserId(userId string, tele *dingding.Tele) (err error) {
//	err = global.GLOAB_DB.Model(&dingding.Tele{}).Where("user_id = ?", userId).Updates(&tele).Error
//	if err != nil {
//		return
//		zap.L().Error(fmt.Sprintf("根据UserId查询用户信息后更新数据库失败，userId = %s", userId), zap.Error(err))
//	}
//	return
//}
