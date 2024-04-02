package cron

import (
	"ding/global"
	"ding/initialize/viper"
	"ding/model/common/localTime"
	"ding/model/dingding"
	"fmt"
	"go.uber.org/zap"
	"runtime"
	"sort"
	"strconv"
)

// 考勤周报
func AttendWeekPaper() (err error) {
	//获取考勤组列表
	var groupList []dingding.DingAttendGroup
	err = global.GLOAB_DB.Find(&groupList).Error
	if err != nil {
		zap.L().Error("获取考勤组列表失败", zap.Error(err))
		return
	}
	for _, group := range groupList {
		if !group.IsRobotAttendance && !group.IsWeekPaper {
			zap.L().Warn(fmt.Sprintf("考勤组：%v 开启未机器人考勤 或者未开启考勤推送", group.GroupName))
			continue
		}
		err := CountWeekPaper(group.GroupId)
		if err != nil {
			zap.L().Error("打卡信息获取失败", zap.Error(err))
			continue
		}
	}
	return err
}

// 开始统计周报
func CountWeekPaper(GroupId int) (err error) {
	//使用定时器,定时什么时候发送,每周的周日发
	spec := ""
	if runtime.GOOS == "windows" {
		spec = "00 02,11,11 9,14,20 * * ?"
	} else if runtime.GOOS == "linux" {
		spec = "0 15 10 ? * SAT"
	}
	curTime := localTime.MySelfTime{}
	err = curTime.GetCurTime(nil)
	if err != nil {
		zap.L().Error("获取localTime.MySelfTime{}失败", zap.Error(err))
	}
	//开启定时任务
	task := func() {
		//获取这个组织里面的成员信息
		token, _ := (&dingding.DingToken{}).GetAccessToken()
		g := dingding.DingAttendGroup{GroupId: GroupId, DingToken: dingding.DingToken{Token: token}}
		depts, err := g.GetGroupDeptNumber()
		if err != nil {
			zap.L().Error("获取考勤组部门成员(已经筛掉了不参与考勤的个人)失败", zap.Error(err))
			return
		}
		//获取考勤数据
		for _, dingUsers := range depts {
			num := make(map[string]int64)
			for _, user := range dingUsers {
				//拿到redis里面存的考勤信息
				WeekSignNum, err := user.GetWeekSignNum(curTime.Semester, curTime.Week)
				if err != nil {
					zap.L().Error("统计失败", zap.Error(err))
					continue
				}
				num[user.UserId] = WeekSignNum
			}
			//查完一个部门后进行排序，然后发给用户
			sortnum := SortResult(num)
			//定义一个排名用于记录该用户排名
			ranking := 1
			for _, n := range sortnum {
				var ids = make([]string, 10)
				ids[0] = n.UserId
				//封装要发送的消息
				message := EditMessage(ranking, n.WeekSignNum)
				//将考勤数据发给该人
				p := &dingding.ParamChat{
					RobotCode: viper.Conf.MiniProgramConfig.RobotCode,
					UserIds:   ids,
					MsgKey:    "sampleText",
					MsgParam:  message,
				}
				err = (&dingding.DingRobot{}).ChatSendMessage(p)
				ranking++
			}
		}
	}
	_, err = global.GLOAB_CORN.AddFunc(spec, task)
	if err != nil {
		zap.L().Error("启动周报推送定时任务失败", zap.Error(err))
		return
	}
	return err
}

// 编辑发送的消息
func EditMessage(ranking int, num int64) (message string) {
	message = "本周打卡记录：" + "\n" +
		"打卡次数：" + strconv.Itoa(int(num)) + "\n" +
		"打卡异常次数：" + strconv.Itoa(int(18-int(num))) + "\n" +
		"你在你部门的排名为：" + strconv.Itoa(ranking)
	return
}

type WeeklyNewPaper struct {
	UserId      string
	WeekSignNum int64
}

// 排序
func SortResult(num map[string]int64) (WeeklyNewPapers []WeeklyNewPaper) {
	for k, v := range num {
		WeeklyNewPapers = append(WeeklyNewPapers, WeeklyNewPaper{k, v})
	}
	//这个地方也可以使用快排排序
	sort.Slice(WeeklyNewPapers, func(i, j int) bool {
		return WeeklyNewPapers[i].WeekSignNum > WeeklyNewPapers[j].WeekSignNum
	})
	return WeeklyNewPapers
	return
}
