package dingding

import (
	"context"
	"ding/global"
	redis2 "ding/initialize/redis"
	"ding/model/common"
	"ding/model/params"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

type Task struct {
	gorm.Model
	TaskID            int                 `json:"task_id" `   //cron第三方定时库给的id
	IsSuspend         bool                `json:"is_suspend"` //是否暂停
	TaskName          string              `json:"task_name"`  //任务名字
	UserId            string              `json:"user_id"`    // 任务所属userID
	UserName          string              `json:"user_name"`  //任务所属用户
	RobotId           string              `json:"robot_id"`   //任务属于机器人
	RobotName         string              `json:"robot_name"`
	DetailTimeForUser string              `json:"detail_time_for_user"` //这个给用户看
	Spec              string              `json:"spec"`                 //这个是cron第三方的定时规则
	FrontRepeatTime   string              `json:"front_repeat_time"`    // 这个是前端传来的原始数据
	FrontDetailTime   string              `json:"front_detail_time"`
	MsgText           *common.MsgText     `json:"msg_text"`
	MsgLink           *common.MsgLink     `json:"msg_link"`
	MsgMarkDown       *common.MsgMarkDown `json:"msg_mark_down"`
	NextTime          time.Time           `json:"next_time"`
}

func (t *Task) GetTaskList(p *ParamGetTaskList, c *gin.Context) (tasks []Task, total int64, err error) {
	limit := p.PageSize
	offset := p.PageSize * (p.Page - 1)
	userid, _ := c.Get(global.CtxUserIDKey)
	db := global.GLOAB_DB.Where("user_id = ?", userid.(string))
	if p.TaskName != "" {
		db = db.Where("task_name like ?", "%"+p.TaskName+"%")
	}
	if p.IsSuspend {
		db = db.Where("is_suspend = ?", p.IsSuspend)
	}
	if p.RobotId != "" {
		db = db.Where("robot_id = ?", p.RobotId)
	}
	err = db.Limit(limit).Offset(offset).Find(&tasks).Count(&total).Error
	if err != nil {
		return
	}
	return
}
func (t *Task) Insert(p *ParamCronTask) (err error) {
	affected := global.GLOAB_DB.First(&Task{}, "task_name = ? and user_id =?", t.TaskName, t.UserId).RowsAffected
	if affected > 0 {
		return errors.New("该定时任务名称已存在")
	}
	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(&t).Error
		if err != nil {
			return err
		}
		// 此处赋值id，是为了让一次性的定时任务给删除掉
		p.ID = t.ID
		tasker := func() {
			task := &Task{Model: gorm.Model{ID: p.ID}}
			err = global.GLOAB_DB.Preload("MsgText.Text").Preload("MsgText.At.AtMobiles").First(&task).Error
			if !task.IsSuspend {
				zap.L().Info("定时任务被暂停，无需执行")
				return
			}
			// 重新组装一个发送参数p
			p := &ParamCronTask{
				MsgText: task.MsgText,
				RobotId: p.RobotId,
			}
			err := (&DingRobot{RobotId: task.RobotId}).SendMessage(p)
			if err != nil {
				zap.L().Error("发送定时任务失败", zap.Error(err))
				return
			}
			// spec以？ 结尾的话，就是一次性定时任务
			if task.Spec[len(task.Spec)-1] == '?' {
				err = task.Remove()
				if err != nil {
					zap.L().Error("移除一次性定时任务失败", zap.Error(err))
				}
				return
			} else {
				zap.L().Info(fmt.Sprintf("此定时任务:%v不是一次性定时任务", task.ID), zap.Error(err))
				// 更新下一次执行时间
				err = global.GLOAB_DB.Model(&task).Update("next_time", global.GLOAB_CORN.Entry(cron.EntryID(task.TaskID)).Next).Error
				if err != nil {
					zap.L().Error(fmt.Sprintf("定时任务%v更新下次执行时间失败", t.ID), zap.Error(err))
				}
				return
			}
		}
		TaskID, err := global.GLOAB_CORN.AddFunc(t.Spec, tasker)
		if err != nil {
			return err
		}
		err = tx.Model(t).Update("task_id", TaskID).Error
		if err != nil {
			return err
		}
		return err
	})

	return
}
func (t *Task) GetAllActiveTask() (tasks []Task, err error) {
	//先删除所有的任务，然后再重新加载一遍
	activeTasksKeys, err := global.GLOBAL_REDIS.Keys(context.Background(), fmt.Sprintf("%s*", redis2.ActiveTask)).Result()
	if err != nil {
		zap.L().Error("从redis中获取旧的活跃任务的key失败", zap.Error(err))
		return
	}
	//删除所有的key
	total, err := global.GLOBAL_REDIS.Del(context.Background(), activeTasksKeys...).Result()
	if err != nil {
		return
	}
	zap.L().Info(fmt.Sprintf("redis中删除活跃定时任务key %s 个", total))
	//拿到所有的任务的id
	entries := global.GLOAB_CORN.Entries()
	//拿到所有任务的id
	var entriesInt = make([]int, len(entries))
	for index, value := range entries {
		entriesInt[index] = int(value.ID)
	}
	// 根据id查询数据库，拿到详细的任务信息，存放到redis中
	global.GLOAB_DB.Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").Where("deleted_at is null").Find(&tasks, entriesInt)
	//查询所有的在线任务
	//把找到的数据存储到redis中 ，现在先写成手动获取
	//应该是存放在一个集合里面，集合里面存放着此条任务的所有信息，以id作为标识
	//哈希特别适合存储对象，所以我们用哈希来存储
	for _, task := range tasks {
		taskValue, err := json.Marshal(task) //把对象序列化成为一个json字符串
		if err != nil {
			zap.L().Info("定时任务序列化失败", zap.Error(err))
			continue
		}
		err = global.GLOBAL_REDIS.Set(context.Background(), redis2.GetTaskKey(strconv.Itoa(task.TaskID)), string(taskValue), 0).Err()
		if err != nil {
			zap.L().Error(fmt.Sprintf("从mysql获取所有活跃任务存入redis失败，失败任务id：%s，任务名：%s,执行人：%s,对应机器人：%s", task.TaskID, task.TaskName, task.UserName, task.RobotName), zap.Error(err))
			continue
		}
	}
	return
}

