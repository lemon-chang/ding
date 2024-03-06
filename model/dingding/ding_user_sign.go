package dingding

import (
	"context"
	"ding/global"
	myselfRedis "ding/initialize/redis"
	"ding/model/common/localTime"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"strconv"
)

func (d *DingUser) Sign(semester string, startWeek, weekDay, MNE int) (getWeekSignNum int64, ConsecutiveSignNum int, err error) {
	// MNE 是上午下午晚上 1 2 3
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v", semester, startWeek, d.UserId)
	// 我们一个key代表的是一周的签到，offset可以帮助我们定位到当前是在一周当中的哪一位
	offset := int64((weekDay-1)*3 + MNE - 1)
	IsSigned := global.GLOBAL_REDIS.GetBit(context.Background(), key, offset).Val()
	if IsSigned == 1 {
		err = errors.New("当前日期已经打卡签到，无需再次打卡签到")
		return
	}
	// 用户没有签到，设置成签到即可
	_, err = global.GLOBAL_REDIS.SetBit(context.Background(), key, offset, 1).Result()
	if err != nil {
		//此处返回的是设置前的值
		zap.L().Error("签到时操作redis中的位图失败", zap.Error(err))
	}
	// 统计用户连续签到次数
	ConsecutiveSignNum, err = d.GetConsecutiveSignNum(semester, startWeek, weekDay, MNE)
	if err != nil {
		zap.L().Error("统计用户连续签到次数失败", zap.Error(err))
	}
	// 统计用户这周签到的总次数（非连续）
	getWeekSignNum, err = d.GetWeekSignNum(semester, startWeek)
	if err != nil {
		zap.L().Error("统计用户这周签到的总次数（非连续）失败", zap.Error(err))
	}
	return
}

// GetConsecutiveSignNum 当前周中，获取用户连续签到数量
func (d *DingUser) GetConsecutiveSignNum(semester string, startWeek, weekDay, MNE int) (ConsecutiveSignNum int, err error) {
	curTime := &localTime.MySelfTime{}
	err = curTime.GetCurTime(nil)
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v", semester, startWeek, d.UserId)
	offset := int64((weekDay-1)*3 + MNE - 1)
	// bitfield可以操作多个位 bitfield sign:2023-2024学年第二学期:2:413550622937553255 u21 0  //u表示无符号位置，7表示往后面统计7位的，0表示从第0位开始统计
	// 如果bitmap中的byte（8个二进制位）没有到达21次的，后续自动补零了
	list, err := global.GLOBAL_REDIS.BitField(context.Background(), key, "GET", "u"+strconv.Itoa(int(offset)+1), "0").Result()
	if err != nil || list == nil || len(list) == 0 || list[0] == 0 {
		return 0, err
	}
	// 此处获得的值是经过二进制转化过来的，总共有21个字节，如果长度是21个字节的话，可能会非常的大，我们如何处理非常大的值呢？具体思路可以使用位运算，具体博客链接
	v := list[0]
	for i := offset; i >= 0; i-- {
		//如果这个很大的数字转化为二进制之后，左移动一位，右移动一位，如果还等于自己，说明最后一位是0，表示没有签到
		if v>>1<<1 == v {
			if i == offset {
				continue
			} else {
				//低位为0 且 非当天早中晚应该签到的时间，签到中断
				break
			}
		} else {
			//说明签到了
			ConsecutiveSignNum++
		}
		//将v右移一位，并重新赋值，相当于最低位提前了一天
		v = v >> 1
	}
	return
}

// GetWeekSignDetail 统计用户当前周签到的详情情况(用于前端构建日历控件)
func (d *DingUser) GetWeekSignDetail(semester string, startWeek int) (result map[int][]bool, err error) {
	result = make(map[int][]bool, 0)
	//使用bitFiled来获取int64，然后使用位运算计算结果
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v", semester, startWeek, d.UserId)
	list, err := global.GLOBAL_REDIS.BitField(context.Background(), key, "GET", "u"+strconv.Itoa(21), "0").Result()
	if err != nil || list == nil || len(list) == 0 || list[0] == 0 {
		zap.L().Error("使用redis中的bitmap失败", zap.Error(err))
		return nil, err
	}
	v := list[0]
	for i := 7; i > 0; i-- {
		for j := 0; j < 3; j++ {
			if v>>1<<1 == v {
				//说明没有签到
				result[i] = append(result[i], false)
			} else {
				//说明签到了
				result[i] = append(result[i], true)
			}
			v = v >> 1
		}
	}
	return
}

// GetWeekSignNum 统计用户一周的签到次数（非连续）
func (d *DingUser) GetWeekSignNum(Semester string, startWeek int) (WeekSignNum int64, err error) {
	//需要使用redis中的bitmap中bigcount方法来统计
	//构建key
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v", Semester, startWeek, d.UserId)
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
