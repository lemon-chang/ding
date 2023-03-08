package redis

import (
	"context"
	"ding/global"
	"time"
)

//通过openConversationId（加密的会话ID）和 senderStaffId （也叫userid）可以确定不同群的唯一成员。
func SetAttendanceState(userid,conversationId string)(err error) {
	err = global.GLOBAL_REDIS.Set(context.Background(), GetAttendanceKey(userid, conversationId), "已存在", 0).Err()
	return err
}

//ding:activeTask:id
func GetAttendanceKey(userId,conversationId string) string {
	return Perfix + Attendance + userId + "-" + conversationId
}
func TTLAttendanceKey(key string) (time.Duration,error) {
	result, err := global.GLOBAL_REDIS.TTL(context.Background(), key).Result()
	return result,err
}