func (t *Task) GetTasks(user_id string, p *params.ParamGetTasks) (tasks []Task, err error) {
	db := global.GLOAB_DB
	if p.TaskName != "" {
		db.Where("task_name = ?", p.TaskName)
	}
	if p.OnlyActive == 1 {
		//拿到所有的任务的id
		entries := global.GLOAB_CORN.Entries()
		//拿到所有任务的id
		var entriesInt = make([]int, len(entries))
		for index, value := range entries {
			entriesInt[index] = int(value.ID)
		}
		db = db.Where("task_id in ?", entriesInt)
	}
	limit := p.PageSize
	offset := p.PageSize * (p.Page - 1)
	err = db.Limit(limit).Offset(offset).Where("user_id = ? ", user_id).Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").Find(&tasks).Error
	return
}
func (t *Task) Remove() (err error) {
	err = global.GLOAB_DB.First(&t).Error
	if err != nil {
		return err
	}
	global.GLOAB_CORN.Remove(cron.EntryID(t.TaskID))
	//到了这里就说明我有这个定时任务，我要移除这个定时任务
	err = global.GLOAB_DB.Unscoped().Delete(&t).Error
	if err != nil {
		zap.L().Error("删除定时任务失败", zap.Error(err))
		return err
	}

	return err
}
func (t *Task) UpdateTask(p *UpdateTask) (err error) {
	err = global.GLOAB_DB.Preload("MsgText.Text").Preload("MsgText.At.AtMobiles").Preload(clause.Associations).First(t).Error
	if err != nil {
		return err
	}
	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		// 检查一下定时规则是否发生了改变
		if t.Spec != p.Spec {
			// 需要重新做定时任务
			global.GLOAB_CORN.Remove(cron.EntryID(t.TaskID))
			if err != nil {
				return err
			}
			tasker := func() {
				task := &Task{Model: gorm.Model{ID: t.ID}}
				err = global.GLOAB_DB.Preload("MsgText.Text").Preload("MsgText.At.AtMobiles").Preload(clause.Associations).First(&task).Error
				if task.IsSuspend {
					zap.L().Info("定时任务暂停，无需执行")
					return
				}
				// 重新组装一个发送参数p
				p := &ParamCronTask{
					MsgText: task.MsgText,
					RobotId: task.RobotId,
				}
				err = (&DingRobot{RobotId: task.RobotId}).SendMessage(p)
				if err != nil {
					zap.L().Error("发送定时任务失败", zap.Error(err))
					return
				}
				// spec以？ 结尾的话，就是一次性定时任务
				if task.Spec[len(task.Spec)-1] == '?' {
					err = task.Remove()
					if err != nil {
						zap.L().Error("移除一次性定时任务失败", zap.Error(err))
					}
				} else {
					zap.L().Info(fmt.Sprintf("此定时任务:%v不是一次性定时任务", task.ID), zap.Error(err))
					// 更新下一次执行时间
					err = global.GLOAB_DB.Model(&task).Update("next_time", global.GLOAB_CORN.Entry(cron.EntryID(task.TaskID)).Next).Error
					if err != nil {
						zap.L().Error(fmt.Sprintf("定时任务%v更新下次执行时间失败", t.ID), zap.Error(err))
					}
					return
				}
			}
			task_id, err := global.GLOAB_CORN.AddFunc(p.Spec, tasker)
			if err != nil {
				return err
			}
			t.TaskID = int(task_id)
			t.NextTime = global.GLOAB_CORN.Entry(cron.EntryID(t.TaskID)).Next

		}
		t.Spec = p.Spec
		t.IsSuspend = p.IsSuspend
		t.TaskName = p.TaskName
		t.RobotId = p.RobotId
		t.DetailTimeForUser = p.DetailTimeForUser
		t.FrontRepeatTime = p.RepeatTime
		t.FrontDetailTime = p.DetailTime
		err = tx.Select("task_name", "is_suspend", "robot_id", "task_id", "detail_time_for_user", "spec", "next_time", "front_detail_time", "front_repeat_time").Updates(t).Error
		t.MsgText.Text.Content = p.MsgText.Text.Content
		t.MsgText.At.IsAtAll = p.MsgText.At.IsAtAll
		err = tx.Select("content").Updates(t.MsgText.Text).Error
		if err != nil {
			return err
		}
		err = tx.Select("is_at_all").Updates(t.MsgText.At).Error
		if err != nil {
			return err
		}
		At := t.MsgText.At
		err = tx.Model(&At).Association("AtMobiles").Replace(p.MsgText.At.AtMobiles)
		return err
	})
	return err
}

func (t *Task) GetTaskDetailByID() (err error) {

	return global.GLOAB_DB.Preload("MsgText.Text").Preload("MsgText.At.AtMobiles").Preload(clause.Associations).First(&t).Error

}
