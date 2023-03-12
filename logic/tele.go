package logic

//func AddTele(c *gin.Context, p *params.ParamAddTele) (err error) {
//	//检查该用户是否拥有该机器人,如果该添加号码的机器人不属于用户，那就直接返回即可
//	robot, err := (&dingding2.DingRobot{RobotId: p.RobotId}).GetRobotByRobotId()
//	if err != nil {
//		return
//	}
//	//再判断电话号码或者当前用户在当前的机器人中是否依据存过了
//	isExist, err := mysql.CheckTeleExist(p, robot)
//	if isExist || err != nil {
//		zap.L().Error("此号码或者此用户名已经存在")
//		return mysql.ErrorTeleOrPersonNameExist
//	}
//	if err != nil {
//		return err
//	}
//	//电话号码可以添加到数据库了
//	//构造一个电话号码实例
//	tele := &dingding2.Tele{
//		Number:     p.Number,
//		Personname: p.PersonName,
//		Robots: []dingding2.DingRobot{
//			*robot,
//		},
//	}
//	err = mysql.InsertTele(tele)
//	return err
//}
//func BatchInsertGroupMembers(c *gin.Context, p *params.ParamBatchInsertGroupMembers) (err error) {
//	//检查该用户是否拥有该机器人,如果该添加号码的机器人不属于用户，那就直接返回即可
//	//robot, isOwn := mysql.CheckIfYouOwnTheRobot(c, p)
//	//if isOwn == false {
//	//	zap.L().Error("该用户未拥有该机器人")
//	//	return mysql.ErrorNotHasRobot
//	//}
//	robot, err := mysql.GetRobotByRobotId(p.RobotID)
//	if err != nil {
//		zap.L().Error("根据id获取机器人失败", zap.Error(err))
//		return err
//	}
//	//获取机器人所在群里的chatId,传递参数CropId
//	//具体的插入逻辑，天骄汇总
//	CropId, err := redis.GetCropId()
//	if err != nil {
//		return
//	}
//
//	chatId, title, err := v2.GetChatIdAndTitle(c, CropId)
//	if err != nil || chatId == "" || title == "" {
//		return
//		zap.L().Error("获取chatId失败", zap.Error(err))
//	}
//	robot.ChatId = chatId
//	robot.Title = title
//	//获取到ChatId，应该把ChatId存储到该机器人对应的表中
//
//	//获取access_token
//
//	access_token, err := (&dingding2.DingToken{}).GetAccessToken()
//	if err != nil || access_token == "" {
//		return
//	}
//	//获取openConverStaionId
//	openConversationId, err := dingding.GetOpenConverstaionId(access_token, chatId)
//	if err != nil || openConversationId == "" {
//		return
//	}
//	robot.OpenConversationID = openConversationId
//	//把openConversationId插入到数据库中
//
//	PUpdateRobot := dingding2.ParamUpdateRobot{
//
//		OpenConversationID: openConversationId,
//		Title:              title,
//	}
//	err = mysql.UpdateRobot(c, &PUpdateRobot)
//	if err != nil {
//		return
//	}
//	//根据token和openConverstaionID获取到userIds
//	userIds, err := dingding.GetUserIds(access_token, openConversationId)
//	if err != nil || userIds == nil {
//		zap.L().Error("通过access_token和openConversationId获取userId失败")
//		return
//	}
//
//	p.UserIds = userIds
//
//	//userId可以添加到数据库了
//	//批量插入，gorm通过切片实现
//	teles := []dingding2.Tele{}
//	for _, userId := range p.UserIds {
//		tele := dingding2.Tele{
//			UserId: userId,
//			Robots: []dingding2.DingRobot{
//				robot,
//			},
//		}
//		teles = append(teles, tele)
//	}
//
//	//批量插入的时候，应该传入机器人对应的ID
//	err = mysql.InsertTeleBatch(p.RobotID, teles)
//	if err != nil {
//		zap.L().Error("用户扫码后批量插入userId失败", zap.Error(err))
//	}
//	//根据userId查询到该群成员的所有信息
//	//用golang封装一个Post请求
//	for _, userId := range p.UserIds {
//		tele, err := dingding.PostGetUserDetail(access_token, userId)
//		if err != nil {
//			zap.L().Error(fmt.Sprintf("通过UserId查询用户详细信息失败，userId = %s", userId), zap.Error(err))
//		}
//		err = mysql.PostUserDetailByUserId(userId, &tele)
//		if err != nil {
//			zap.L().Error(fmt.Sprintf("根据UserId查询用户信息后更新数据库失败，userId = %s", userId), zap.Error(err))
//		}
//
//	}
//	return err
//}
//
//func UpdateTele(c *gin.Context, p *params.ParamUpdateTele) (err error) {
//	//更新的时候，用户肯定会有这个机器人,也会有这个用户
//	return mysql.UpdateTele(c, p)
//}
//func RemoveTele(c *gin.Context, p *params.ParamRemoveTele) (err error) {
//	return mysql.RemoveTele(c, p)
//}
//func AddTeleToRedis(c *gin.Context, p *params.ParamAddTele) (err error) {
//	redisValue, err := json.Marshal(&p)
//	if err != nil {
//		return
//	}
//	zap.L().Info(string(redisValue))
//	//检查该用户是否拥有该机器人,如果该添加号码的机器人不属于用户，那就直接返回即可
//	userID, err := global.GetCurrentUserID(c)
//	user, err := mysql.GetUserByID(userID)
//	robot, isOwn := mysql.CheckIfYouOwnTheRobot(c, p)
//	if isOwn == false {
//		zap.L().Error("该用户未拥有该机器人")
//		return mysql.ErrorNotHasRobot
//	}
//	//再判断电话号码在当前的机器人中是否依据存过了，可以使用redis里面的集合，redis中的集合，一个键可以对应多个值，且集合中的元素不可重复
//	//集合里面应该嵌套hash,hash适合存储对象类型的数据
//	//dingding:user:当前用户名:robotname:
//	err = redis.SetRedisKey(redis.GetRobotNumberInfoKey(user.Username, robot.RobotName, p.PersonName), string(redisValue), 0)
//	return err
//}
//
//// GetTeles得到所有的电话号码
//func GetTeles(c *gin.Context, p *params.ParamGetTeles) (data []dingding2.Tele, err error) {
//	//查看此机器人的姓名是否正确
//	robot, exist := mysql.CheckIfYouOwnTheRobot(c, p)
//	if exist == false {
//		return nil, mysql.ErrorNotHasRobot
//	}
//	return mysql.GetTelesByRobotId(robot)
//}
//func GetTelesFromRedis(c *gin.Context, p *params.ParamGetTeles, user model.User) (data []dingding2.RedisInfo, err error) {
//	//查看此机器人的姓名是否正确
//	_, exist := mysql.CheckIfYouOwnTheRobot(c, p)
//	if exist == false {
//		return nil, mysql.ErrorNotHasRobot
//	}
//
//	Prekey := redis.GetRedisKey("user:" + user.Username + ":" + "robot:" + p.RobotId)
//	keys, err := redis.GetRedisKeys(Prekey)
//	data = make([]dingding2.RedisInfo, len(keys))
//	//遍历所有的key，然后通过key来取到值，取到值后反序列化一下到结构体
//	for i, v := range keys {
//		info, _ := global.GLOBAL_REDIS.Get(context.Background(), v).Result()
//		fmt.Println(info)
//		err := json.Unmarshal([]byte(info), &data[i])
//		if err != nil {
//			return nil, err
//		}
//	}
//	return data, err
//}
//func SearchTele(p *params.ParamSearchUser) (teles []dingding2.Tele, err error) {
//	return mysql.SearchTele(p)
//}
