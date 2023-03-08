package dingding

import (
	"ding/global"
	"ding/model/common"
	"fmt"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	TaskID            string             `json:"task_id"`   //cron第三方定时库给的id
	TaskName          string             `json:"task_name"` //任务名字
	UserId            string             `json:"user_id"`   // 任务所属userID
	UserName          string             `json:"user_name"` //任务所属用户
	RobotId           string             `json:"robot_id"`  //任务属于机器人
	DingRobotID       uint               `json:"ding_robot_id"`
	RobotSecret       string             `json:"robot_secret"`
	RobotName         string             `json:"robot_name"`
	DetailTimeForUser string             `json:"detail_time_for_user"` //这个给用户看
	Spec              string             `json:"spec"`                 //这个是cron第三方的定时规则
	FrontRepateTime   string             `json:"front_repate_time"`    // 这个是前端传来的原始数据
	FrontDetailTime   string             `json:"front_detail_time"`
	MsgText           common.MsgText     `json:"msg_text"`
	MsgLink           common.MsgLink     `json:"msg_link"`
	MsgMarkDown       common.MsgMarkDown `json:"msg_mark_down"`
}

func (t *Task) InsertTask() (err error) {
	//我先找一下数据库中与该任务相同的id号码，如果相同的话，说明数据库中有死掉的任务，需要加上软删除
	Dtask := []Task{}
	//找到所有的死任务，进行软删除
	global.GLOAB_DB.Where("task_id = ?", t.TaskID).Find(&Dtask)
	for i := 0; i < len(Dtask); i++ {
		err := global.GLOAB_DB.Delete(&Dtask[i]).Error
		fmt.Println(err)
	}
	//然后再创建任务
	resultDB := global.GLOAB_DB.Create(&t)
	if resultDB.Error != nil {
		return resultDB.Error
	}
	return err
}
