package dingding

import (
	"context"
	"ding/global"
	myselfRedis "ding/initialize/redis"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"strconv"
)

func (d *DingUser) Sign(year, uporDown, startWeek, weekDay, MNE int) (ConsecutiveSignNum int, err error) {
	//MNE 是上午下午晚上 1 2 3
	//构建redis中的key //singKey  user:sign:5:2023:1:19周        用户5在2023上半年第19周签到的记录
	//构建redis中的key //singKey  user:sign:5:2023:2:19周        用户5在2023下半年第19周签到的记录
	if weekDay == 0 {
		weekDay = 7
	}
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v:%v", d.UserId, year, uporDown, startWeek)
	//根据date能够判断出来，现在是第几周的上午下午晚上等
	offset := int64((weekDay-1)*3 + MNE - 1)
	IsSigned := global.GLOBAL_REDIS.GetBit(context.Background(), key, offset).Val()
	if IsSigned == 1 {
		ConsecutiveSignNum, _ = d.GetConsecutiveSignNum(year, uporDown, startWeek, weekDay, MNE)
		return ConsecutiveSignNum, errors.New("当前日期已经打卡签到，无需再次打卡签到")
	}
	//用户没有签到，设置成签到即可
	i, err := global.GLOBAL_REDIS.SetBit(context.Background(), key, offset, 1).Result()
	if err != nil || i != 1 {
		//此处返回的是设置前的值
		zap.L().Error("签到时操作redis中的位图失败", zap.Error(err))
	}
	ConsecutiveSignNum, _ = d.GetConsecutiveSignNum(year, uporDown, startWeek, weekDay, MNE)
	return
}
func (d *DingUser) GetConsecutiveSignNum(year, uporDown, startWeek, weekDay, MNE int) (ConsecutiveSignNum int, err error) {
	if startWeek == 0 {
		startWeek = 7
	}
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v:%v", d.UserId, year, uporDown, startWeek)
	//bitfield可以操作多个位 bitfile user:sign:2023:1:19 u7 0  //从索引零开始，往后面统计7天的
	//cmd := global.GLOBAL_REDIS.Do(context.Background(), "BITFIELD", key, "GET", "u"+strconv.Itoa(weekDay), "0").
	list, err := global.GLOBAL_REDIS.BitField(context.Background(), key, "GET", "u"+strconv.Itoa(weekDay), "0").Result()
	if err != nil || list == nil || len(list) == 0 || list[0] == 0 {
		return 0, nil
	}
	// 此处获得的值是经过二进制转化过来的，总共有21个字节，如果长度是21个字节的话，可能会非常的大，我们如何处理非常大的值呢？
	//具体思路可以使用位运算，具体博客链接
	v := list[0]
	for i := weekDay; i > 0; i-- {
		for j := 0; j < 3; j++ {
			//如果这个很大的数字转化为二进制之后，左移动一位，右移动一位，如果还等于自己，说明最后一位是0，表示没有签到
			if v>>1<<1 == v {
				if !(i == weekDay && j == MNE) {
					//低位为0 且 非当天早中晚应该签到的时间，签到中断
					break
				}
			} else {
				//说明签到了
				ConsecutiveSignNum++
			}
		}
		//将v右移一位，并重新复制，相当于最低位提前了一天
		v = v >> 1
	}
	return
}

// 统计用户当前周签到的详情情况–
func (d *DingUser) GetWeekSignDetail(year, uporDown, startWeek int) (result map[int][]bool, err error) {
	result = make(map[int][]bool, 0)
	//if year == 0 || uporDown == 0 || startWeek == 0 {
	//	curTime, _ := (&localTime.MySelfTime{}).GetCurTime(nil)
	//
	//}
	//使用bitFiled来获取int64，然后使用位运算计算结果
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v:%v", d.UserId, year, uporDown, startWeek)
	fmt.Println(key)
	list, err := global.GLOBAL_REDIS.BitField(context.Background(), key, "GET", "u"+strconv.Itoa(21), "0").Result()
	if err != nil || list == nil || len(list) == 0 || list[0] == 0 {
		zap.L().Error("使用redis中的bitmap失败", zap.Error(err))
		return nil, errors.New("使用redis中的bitmap失败")
	}
	v := list[0]
	//110001111111111101111000
	//for i := 1; i <= 8; i++ {
	//	if v>>1<<1 == v {
	//		//说明没有签到
	//		result[i] = append(result[i], false)
	//	} else {
	//		//说明签到了
	//		result[i] = append(result[i], true)
	//	}
	//	v = v >> 1
	//	x = x >> 1
	//}
	for i := 7; i > 0; i-- {
		for j := 0; j < 3; j++ {
			if v>>1<<1 == v {
				//说明没有签到
				result[i] = append(result[i], false)
			} else {
				//说明签到了
				result[i] = append(result[i], true)
				//result[i][j] = true
			}
			v = v >> 1
		}
	}
	return
}

// 统计用户一周的签到次数（非连续）
func (d *DingUser) GetWeekSignNum(year, uporDown, startWeek int) (WeekSignNum int64, err error) {
	//需要使用redis中的bitmap中bigcount方法来统计
	//构建key
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v:%v", d.UserId, year, uporDown, startWeek)

	bitCount := &redis.BitCount{
		Start: 0, //都设置成0就是涵盖整个bitmap
		End:   0,
	}
	WeekSignNum, err = global.GLOBAL_REDIS.BitCount(context.Background(), key, bitCount).Result()
	if err != nil {
		zap.L().Error("使用redis的BitCount失败", zap.Error(err))
		return
	}
	return
}
func (d *DingUser) SendFrequencyLeave(start int) error {
	fmt.Println("推送个人请假频率")
	return nil
}
func (d *DingUser) CountFrequencyLeave(startWeek int, result map[string][]DingAttendance) error {
	fmt.Println("存储个人请假频率")
	return nil
}
