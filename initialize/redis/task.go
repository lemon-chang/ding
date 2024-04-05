package redis

//ding:activeTask:id
func GetTaskKey(task_id string) string {
	return Prefix + ActiveTask + task_id
}
