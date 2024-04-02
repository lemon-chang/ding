package chatRobot

import (
	"ding/global"
	"errors"
	"gorm.io/gorm"
)

type RobotStream struct {
	ID      int64          `gorm:"primaryKey" json:"userid"`
	Deleted gorm.DeletedAt `json:"omitempty"`
	UserId  string         `json:"user_id"`
	DeptId  int            `json:"dept_id"`
	Type    int            `json:"type"`
	Title   string         `json:"title"`
	Msg     string         `json:"msg"`
}

func NewRobotStream() *RobotStream {
	return &RobotStream{}
}

func (d *RobotStream) GetPersonMsg(title string) (msg map[string]string, err error) {
	robotStreams := []RobotStream{}
	likeStr := "%" + title + "%"
	err = global.GLOAB_DB.Model(d).Select("msg").Where("user_id=? and type = ? and title like ?", d.UserId, 1, likeStr).Find(&robotStreams).Error

	for i := 0; i < len(robotStreams); i++ {
		msg[robotStreams[i].Title] = robotStreams[i].Msg
	}
	return
}
func (d *RobotStream) GetDeptMsg(title string) (msg map[string]string, err error) {
	robotStreams := []RobotStream{}
	likeStr := "%" + title + "%"
	err = global.GLOAB_DB.Model(d).Select("msg").Where("dept_id=? and  type = ? and title like ?", d.UserId, 2, likeStr).Find(&robotStreams).Error
	for i := 0; i < len(robotStreams); i++ {
		msg[robotStreams[i].Title] = robotStreams[i].Msg
	}
	return
}
func (d *RobotStream) GetAllUserMsg(title string) (msg map[string]string, err error) {
	robotStreams := []RobotStream{}
	likeStr := "%" + title + "%"
	err = global.GLOAB_DB.Model(d).Select("msg").Where("type = ? and title like ?", 3, likeStr).Find(&robotStreams).Error
	for i := 0; i < len(robotStreams); i++ {
		msg[robotStreams[i].Title] = robotStreams[i].Msg
	}
	return
}

// 添加信息
func (d *RobotStream) CreateRobotStream() (err error) {
	var tmp RobotStream
	err = global.GLOAB_DB.Model(tmp).Find(tmp).Limit(1).Error
	if err != nil {
		return
	}
	if tmp.Title != "" {
		return errors.New("存在相同的")
	}
	err = global.GLOAB_DB.Create(d).Error
	return
}

// 修改信息
func (d *RobotStream) UpdateRobotStream() (err error) {
	err = global.GLOAB_DB.Model(d).Select("").Updates(d).Error
	return
}
