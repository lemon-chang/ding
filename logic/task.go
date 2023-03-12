package logic

//func StopTask(c *gin.Context, Dtask *params.ParamStopTask) (err error) {
//	task := dingding.Task{
//		TaskID: Dtask.TaskId,
//	}
//	taskID, err := mysql.StopTask(task)
//
//	if errors.Is(err, mysql.ErrorNotHasTask) {
//		return mysql.ErrorNotHasTask
//	}
//	global.GLOAB_CORN.Remove(cron.EntryID(taskID))
//	return err
//}
//func ReStartTaskData(c *gin.Context, Ptask params.ParamReStartTask) (task dingding.Task, err error) {
//	return mysql.ReStartTaskData(c, Ptask)

//}
//func RemoveTask(c *gin.Context, p params.ParamRemoveTask) (err error) {
//	task := dingding.Task{
//		TaskID: p.TaskId,
//	}
//	taskID, err := mysql.RemoveTask(task)
//
//	global.GLOAB_CORN.Remove(cron.EntryID(taskID))
//	return err
//}
