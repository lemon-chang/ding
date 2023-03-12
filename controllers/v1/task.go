package v1

//func RemoveTask(c *gin.Context) {
//	var p params.ParamRemoveTask
//	if err := c.ShouldBindJSON(&p); err != nil {
//		zap.L().Error("RemoveTask invaild param", zap.Error(err))
//		errs, ok := err.(validator.ValidationErrors)
//		if !ok {
//			response.ResponseError(c, response.CodeInvalidParam)
//			return
//		}
//		response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
//		return
//	}
//	err := logic.RemoveTask(c, p)
//	if err != nil {
//		zap.L().Error("RemoveTask failed", zap.Error(err))
//		if errors.Is(err, mysql.ErrorNotHasTask) {
//			response.ResponseError(c, response.CodeNotRemoveTask)
//			return
//		}
//		response.ResponseError(c, response.CodeServerBusy)
//		return
//	} else {
//		response.ResponseSuccess(c, gin.H{
//			"message": "移除成功",
//		})
//	}
//}

//func GetAllActiveTask(c *gin.Context) {
//	//先删除所有的任务，然后再重新加载一遍
//	activeTasksKeys, err := global.GLOBAL_REDIS.Keys(context.Background(), fmt.Sprintf("%s*", redis.Perfix+redis.ActiveTask)).Result()
//	if err != nil {
//		zap.L().Error("从redis中获取旧的活跃任务的key失败", zap.Error(err))
//		return
//	}
//	//删除所有的key
//	global.GLOBAL_REDIS.Del(context.Background(), activeTasksKeys...)
//
//	//拿到所有的任务的id
//	entries := global.GLOAB_CORN.Entries()
//	//拿到所有任务的id
//	var entriesInt = make([]int, len(entries))
//	for index, value := range entries {
//		entriesInt[index] = int(value.ID)
//	}
//	// 根据id查询数据库，拿到详细的任务信息，存放到redis中
//	var tasks []dingding.Task                                          //拿到所有的活跃任务
//	global.GLOAB_DB.Table("tasks").Where("spec != ?", "").Find(&tasks) //查询所有的在线任务
//	//把找到的数据存储到redis中 ，现在先写成手动获取
//	//应该是存放在一个集合里面，集合里面存放着此条任务的所有信息，以id作为标识
//	//哈希特别适合存储对象，所以我们用哈希来存储
//	for _, task := range tasks {
//		taskValue, err := json.Marshal(task) //把对象序列化成为一个json字符串
//		if err != nil {
//			return
//		}
//		err = global.GLOBAL_REDIS.Set(context.Background(), redis.GetTaskKey(task.TaskID), string(taskValue), 0).Err()
//		if err != nil {
//			zap.L().Error(fmt.Sprintf("从mysql获取所有活跃任务存入redis失败，失败任务id：%s，任务名：%s,执行人：%s,对应机器人：%s", task.TaskID, task.TaskName, task.UserName, task.RobotName), zap.Error(err))
//			return
//		}
//	}
//	zap.L().Info("获取所有获取定时任务成功")
//	response.ResponseSuccess(c, "获取所有获取定时任务成功")
//}

//注意方法名要大写

//func StopTask(c *gin.Context) {
//	//如何停止任务呢？本来想的思路是通过cron生成的内部id去暂停，但是突然发现，没法这样暂停，根据id只能实现删除
//	//所以暂停功能就是把这条任务删除掉，恢复就是重新把定时任务生成一遍
//	var p params.ParamStopTask
//	if err := c.ShouldBindJSON(&p); err != nil {
//		zap.L().Error("StopTask invaild param", zap.Error(err))
//		errs, ok := err.(validator.ValidationErrors)
//		if !ok {
//			response.ResponseError(c, response.CodeInvalidParam)
//			return
//		}
//		response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
//		return
//	}
//	err := logic.StopTask(c, &p)
//	if err != nil {
//		zap.L().Error("RemoveTask failed", zap.Error(err))
//		if errors.Is(err, mysql.ErrorNotHasTask) {
//			response.ResponseError(c, response.CodeNotRemoveTask)
//			return
//		}
//		response.ResponseError(c, response.CodeServerBusy)
//		return
//	} else {
//		response.ResponseSuccess(c, gin.H{
//			"message": "暂停成功",
//		})
//	}
//}

//关闭后重新开启定时任务
//func ReStartTask(c *gin.Context) {
//	var p params.ParamReStartTask
//	if err := c.ShouldBindJSON(&p); err != nil {
//		zap.L().Error("ReStartTask invaild param", zap.Error(err))
//		errs, ok := err.(validator.ValidationErrors)
//		if !ok {
//			response.ResponseError(c, response.CodeInvalidParam)
//			return
//		}
//		response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
//		return
//	}
//	//此处只是把数据库里面的恢复了一下，具体的重启逻辑在下面
//	task, err := logic.ReStartTaskData(c, p)
//	if err != nil {
//		zap.L().Error("logic.ReStartTask() 失败", zap.Error(err))
//		response.ResponseErrorWithMsg(c, response.CodeInvalidParam, "手动重启定时任务失败")
//		return
//	}
//	//数据库中已经更新了，后面需要再跑一边定时任务的逻辑
//	//具体逻辑怎么跑呢？数据库中已经存了后端cron包所需的spec定时规则，我们直接取出来用即可
//	p1 := &dingding.ParamCronTask{}
//	err = global.GLOAB_DB.Model(&task).Preload("msg_texts.at_mobiles").First(p1).Error
//	if err != nil {
//		return
//	}
//	//p1.RobotId = task.RobotId
//	//p1.PersonNames = strings.Split(task.PersonNames, ",")
//	err, task = logic.ReStart(c, p1, task)
//	if err != nil {
//		zap.L().Error("logic.Send()失败", zap.Error(err))
//		response.ResponseErrorWithMsg(c, response.CodeInvalidParam, "logic.send函数，调用钉钉机器人第三方接口出错")
//		return
//	}
//
//	response.ResponseSuccess(c, gin.H{
//		"data":        "开启成功",
//		"new_task_id": task.TaskID,
//	})
//
//}

//管理员权限,查看所有用户信息
