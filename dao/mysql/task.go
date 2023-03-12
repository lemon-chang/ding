package mysql

//func InsertTask(task dingding.Task) (err error) {
//	//我先找一下数据库中与该任务相同的id号码，如果相同的话，说明数据库中有死掉的任务，需要加上软删除
//	Dtask := []dingding.Task{}
//	//找到所有的死任务，进行软删除
//	global.GLOAB_DB.Where("task_id = ?", task.TaskID).Find(&Dtask)
//	for i := 0; i < len(Dtask); i++ {
//		err := global.GLOAB_DB.Delete(&Dtask[i]).Error
//		fmt.Println(err)
//	}
//	//然后再创建任务
//	resultDB := global.GLOAB_DB.Create(&task)
//	if resultDB.Error != nil {
//		return resultDB.Error
//	}
//	return err
//}
//func UpdateTask(task dingding.Task) (err error) {
//	err = global.GLOAB_DB.Model(&task).Update("task_id", task.TaskID).Error
//	return err
//}
//
////移除这个任务并返回这个任务的id
//func StopTask(task dingding.Task) (taskID int, err error) {
//	////先来判断一下是否拥有这个定时任务
//	//var task1 dingding.Task
//	//result := global.GLOAB_DB.Where("task_id = ?", task.TaskID).First(&task1)
//	//if result.Error != nil {
//	//	return 0, result.Error
//	//}
//	//taskID, err = strconv.Atoi(task1.TaskID)
//	//if err != nil {
//	//	return 0, err
//	//}
//	////到了这里就说明我有这个定时任务，我要移除这个定时任务
//	//result = global.GLOAB_DB.Delete(&task1)
//	//if result.Error != nil {
//	//	return 0, result.Error
//	//}
//	return taskID, err
//}
//func RemoveTask(task dingding.Task) (taskID int, err error) {
//	////先来判断一下是否拥有这个定时任务
//	//var task1 dingding.Task
//	//result := global.GLOAB_DB.Unscoped().Where("task_id = ?", task.TaskID).First(&task1)
//	//if result.Error != nil {
//	//	return 0, result.Error
//	//}
//	//taskID, err = strconv.Atoi(task1.TaskID)
//	//if err != nil {
//	//	return 0, err
//	//}
//	////到了这里就说明我有这个定时任务，我要移除这个定时任务
//	//result = global.GLOAB_DB.Unscoped().Delete(&task1) //硬删除
//	//if result.Error != nil {
//	//	return 0, result.Error
//	//}
//	return taskID, err
//}
//
////通过定时任务的id（非task_id）获取到定时任务
//func GetTaskByID(c *gin.Context, id uint) (task dingding.Task, err error) {
//	err = global.GLOAB_DB.Unscoped().First(&task, id).Error
//	if err != nil {
//		zap.L().Error("通过主键id查询定时任务失败", zap.Error(err))
//		return
//	}
//	return
//}

//定时任务关闭后重新启动定时任务
//func ReStartTaskData(c *gin.Context, Ptask params.ParamReStartTask) (task dingding.Task, err error) {
//	task, err = GetTaskByID(c, Ptask.ID)
//	//根据这个id主键查询到被删除的数据
//	err = global.GLOAB_DB.Unscoped().Model(&task).Update("deleted_at", nil).Error //这个地方必须加上Unscoped()，否则不报错，但是却无法更新
//	if err != nil {
//		username, _ := c.Get(global.CtxUserNameKey)
//		zap.L().Error(fmt.Sprintf("根据任务的id主键查询被删除的任务失败，失败id:%s，任务所属人：%s", Ptask.ID, username), zap.Error(err))
//		return task, err
//	}
//
//	return task, err
//}
