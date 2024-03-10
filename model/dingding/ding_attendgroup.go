package dingding

import (
	"context"
	"crypto/tls"
	"ding/global"
	"ding/initialize/redis"
	"ding/model/classCourse"
	"ding/model/common"
	"ding/model/common/localTime"
	"ding/model/common/request"
	"ding/utils"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	redisZ "github.com/go-redis/redis/v8"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DingAttendGroup struct {
	CreatedAt     time.Time      `gorm:"column:create_at"`
	UpdatedAt     time.Time      `gorm:"column:update_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at"`
	GroupId       int            `gorm:"primaryKey" json:"group_id"` //考勤组id
	GroupName     string         `json:"group_name"`                 //考勤组名称
	MemberCount   int            `json:"member_count"`               //参与考勤人员总数
	WorkDayList   []string       `gorm:"-" json:"work_day_list"`     //0表示休息,数组内的值，从左到右依次代表周日到周六，每日的排班情况。
	ClassesList   []string       `gorm:"-" json:"classes_list"`      // 一周的班次时间展示列表
	SelectedClass []struct {
		Setting struct {
			PermitLateMinutes int `json:"permit_late_minutes"` //允许迟到时长
		} `gorm:"-" json:"setting"`
		Sections []struct {
			Times []struct {
				CheckTime string `json:"check_time"` //打卡时间
				CheckType string `json:"check_type"` //打卡类型
			} `gorm:"-" json:"times"`
		} `gorm:"-" json:"sections"`
	} `gorm:"-" json:"selected_class"`
	DingToken         `gorm:"-"`
	IsRobotAttendance bool       `json:"is_robot_attendance"`  //该考勤组是否开启机器人查考勤 （相当于是总开关）
	RobotAttendTaskID int        `json:"robot_attend_task_id"` // 考勤组对应的task_id
	RobotAlterTaskID  int        `json:"robot_alter_task_id"`  // 考勤组提醒对应的task_id
	IsSendFirstPerson int        `json:"is_send_first_person"` //该考勤组是否开启推送每个部门第一位打卡人员 （总开关）
	IsInSchool        bool       `json:"is_in_school"`         //是否在学校，如果在学校，开启判断是否有课
	IsAlert           bool       `json:"is_alert"`             //是否预备
	AlertTime         int        `json:"alert_time"`           //如果预备了，提前几分钟
	DelayTime         int        `json:"delay_time"`           //推迟多少分钟
	NextTime          string     `json:"next_time"`            //下次执行时间
	IsSecondClass     int        `json:"is_second_class"`      //是否开启第二节课考勤
	RestTimes         []RestTime `json:"rest_times" gorm:"foreignKey:AttendGroupID;references:group_id"`
	IsWeekPaper       bool       `json:"is_week_paper"`
}
type RestTime struct {
	gorm.Model    // 1 2 2 0 2 1
	WeekDay       int
	MAE           int // 0 1 2
	AttendGroupID int
}
type DingAttendanceGroupMemberList struct {
	AtcFlag  string `json:"atc_flag"`
	Type     string `json:"type"`
	MemberID string `json:"member_id"`
}

func (DingAttendGroup *DingAttendGroup) BeforeCreate(tx *gorm.DB) (err error) {
	DingAttendGroup.CreatedAt = time.Now()
	return
}

func (DingAttendGroup *DingAttendGroup) BeforeUpdate(tx *gorm.DB) (err error) {
	DingAttendGroup.UpdatedAt = time.Now()
	return
}

// 批量获取考勤组
func (a *DingAttendGroup) GetAttendancesGroups(offset int, size int) (groups []DingAttendGroup, err error) {
	if !a.DingToken.IsLegal() {
		return nil, errors.New("token不合法")
	}
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/attendance/getsimplegroups?access_token=" + a.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		Offset int
		Size   int
	}{
		Offset: offset,
		Size:   size,
	}
	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		DingResponseCommon
		Result struct {
			Groups []DingAttendGroup `json:"groups"`
		} `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	groups = r.Result.Groups
	err = global.GLOAB_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"group_name", "member_count"}),
	}).Create(&groups).Error
	return groups, nil
}

// 获取一天的上下班时间
// map["OnDuty"] map["OffDuty"]
func (a *DingAttendGroup) GetCommutingTimeAndSpec() (commutingTime, AlterTime map[string][]string, AttendSpec string, AlertSpec string, restTime []RestTime, isInSchool bool, err error) {
	commutingTime, AlterTime = make(map[string][]string, 2), make(map[string][]string, 2)
	timeNowYMD := time.Now().Format("2006-01-02")
	err = a.GetAttendancesGroupById()
	if err != nil {
		return
	}
	commutingOnDutyTime := make([]string, 0)
	commutingOffDutyTime := make([]string, 0)
	AlterOnDutyTime := make([]string, 0)
	AlterOffDutyTime := make([]string, 0)
	for _, section := range a.SelectedClass[0].Sections {
		for _, time := range section.Times {
			b := []byte(time.CheckTime[len(time.CheckTime)-8:])
			if time.CheckType == "OnDuty" {
				s := strings.Split(string(b), ":")
				h, _ := strconv.Atoi(s[0])
				m, _ := strconv.Atoi(s[1])
				totalMin := h*60 + m //先转化成分钟
				// 拼装alert 对应的上下班时间
				m, h = (totalMin-a.AlertTime)%60, (totalMin-a.AlertTime)/60
				minute, hour, second := strconv.Itoa(m)+":", strconv.Itoa(h)+":", "00"
				times := hour + minute + second
				AlterOnDutyTime = append(AlterOnDutyTime, timeNowYMD+" "+times)

				// 拼装考勤上下班时间
				m, h = (totalMin+a.DelayTime)%60, (totalMin+a.DelayTime)/60
				if m < 10 && h < 10 {
					minute, hour, second = "0"+strconv.Itoa(m)+":", "0"+strconv.Itoa(h)+":", "00"
				} else if h < 10 {
					minute, hour, second = strconv.Itoa(m)+":", "0"+strconv.Itoa(h)+":", "00"
				} else if m < 10 {
					minute, hour, second = "0"+strconv.Itoa(m)+":", strconv.Itoa(h)+":", "00"
				} else {
					minute, hour, second = strconv.Itoa(m)+":", strconv.Itoa(h)+":", "00"
				}
				times = hour + minute + second

				commutingOnDutyTime = append(commutingOnDutyTime, timeNowYMD+" "+times)
			} else {
				//OffDutyTime = append(OffDutyTime, timeNowYMD+" "+time.CheckTime[l-8:])
				commutingOffDutyTime = append(commutingOffDutyTime, timeNowYMD+" "+string(b))
				AlterOffDutyTime = append(AlterOffDutyTime, timeNowYMD+" "+string(b))
			}

		}
	}
	commutingTime["OnDuty"], commutingTime["OffDuty"] = commutingOnDutyTime, commutingOffDutyTime
	AlterTime["OnDuty"], AlterTime["OffDuty"] = AlterOnDutyTime, AlterOffDutyTime
	//获取到上班时间和提醒打卡时间
	OnDutyTimeList, AlterTimeList := commutingTime["OnDuty"], AlterTime["OnDuty"]
	//获取到不考勤时间
	err = global.GLOAB_DB.Where("attend_group_id", a.GroupId).Find(&restTime).Error
	if err != nil {
		zap.L().Error("根据考勤组获取休息时间失败", zap.Error(err))
		return
	}
	//把时间格式拼装处理一下，拼装成corn定时库spec定时规则能够使用的格式
	minute, hour := "", ""
	for i := 0; i < len(OnDutyTimeList); i++ {
		s := strings.Split(strings.Split(OnDutyTimeList[i], " ")[1], ":")
		hour += s[0] + ","
		minute += s[1] + ","
	}
	hour = hour[:len(hour)-1]
	minute = minute[:len(minute)-1]
	if runtime.GOOS == "windows" {
		AttendSpec = "00 07,24,47 15,17,22 * * ?"
	} else if runtime.GOOS == "linux" {
		AttendSpec = "00 " + minute + " " + hour + " * * ?"
	} else if runtime.GOOS == "darwin" {
		AttendSpec = utils.AttendSpec
	}

	minute = ""
	hour = ""

	for i := 0; i < len(AlterTimeList); i++ {
		s := strings.Split(strings.Split(AlterTimeList[i], " ")[1], ":")
		hour += s[0] + ","
		minute += s[1] + ","
	}
	hour = hour[:len(hour)-1]
	minute = minute[:len(minute)-1]
	if runtime.GOOS == "windows" {
		AlertSpec = "00 07,24,47 15,17,22 * * ?"
	} else if runtime.GOOS == "linux" {
		AlertSpec = "00 " + minute + " " + hour + " * * ?"
	} else if runtime.GOOS == "darwin" {
		AlertSpec = "00 " + minute + " " + hour + " * * ?"
		AlertSpec = utils.AlertSpec
	}
	err = global.GLOAB_DB.Model(&a).Select("is_in_school").Scan(&isInSchool).Error
	if err != nil {
		zap.L().Error("通过考勤组判断是否在学校（加入课表小程序数据失败）", zap.Error(err))
		isInSchool = false
	}
	return
}

// 根据id获取详细的考勤组
// https://open.dingtalk.com/document/orgapp-server/queries-attendance-group-list-details
func (a *DingAttendGroup) GetAttendancesGroupById() (err error) {
	groups, err := a.GetAttendancesGroups(0, 50)
	if err != nil {
		return
	}
	for _, attendGroup := range groups {
		if strconv.Itoa(attendGroup.GroupId) == strconv.Itoa(a.GroupId) {
			a.SelectedClass = make([]struct {
				Setting struct {
					PermitLateMinutes int `json:"permit_late_minutes"`
				} `gorm:"-" json:"setting"`
				Sections []struct {
					Times []struct {
						CheckTime string `json:"check_time"`
						CheckType string `json:"check_type"`
					} `gorm:"-" json:"times"`
				} `gorm:"-" json:"sections"`
			}, 1)
			a.SelectedClass[0].Sections = attendGroup.SelectedClass[0].Sections
			break
		}
	}
	return
}

// 获取考勤组中的部门成员，已经筛掉了不参与考勤的人员
func (a *DingAttendGroup) GetGroupDeptNumber() (DeptUsers map[string][]DingUser, err error) {
	DeptUsers = make(map[string][]DingUser)
	result, err := a.GetAttendancesGroupMemberList("413550622937553255")
	//存储不参与考勤人员，键是用户id，值是用户名
	NotAttendanceUserIdListMap := make(map[string]string)
	DeptAllUserList := make([]DingUser, 0)
	for _, Member := range result {
		if Member.Type == "0" && Member.AtcFlag == "1" { //单个人且不参与考勤
			u := DingUser{
				UserId:    Member.MemberID,
				DingToken: a.DingToken,
			}
			NotAttendanceUser, err := u.GetUserDetailByUserId()
			if err != nil {
				zap.L().Error(fmt.Sprintf("找不到单个人且不参与考勤 的个人信息，跳过%v", u))
				continue
			}
			NotAttendanceUserIdListMap[Member.MemberID] = NotAttendanceUser.Name
		}
	}
	for _, Member := range result {
		DeptAttendanceUserList := make([]DingUser, 0)
		if Member.Type == "1" && Member.AtcFlag == "0" { //部门且参与考勤
			deptId, _ := strconv.Atoi(Member.MemberID)
			d := DingDept{DeptId: deptId, DingToken: DingToken{Token: a.DingToken.Token}}
			DeptAllUserList, _, err = d.GetUserListByDepartmentID(0, 100)
			if err != nil {
				zap.L().Error(fmt.Sprintf("通过部门id:%v获取部门用户列表失败", deptId), zap.Error(err))
				continue
			}
			for _, value := range DeptAllUserList {
				if _, ok := NotAttendanceUserIdListMap[value.UserId]; ok {
					continue
				}
				DeptAttendanceUserList = append(DeptAttendanceUserList, value)
			}
			DeptUsers[Member.MemberID] = DeptAttendanceUserList
		}
	}
	return
}

// 获取考勤组人员（部门id和成员id）https://open.dingtalk.com/document/isvapp-server/batch-query-of-attendance-group-members
func (a *DingAttendGroup) GetAttendancesGroupMemberList(OpUserID string) (R []DingAttendanceGroupMemberList, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/attendance/group/member/list?access_token=" + a.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		OpUserID string `json:"op_user_id"`
		GroupID  int    `json:"group_id"`
	}{
		OpUserID: OpUserID,
		GroupID:  a.GroupId,
	}
	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		DingResponseCommon
		Result struct {
			DingAttendanceGroupMemberList []DingAttendanceGroupMemberList `json:"result"`
		} `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	R = r.Result.DingAttendanceGroupMemberList
	return R, nil
}

// 通过部门id获取部门所有成员user_id（非详细信息） https://open.dingtalk.com/document/isvapp-server/query-the-list-of-department-userids
func (a *DingAttendGroup) GetUserListByDepartmentID(token string, deptId, cursor, size int) (userList []DingUser, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/list?access_token=" + token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		DeptID int `json:"dept_id"`
		Cursor int `json:"cursor"`
		Size   int `json:"size"`
	}{
		DeptID: deptId,
		Cursor: cursor,
		Size:   size,
	}
	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		DingResponseCommon
		Result struct {
			HasMore bool       `json:"has_more"`
			List    []DingUser `json:"list"`
		} `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	return r.Result.List, nil
}

// 更新数据库考勤组
func (a *DingAttendGroup) UpdateAttendGroup() (err error) {
	return global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		var old DingAttendGroup
		err = tx.First(&old, a.GroupId).Error
		if err != nil {
			return err
		}
		AttendGroup := &DingAttendGroup{GroupId: a.GroupId, IsSendFirstPerson: a.IsSendFirstPerson, IsRobotAttendance: a.IsRobotAttendance, IsAlert: a.IsAlert, AlertTime: a.AlertTime}
		//err = tx.Updates(AttendGroup).Error
		//if err != nil {
		//	return err
		//}
		if old.IsRobotAttendance == false && AttendGroup.IsRobotAttendance == true {
			zap.L().Error("更新考勤组开启定时任务")
			//开启定时任务
			taskID, _, err := a.AllDepartAttendByRobot(a.GroupId)
			if err != nil {
				zap.L().Error("开启定时任务AllDepartAttendByRobot()失败", zap.Error(err))
				return err
			}
			AttendGroup.RobotAttendTaskID = int(taskID)
			AttendGroup.IsRobotAttendance = true
			err = tx.Model(&AttendGroup).Updates(AttendGroup).Error
			if err != nil {
				zap.L().Error("mysql更新考勤组定时任务task_id失败")
			}
			zap.L().Info(fmt.Sprintf("开启考勤组考勤定时任务成功！定时任务id为%d", taskID))
			return err
		} else if old.IsRobotAttendance == true && AttendGroup.IsRobotAttendance == false {
			zap.L().Error("更新考勤组关闭定时任务")
			AttendGroup.RobotAttendTaskID = -1
			AttendGroup.IsRobotAttendance = false
			err = tx.Updates(AttendGroup).Error
			if err != nil {
				zap.L().Error("更新考勤组定时任务id为-1失败", zap.Error(err))
			}
			//updates不会更新零值，所以我们使用update单独更新一下
			err = tx.Model(&AttendGroup).Update("is_robot_attendance", false).Error
			if err != nil {
				return err
			}
			zap.L().Info(fmt.Sprintf("关闭cron定时任务，定时任务id为：%v", old.RobotAttendTaskID))
			global.GLOAB_CORN.Remove(cron.EntryID(old.RobotAttendTaskID))
			zap.L().Info("关闭考勤组考勤定时任务成功！")
		}
		return err
	})
}

// 获取数据库考勤组数据
func (a *DingAttendGroup) GetAttendanceGroupListFromMysql(info *request.PageInfo) (DingAttendGroupList []DingAttendGroup, err error) {
	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		limit := info.PageSize
		offset := info.PageSize * (info.Page - 1)
		err = tx.Limit(limit).Offset(offset).Find(&DingAttendGroupList).Error
		if err != nil {
			return err
		}
		return err
	})
	return DingAttendGroupList, err
}

// 判断是否在正确的执行时间
func CronHandle(spec string, curTime *localTime.MySelfTime) (Ok bool) {
	s := strings.Split(spec, " ")
	min := strings.Split(s[1], ",")
	hour := strings.Split(s[2], ",")
	minHour := make([]string, 0)
	if len(min) != len(hour) && len(min) != 1 && len(hour) != 1 {
		zap.L().Error("spec不合法")
		return
	} else if len(min) > 1 && len(hour) > 1 && len(min) == len(hour) {
		zap.L().Info("使用spec一个表达式在多个不同的时间执行，很特殊的一种用法")
		//拼装时间
		for i := 0; i < len(min); i++ {
			minHour = append(minHour, hour[i]+":"+min[i])
		}
		curDate := curTime.Format[0:10]
		for i := 0; i < len(minHour); i++ {
			//拼装成完整的一天的该要运行的时间点
			minHour[i] = curDate + " " + minHour[i] + ":00"
		}
	}

	stamps := make([]int64, 0)
	for i := 0; i < len(minHour); i++ {
		//把时间转化成时间戳
		stamp, err := (&localTime.MySelfTime{}).StringToStamp(minHour[i])
		if err != nil {
			zap.L().Error("把一天的该要运行的时间点 字符串转化成int64时间戳失败", zap.Error(err))
			return
		}
		stamps = append(stamps, stamp)
	}
	OK := false
	//现在把需要运行的时间戳整了出来，不需要运行的直接跳过即可
	for i := 0; i < len(stamps); i++ {
		if curTime.TimeStamp > stamps[i]-1000*60 && curTime.TimeStamp < stamps[i]+1000*60 {
			OK = true
			break
		}
	}

	return OK
}

// 该考勤组进行机器人考勤
func (a *DingAttendGroup) AllDepartAttendByRobot(groupid int) (taskID cron.EntryID, AttendSpec string, err error) {
	g := &DingAttendGroup{}
	err = global.GLOAB_DB.First(g, groupid).Error

	//判断一下是否需要需要课表小程序的数据
	token, err := (&DingToken{}).GetAccessToken()
	if err != nil || token == "" {
		zap.L().Error("从redis中取出token失败", zap.Error(err))
		return
	}
	g.Token = token
	_, _, AttendSpec, _, _, _, err = g.GetCommutingTimeAndSpec()
	if err != nil {
		zap.L().Error("根据考勤组获取上下班时间失败", zap.Error(err))
		return
	}

	zap.L().Info(fmt.Sprintf("根据钉钉考勤组数据拼装spec:%v", AttendSpec))
	AttendTask := func() {
		zap.L().Info(fmt.Sprintf("进入定时任务，定时任务id:%v，对应考勤组:%v", taskID, g.GroupName))
		if int(taskID) != 0 {
			nextTime := global.GLOAB_CORN.Entry(taskID).Next.Format("2006-01-02 15:04:05")
			g.NextTime = nextTime
			err = global.GLOAB_DB.Updates(&g).Error
			if err != nil {
				zap.L().Error("获取定时任务下一次执行时间有误", zap.Error(err))
				return
			}
		}

		token, err = (&DingToken{}).GetAccessToken()
		g.Token = token
		//获取一天上下班的时间
		commutingTimes, _, _, _, restTime, isInSchool, err := g.GetCommutingTimeAndSpec()
		if err != nil {
			zap.L().Error("根据考勤组id获取一天上下班失败失败", zap.Error(err))
			return
		}
		zap.L().Info(fmt.Sprintf("上班时间：%v", commutingTimes["OnDuty"]) + fmt.Sprintf("下班时间：%v", commutingTimes["OffDuty"]))
		//获取当前时间，curTime是自己封装的时间类型，有各种格式的时间
		curTime := &localTime.MySelfTime{}
		err = curTime.GetCurTime(commutingTimes)
		if err != nil {
			zap.L().Error("获取当前时间失败", zap.Error(err))
			return
		}
		//判断当前时间是否需要运行，我们使用的是cron定时器，corn定时器不支持一些不规则的定时，我们此处做一些判断，跳过一些不合法的时间
		if CronHandle(AttendSpec, curTime) == false {
			zap.L().Info("当前时间cron执行，但是不是我们想要的时间，跳过执行")
			return
		}
		//获取考勤组部门成员，已经筛掉了不参与考勤的个人，每个部门都要设置无需考勤的，同一个人如果需要的话，需要在每个考勤组里面设置多次
		//注意一定要放在task里面，这样当纪检部更新了考勤组之后，每次加载人员都是最新的
		deptAttendanceUser, err := g.GetGroupDeptNumber()
		if err != nil {
			zap.L().Error("获取考勤组部门成员(已经筛掉了不参与考勤的个人)失败", zap.Error(err))
			return
		}
		//判断是不是freetime时间
		for _, rest := range restTime {
			if curTime.Week == rest.WeekDay && curTime.Duration == rest.MAE {
				zap.L().Info("本考勤组freetime不再执行考勤")
				return
			}
		}
		for DeptId, _ := range deptAttendanceUser {
			atoi, _ := strconv.Atoi(DeptId)
			DeptDetail := &DingDept{DeptId: atoi}
			DeptDetail.UserList = deptAttendanceUser[DeptId]
			err := DeptDetail.GetDeptByIDFromMysql()
			DeptDetail.Token = token
			if err != nil {
				zap.L().Error(fmt.Sprintf("通过部门id：%s获取部门详情失败，继续执行下一轮循环", DeptId), zap.Error(err))
				continue
			}
			//todo 判断一下此部门是否开启推送考勤
			if DeptDetail.IsRobotAttendance == false || DeptDetail.RobotToken == "" {
				zap.L().Error(fmt.Sprintf("该部门:%s为开启考勤或者机器人robotToken:%s是空，跳过", DeptDetail.Name, DeptDetail.RobotToken))
				continue
			}
			zap.L().Info(fmt.Sprintf("该部门:%s开启考勤,机器人robotToken:%s", DeptDetail.Name, DeptDetail.RobotToken))
			//根据用户id获取用户打卡情况，同时返回了没有考勤数据的同学
			result, _, NotRecordUserIdList, err := DeptDetail.GetAttendanceData(GetUserIdListByUserList(deptAttendanceUser[DeptId]), curTime, commutingTimes["OnDuty"], commutingTimes["OffDuty"], isInSchool)
			if err != nil {
				zap.L().Error("根据部门用户id列表获取用户打卡情况失败", zap.Error(err))
			}
			zap.L().Info(fmt.Sprintf("有考勤记录同学已经处理完成，接下来开始处理没有考勤数据的同学"))
			/*
				获取课表小程序有课的同学
				课表小程序有一个接口，可以获取到大家的有课无课情况，其中参数有
				当前周、高级筛选中的部门，我们找到部门中有课的同学，然后跳过即可
			*/
			//处理没有考勤记录的同学，看看其是否有课，map传递的引用类型
			if isInSchool {
				//调用课表小程序接口，判断没有考勤数据的人是否请假了
				//需要参数：当前周、周几、第几节课，NotRecordUserIdList
				//此处传递的两个参数 NotRecordUserIdList、result 都是引用类型，NotRecordUserIdList处理之后已经不含有课的成员了
				handle := HasCourseHandle(NotRecordUserIdList, curTime.ClassNumber, curTime.StartWeek, curTime.Week, result)
				NotRecordUserIdList = handle
			}
			err = LeaveLateHandle(DeptDetail, NotRecordUserIdList, token, result, curTime, true) // flag为true开启统计信息到redis中
			if err != nil {
				zap.L().Error("处理请假和迟到有误", zap.Error(err))
			}
			zap.L().Info("没有考勤数据的同学已经处理完成")
			//在此处使用bitmap来实现存储功能
			err = BitMapHandle(result, curTime)
			if err != nil {
				zap.L().Error("使用bitmap存储每个人的记录失败", zap.Error(err))
			}
			SendAttendResultHandler(DeptDetail, result, curTime)

			if int(time.Now().Weekday()) == 0 && curTime.Duration == 2 { //周日下午考勤自动发
				DeptDetail.SendFrequencyPrivateLeave(curTime.StartWeek)
				DeptDetail.SendSubSectorPrivateLeave(curTime.StartWeek)
			}

		}

		// 发送整个考勤组的信息

		return
	}
	taskID, err = global.GLOAB_CORN.AddFunc(AttendSpec, AttendTask)
	if err != nil {
		zap.L().Error("启动机器人查考勤定时任务失败", zap.Error(err))
		return
	}
	nextTime := global.GLOAB_CORN.Entry(taskID).Next.Format("2006-01-02 15:04:05")
	g.NextTime = nextTime
	err = global.GLOAB_DB.Updates(&g).Error
	if err != nil {
		zap.L().Error("获取定时任务下一次执行时间有误", zap.Error(err))
		return
	}
	return
}

// AlerdAttent 提醒未打卡的同学考勤
func (a *DingAttendGroup) AlertAttendByRobot() (taskID cron.EntryID, AlertSpec string, err error) {
	//判断一下是否需要需要课表小程序的数据
	token, err := (&DingToken{}).GetAccessToken()
	if err != nil || token == "" {
		zap.L().Error("从redis中取出token失败", zap.Error(err))
		return
	}
	a.Token = token
	_, _, _, AlertSpec, _, _, err = a.GetCommutingTimeAndSpec()
	if err != nil {
		zap.L().Error("根据考勤组获取上下班时间失败", zap.Error(err))
		return
	}
	AlertTask := func() {
		if int(taskID) != 0 {
			nextTime := global.GLOAB_CORN.Entry(taskID).Next.Format("2006-01-02 15:04:05")
			a.NextTime = nextTime
			err = global.GLOAB_DB.Updates(a).Error
			if err != nil {
				zap.L().Error("更新定时任务下一次执行时间有误", zap.Error(err))
				return
			}
		}
		token, err = (&DingToken{}).GetAccessToken()
		a.Token = token
		//获取一天上下班的时间
		_, AlterTime, _, _, restTime, isInSchool, err := a.GetCommutingTimeAndSpec()
		if err != nil {
			zap.L().Error("根据考勤组id获取一天上下班失败失败", zap.Error(err))
			return
		}
		zap.L().Info(fmt.Sprintf("上班时间：%v", AlterTime["OnDuty"]) + fmt.Sprintf("下班时间：%v", AlterTime["OffDuty"]))
		//获取当前时间，curTime是自己封装的时间类型，有各种格式的时间
		curTime := &localTime.MySelfTime{}
		err = curTime.GetCurTime(AlterTime)
		if err != nil {
			zap.L().Error("获取当前时间失败", zap.Error(err))
			return
		}
		//判断当前时间是否需要运行，我们使用的是cron定时器，corn定时器不支持一些不规则的定时，我们此处做一些判断，跳过一些不合法的时间
		if CronHandle(AlertSpec, curTime) == false {
			zap.L().Info("当前时间cron执行，但是不是我们想要的时间，跳过执行")
			return
		}
		//获取考勤组部门成员，已经筛掉了不参与考勤的个人，每个部门都要设置无需考勤的，同一个人如果需要的话，需要在每个考勤组里面设置多次
		//注意一定要放在task里面，这样当纪检部更新了考勤组之后，每次加载人员都是最新的
		deptAttendanceUser, err := a.GetGroupDeptNumber()
		if err != nil {
			zap.L().Error("获取考勤组部门成员(已经筛掉了不参与考勤的个人)失败", zap.Error(err))
			return
		}
		//判断是不是freetime时间
		for _, rest := range restTime {
			if curTime.Week == rest.WeekDay && curTime.Duration == rest.MAE {
				zap.L().Info("本考勤组freetime不再执行考勤")
				return
			}
		}
		for DeptId, _ := range deptAttendanceUser {
			atoi, _ := strconv.Atoi(DeptId)

			DeptDetail := &DingDept{DeptId: atoi}
			err := DeptDetail.GetDeptByIDFromMysql()
			DeptDetail.Token = token
			DeptDetail.UserList = deptAttendanceUser[DeptId]

			if err != nil {
				zap.L().Error(fmt.Sprintf("通过部门id：%s获取部门详情失败，继续执行下一轮循环", DeptId), zap.Error(err))
				continue
			}
			//todo 判断一下此部门是否开启推送考勤
			if DeptDetail.IsRobotAttendance == false || DeptDetail.RobotToken == "" {
				zap.L().Error(fmt.Sprintf("该部门:%s为开启考勤或者机器人robotToken:%s是空，跳过", DeptDetail.Name, DeptDetail.RobotToken))
				continue
			}
			zap.L().Info(fmt.Sprintf("该部门:%s开启考勤,机器人robotToken:%s", DeptDetail.Name, DeptDetail.RobotToken))
			//根据用户id获取用户打卡情况，同时返回了没有考勤数据的同学
			result, _, NotRecordUserIdList, err := DeptDetail.GetAttendanceData(GetUserIdListByUserList(deptAttendanceUser[DeptId]), curTime, AlterTime["OnDuty"], AlterTime["OffDuty"], isInSchool)
			if err != nil {
				zap.L().Error("根据部门用户id列表获取用户打卡情况失败", zap.Error(err))
			}
			zap.L().Info(fmt.Sprintf("有考勤记录同学已经处理完成，接下来开始处理没有考勤数据的同学"))
			/*
				获取课表小程序有课的同学
				课表小程序有一个接口，可以获取到大家的有课无课情况，其中参数有
				当前周、高级筛选中的部门，我们找到部门中有课的同学，然后跳过即可
			*/
			//处理没有考勤记录的同学，看看其是否有课，map传递的引用类型
			if isInSchool {
				//调用课表小程序接口，判断没有考勤数据的人是否请假了
				//需要参数：当前周、周几、第几节课，NotRecordUserIdList
				//此处传递的两个参数 NotRecordUserIdList、result 都是引用类型，NotRecordUserIdList处理之后已经不含有课的成员了
				handle := HasCourseHandle(NotRecordUserIdList, curTime.ClassNumber, curTime.StartWeek, curTime.Week, result)
				NotRecordUserIdList = handle
			}

			err = LeaveLateHandle(DeptDetail, NotRecordUserIdList, token, result, curTime, false)
			if err != nil {
				zap.L().Error("处理请假和迟到有误", zap.Error(err))
			}
			zap.L().Info("没有考勤数据的同学已经处理完成")
			if runtime.GOOS == "linux" {
				p := &ParamChat{
					RobotCode: "dinglyjekzn80ebnlyge",
					UserIds:   NotRecordUserIdList,
					MsgKey:    "sampleText",
					MsgParam:  fmt.Sprintf("还有%v分钟上班，你还没有打卡", a.AlertTime),
				}
				err = (&DingRobot{}).SingleChat(p)
				if err != nil {
					zap.L().Error("发送提醒信息失败", zap.Error(err))
				}
			}
		}
		return
	}
	//添加一个定时任务
	taskID, err = global.GLOAB_CORN.AddFunc(AlertSpec, AlertTask)
	if err != nil {
		zap.L().Error("启动机器人查考勤定时任务失败", zap.Error(err))
		return
	}
	nextTime := global.GLOAB_CORN.Entry(taskID).Next.Format("2006-01-02 15:04:05")
	a.NextTime = nextTime
	err = global.GLOAB_DB.Updates(&a).Error
	if err != nil {
		zap.L().Error("获取定时任务下一次执行时间有误", zap.Error(err))
		return
	}
	return
}

func BitMapHandle(result map[string][]DingAttendance, curTime *localTime.MySelfTime) (err error) {
	//把有课，打卡的，请假的放入sign切片中
	sign := make([]DingAttendance, 0)
	sign = append(sign, result["Normal"]...)
	sign = append(sign, result["Leave"]...)
	sign = append(sign, result["HasCourse"]...)
	for i := 0; i < len(sign); i++ {
		//让每一个用户进行签到
		getWeekSignNum, consecutiveSignNum, err := (&DingUser{UserId: sign[i].UserID}).Sign(curTime.Semester, curTime.StartWeek, curTime.Week, curTime.Duration)
		if err != nil {
			zap.L().Error(fmt.Sprintf("用户:%s 打卡后签到存储redis失败", sign[i].UserName), zap.Error(err))
		} else {
			zap.L().Info(fmt.Sprintf("用户打卡后签到存储redis成功，用户%v，连续签到次数：%v 本周总签到次数：%v", sign[i].UserName, consecutiveSignNum, getWeekSignNum))
		}
	}
	return err
}

func MessageHandle(curTime *localTime.MySelfTime, DeptDetail *DingDept, result map[string][]DingAttendance) (message string) {
	MANCourseNum := ""
	if curTime.Duration == 1 {
		if curTime.ClassNumber == 1 {
			MANCourseNum = "上午第一节"
		} else if curTime.ClassNumber == 2 {
			MANCourseNum = "上午第二节"
		}

	} else if curTime.Duration == 2 {
		if curTime.ClassNumber == 1 {
			MANCourseNum = "下午第一节"
		} else if curTime.ClassNumber == 2 {
			MANCourseNum = "下午第二节"
		}
	} else if curTime.Duration == 3 {
		MANCourseNum = "晚上"
	}
	message = MANCourseNum + DeptDetail.Name + "考勤结果如下:\n"

	for key, DingAttendance := range result {
		if key == "Normal" {
			message += "正常: "
		} else if key == "Late" {
			message += "迟到: "
		} else if key == "Leave" {
			message += "请假: "
		} else if key == "HasCourse" {
			message += "有课: "
		}
		//下面的循环每次统计一个部门的一种情况
		for _, attendance := range DingAttendance {
			if key == "Leave" {
				//我们把请假的信息给存入到redis中
				//我们使用人名作为key，使用请假次数作为value
			}
			message += attendance.UserName + " "
		}
		message += "\n"
	}
	return message
}

func LeaveLateHandle(DeptDetail *DingDept, NotRecordUserIdList []string, token string, result map[string][]DingAttendance, curTime *localTime.MySelfTime, flag bool) (err error) {
	var dl DingLeave
	dl.DingToken.Token = token
	limit, Offset, hasMore := 20, 0, true
	//遍历每一个没有考勤记录的同学
	for i := 0; i < len(NotRecordUserIdList); i++ {
		var u DingUser
		u.DingToken.Token = token
		u.UserId = NotRecordUserIdList[i]
		NotAttendanceUser, err := u.GetUserDetailByUserId()
		if err != nil {
			zap.L().Error(fmt.Sprintf("遍历每一个没有考勤记录也没有课的同学的过程中,通过钉钉用户id:%s获取钉钉用户详情失败", NotRecordUserIdList[i]), zap.Error(err))
			continue
		}
		zap.L().Info(fmt.Sprintf("%s没有考勤数据,没有有课信息，接下来开始获取其请假数据", NotAttendanceUser.Name))
		leaveStatusList := make([]DingLeave, 0)
		hasMore = true
		for hasMore {
			zap.L().Info(fmt.Sprintf("姓名：%v提交请假开始时间%v,提交请假结束时间:%v ，把时间戳转化为可以看懂的时间，开始:%s,结束:%s", NotAttendanceUser.Name, curTime.TimeStamp-10*86400000, curTime.TimeStamp, (&localTime.MySelfTime{}).StampToString(curTime.TimeStamp-10*86400000), curTime.Format))
			leaveStatusListSection, HasMore, err := dl.GetLeaveStatus(curTime.TimeStamp-10*86400000, curTime.TimeStamp, Offset, limit, NotRecordUserIdList[i])
			if err != nil {
				zap.L().Error("获取请假状态失败，跳过继续执行下一条数据", zap.Error(err))
				continue
			}
			leaveStatusList = append(leaveStatusList, leaveStatusListSection...)
			hasMore = HasMore
			if hasMore {
				Offset = Offset + 1
			}
		}
		leave := DingLeave{}

		if len(leaveStatusList) > 0 {
			sort.Slice(leaveStatusList, func(i, j int) bool {
				return leaveStatusList[i].EndTime > leaveStatusList[j].StartTime
			})
			leave = leaveStatusList[0]
			zap.L().Info(fmt.Sprintf("%v获取到了请假数据，只保留最后一条请假记录，请假生效时间:%v,请假结束时间:%v", NotAttendanceUser.Name, (&localTime.MySelfTime{}).StampToString(leave.StartTime), (&localTime.MySelfTime{}).StampToString(leave.EndTime)))
		}

		if leave.StartTime != 0 && leave.StartTime < curTime.TimeStamp && leave.EndTime > curTime.TimeStamp-utils.Delay*1000 {
			result["Leave"] = append(result["Leave"], DingAttendance{TimeResult: "Leave", CheckType: "OnDuty", UserID: NotRecordUserIdList[i], UserName: NotAttendanceUser.Name})
			zap.L().Info(fmt.Sprintf("%s在合法时间段请假，被判定为请假", NotAttendanceUser.Name))
		} else {
			zap.L().Info(fmt.Sprintf("%s未在合法时间段请假，被判定为迟到", NotAttendanceUser.Name))
			result["Late"] = append(result["Late"], DingAttendance{TimeResult: "Late", CheckType: "OnDuty", UserID: NotRecordUserIdList[i], UserName: NotAttendanceUser.Name})
		}
	}
	if flag {
		zap.L().Info(fmt.Sprintf("部门：%v开始统计请假迟到信息到redis中", DeptDetail.Name))
		// 统计所有考勤组中的部门平均请假次数排行
		leaveCount, deptNumbers := len(result["Leave"]), float64(len(DeptDetail.UserList))
		ZsetByAllDeptLeaveAveKey := redis.KeyDeptAveLeave + strconv.Itoa(curTime.StartWeek) + ":所有考勤组中的部门平均请假次数排行"
		preScore := global.GLOBAL_REDIS.ZScore(context.Background(), ZsetByAllDeptLeaveAveKey, DeptDetail.Name).Val()
		score, _ := strconv.ParseFloat(fmt.Sprintf("%.6f", (preScore*deptNumbers+float64(leaveCount))/deptNumbers), 64)
		pipeline := global.GLOBAL_REDIS.TxPipeline()
		if err = pipeline.ZAdd(context.Background(), ZsetByAllDeptLeaveAveKey, &redisZ.Z{
			Score:  score,
			Member: DeptDetail.Name,
		}).Err(); err != nil {
			return err
		}
		// 统计所有考勤组中的部门总数请假次数排行
		ZsetByAllDeptLeaveCountKey := redis.KeyDeptAveLeave + strconv.Itoa(curTime.StartWeek) + ":所有考勤组部门总请假次数"
		preScoreTotal := global.GLOBAL_REDIS.ZScore(context.Background(), ZsetByAllDeptLeaveCountKey, DeptDetail.Name).Val()

		err = pipeline.ZAdd(context.Background(), ZsetByAllDeptLeaveCountKey, &redisZ.Z{
			Score:  preScoreTotal + float64(leaveCount),
			Member: DeptDetail.Name,
		}).Err()
		if err != nil {
			return err
		}

		// 统计该部门总数请假次数排行
		for i := 0; i < len(result["Leave"]); i++ {
			err = pipeline.ZIncrBy(context.Background(), redis.KeyDeptAveLeave+strconv.Itoa(curTime.StartWeek)+":dept:"+DeptDetail.Name+":detail:", 1, result["Leave"][i].UserName).Err()
			if err != nil {
				return err
			}
		}
		// 提交事务
		_, err = pipeline.Exec(context.Background())
		// 命令执行失败，取消提交
		if err != nil {
			zap.L().Error(DeptDetail.Name+"redis请假事务失败", zap.Error(err))
			return err
		}
		pipeline.Close()

		// 统计所有考勤组中的部门平均迟到次数排行
		pipeline = global.GLOBAL_REDIS.TxPipeline()
		lateCount := len(result["Late"])
		ZsetByAllDeptLateAveKey := redis.KeyDeptAveLate + strconv.Itoa(curTime.StartWeek) + ":所有考勤组中的部门平均迟到次数排行"
		preAveLateScore := global.GLOBAL_REDIS.ZScore(context.Background(), ZsetByAllDeptLateAveKey, DeptDetail.Name).Val()
		scoreAveLate, err := strconv.ParseFloat(fmt.Sprintf("%.6f", (preAveLateScore*float64(len(DeptDetail.UserList))+float64(lateCount))/float64(len(DeptDetail.UserList))), 64)
		pipeline.ZAdd(context.Background(), ZsetByAllDeptLateAveKey, &redisZ.Z{
			Score:  scoreAveLate,
			Member: DeptDetail.Name,
		})
		// 统计所有考勤组中的部门总数迟到次数排行
		ZsetByAllDeptLateCountKey := redis.KeyDeptAveLate + strconv.Itoa(curTime.StartWeek) + ":所有考勤组部门总迟到次数"
		preScoreTotal = global.GLOBAL_REDIS.ZScore(context.Background(), ZsetByAllDeptLateCountKey, DeptDetail.Name).Val()
		err = pipeline.ZAdd(context.Background(), ZsetByAllDeptLateCountKey, &redisZ.Z{
			Score:  preScoreTotal + float64(lateCount),
			Member: DeptDetail.Name,
		}).Err()
		if err != nil {
			return err
		}
		// 统计该部门总数请假次数排行
		for i := 0; i < len(result["Late"]); i++ {
			err = global.GLOBAL_REDIS.ZIncrBy(context.Background(), redis.KeyDeptAveLate+strconv.Itoa(curTime.StartWeek)+":dept:"+DeptDetail.Name+":detail:", 1, result["Late"][i].UserName).Err()
			if err != nil {
				return err
			}
		}
		pipeline.Close()

	}
	return
}

func HasCourseHandle(NotRecordUserIdList []string, CourseNumber int, startWeek int, weekday int, result map[string][]DingAttendance) []string {
	if len(NotRecordUserIdList) > 0 {
		ByClass, err := classCourse.GetIsHasCourse(CourseNumber, startWeek, 0, NotRecordUserIdList, weekday)
		if err != nil {
			zap.L().Error("获取当前部门参与考勤的人是否有课失败", zap.Error(err))
		}
		for _, class := range ByClass {
			//找到有课同学的下标，然后在NotRecordUserIdList中把该下标对应的元素移除掉
			for i := 0; i < len(NotRecordUserIdList); i++ {
				if NotRecordUserIdList[i] == class.Userid {
					zap.L().Info(fmt.Sprintf("%v没有打卡记录，查询到其有课，属于正常情况", class.UName))
					result["HasCourse"] = append(result["HasCourse"], DingAttendance{TimeResult: "HasCourse", CheckType: "OnDuty", UserID: NotRecordUserIdList[i], UserName: class.UName})
					NotRecordUserIdList = append(NotRecordUserIdList[:i], NotRecordUserIdList[i+1:]...)
					break
				}
			}
		}
	} else {
		zap.L().Info("该部门全部出勤，不再判断是否有课")
	}
	return NotRecordUserIdList
}

func SendAttendResultHandler(DeptDetail *DingDept, result map[string][]DingAttendance, curTime *localTime.MySelfTime) {
	err := DeptDetail.SendFrequencyLeave(curTime.StartWeek)
	if err != nil {
		zap.L().Error("发送部门每位同学请假详情失败", zap.Error(err))
	}
	err = DeptDetail.SendFrequencyLate(curTime.StartWeek)
	if err != nil {
		zap.L().Error("发送部门每位同学迟到详情失败", zap.Error(err))
	}
	openMessage := MessageHandle(curTime, DeptDetail, result)
	pSend := &ParamCronTask{
		MsgText: &common.MsgText{
			At: common.At{IsAtAll: false},
			Text: common.Text{
				Content: openMessage,
			},
			Msgtype: "text",
		},
		RobotId: DeptDetail.RobotToken,
	}
	zap.L().Info(fmt.Sprintf("正在发送信息，信息参数为%v", pSend))
	if runtime.GOOS == "linux" {
		err := (&DingRobot{RobotId: DeptDetail.RobotToken}).SendMessage(pSend)
		if err != nil {
			zap.L().Error(fmt.Sprintf("发送信息失败，信息参数为%v", pSend), zap.Error(err))
		}
	}

	//将考勤数据发给部门负责人以及管理人员
	var userids []string
	err = global.GLOAB_DB.Table("user_dept").Where("is_responsible = ? and ding_dept_dept_id = ?", true, DeptDetail.DeptId).Select("ding_user_user_id").Find(&userids).Error
	if err != nil || len(userids) == 0 {
		zap.L().Error(fmt.Sprintf("mysql获取部门管理人员失败或者该部门：%s 没有设置管理员", DeptDetail.Name), zap.Error(err))
	} else {
		p := &ParamChat{
			RobotCode: "dinglyjekzn80ebnlyge",
			UserIds:   userids,
			MsgKey:    "sampleText",
			MsgParam:  openMessage,
		}
		if runtime.GOOS == "linux" {
			err = (&DingRobot{}).SingleChat(p)
			if err != nil {
				zap.L().Error("发送至部门负责人失败", zap.Error(err))
			}
		}

	}
}

// SundayAfternoonExec 此函数周日下午执行
func SundayAfternoonExec(startWeek int) {
	r := &DingRobot{RobotId: "aba857cf3ba132581d1a99f3f5c9c5fe2754ffd57a3e7929b6781367b9325e40"}
	// 此函数报告本周的请假情况
	SundayLeaveExec(startWeek, r)
	// 此函数报告本周的迟到情况
	SundayLateExec(startWeek, r)
}

// SundayLeaveExec 此函数报告本周的请假情况
func SundayLeaveExec(startWeek int, r *DingRobot) {
	leaveResult, err := global.GLOBAL_REDIS.ZRevRangeWithScores(context.Background(), redis.KeyDeptAveLeave+strconv.Itoa(startWeek)+":", 0, 100).Result()
	if err != nil {
		zap.L().Info("平均请假次数排行信息获取失败")
		return
	}
	message := "各部门平均请假次数排行如下:\n"
	for i, v := range leaveResult {
		// 获取此部门名称
		deptName := strings.Split(v.Member.(string), "部门名称:")[1]
		// 获取此部门的请假总次数
		deptCount := global.GLOBAL_REDIS.Get(context.Background(), redis.KeyDeptAveLeave+strconv.Itoa(startWeek)+":dept:"+deptName).Val()
		// 获取此部门的平均请假次数
		deptAveCount := v.Score
		message += fmt.Sprintf("%v. %v请假总次数为: %v, 平均请假次数为: %v\n", i+1, deptName, deptCount, deptAveCount)
	}
	pSend := &ParamCronTask{
		MsgText: &common.MsgText{
			At: common.At{IsAtAll: false},
			Text: common.Text{
				Content: message,
			},
			Msgtype: "text",
		},
	}
	if err = r.SendMessage(pSend); err != nil {
		zap.L().Info("机器人发送平均请假次数排行信息失败")
		return
	}

}

// SundayLateExec 此函数报告本周的迟到情况
func SundayLateExec(startWeek int, r *DingRobot) {
	lateResult, err := global.GLOBAL_REDIS.ZRevRangeWithScores(context.Background(), redis.KeyDeptAveLate+strconv.Itoa(startWeek)+":", 0, 100).Result()
	if err != nil {
		zap.L().Info("平均迟到次数排行信息获取失败")
		return
	}
	message := "各部门平均迟到次数排行如下:\n"
	for i, v := range lateResult {
		// 获取此部门名称
		deptName := strings.Split(v.Member.(string), "部门名称:")[1]
		// 获取此部门的迟到总次数
		deptCount := global.GLOBAL_REDIS.Get(context.Background(), redis.KeyDeptAveLate+strconv.Itoa(startWeek)+":dept:"+deptName).Val()
		// 获取此部门的平均请假次数
		deptAveCount := v.Score
		message += fmt.Sprintf("%v. %v迟到总次数为: %v, 平均迟到次数为: %v\n", i+1, deptName, deptCount, deptAveCount)
	}
	// 要发送的信息
	pSend := &ParamCronTask{
		MsgText: &common.MsgText{
			At: common.At{IsAtAll: false},
			Text: common.Text{
				Content: message,
			},
			Msgtype: "text",
		},
	}
	if err = r.SendMessage(pSend); err != nil {
		zap.L().Info("机器人发送平均迟到次数排行信息失败")
		return
	}
}

type Message struct {
	DepartmentName string
	FirstName      string
	Time           int64
}

func GetUserIdListByUserList(UserList []DingUser) (UserIdList []string) {
	for _, val := range UserList {
		UserIdList = append(UserIdList, val.UserId)
	}
	return UserIdList
}

func MessageToString(m *Message) string {
	tran := timestampTran("15:04:05", m.Time)
	return fmt.Sprintf("部门：%s,人员：%s,打卡时间：%v\n", m.DepartmentName, m.FirstName, tran)
}
func timestampTran(format string, t int64) (s string) {
	t = t / 1000
	s = time.Unix(t, 0).Format("2006:01:02 15:04:05")
	if format == "15:04:05" {
		return s[len(s)-8 : len(s)]
	} else if format == "2006:01:02 15:04:05" {
		return s
	} else {
		return s
	}
}
