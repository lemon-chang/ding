package dingding

import (
	"bytes"
	"context"
	"crypto/tls"
	"ding/global"
	"ding/initialize/redis"
	"ding/initialize/viper"
	"ding/model/classCourse"
	"ding/model/common"
	"ding/model/common/localTime"
	"ding/model/params"
	"ding/model/params/ding"
	"ding/model/params/ding/leave"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DingDept struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	UserList  []DingUser `gorm:"many2many:user_dept"`
	DeptId    int        `gorm:"primaryKey" json:"dept_id"`
	Deleted   gorm.DeletedAt
	Name      string `json:"name"`
	ParentId  int    `json:"parent_id"`
	DingToken
	IsSendFirstPerson int        `json:"is_send_first_person"` // 0为不推送，1为推送
	RobotToken        string     `json:"robot_token"`
	IsRobotAttendance int        `json:"is_robot_attendance"` //是否
	IsJianShuOrBlog   int        `json:"is_jianshu_or_blog" gorm:"column:is_jianshu_or_blog"`
	IsLeetCode        int        `json:"is_leet_code"`
	IsStudyWeekPaper  int        `json:"is_study_week_paper"` // 学习周报
	ResponsibleUsers  []DingUser `gorm:"-"`
	Children          []DingDept `gorm:"-"`
}
type UserDept struct {
	DingUserUserID string
	DingDeptDeptID string
	IsResponsible  bool
	Deleted        gorm.DeletedAt
}

// 自定义表名建表
func (UserDept) user_dept() string {
	return "user_dept"
}
func (d *DingDept) Insert() (err error) {
	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		err = d.GetDeptDetailByDeptId()
		if err != nil {
			return err
		}
		err = tx.Create(d).Error
		UserIdList, err := d.GetUserIDListByDepartmentID()
		if err != nil {
			return err
		}
		UserList := make([]DingUser, len(UserIdList))
		for i := 0; i < len(UserList); i++ {
			UserList[i].UserId = UserIdList[i]
		}
		err = tx.Model(d).Association("UserList").Replace(UserList)
		return err
	})
	return err
}
func (d *DingDept) UpdateByDingEvent() (err error) {
	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		err = d.GetDeptDetailByDeptId()
		if err != nil {
			return err
		}
		err = tx.Select("parent_id", "name").Updates(d).Error
		UserIdList, err := d.GetUserIDListByDepartmentID()
		if err != nil {
			return err
		}
		UserList := make([]DingUser, len(UserIdList))
		for i := 0; i < len(UserList); i++ {
			UserList[i].UserId = UserIdList[i]
		}
		err = tx.Model(d).Association("UserList").Replace(UserList)
		return err
	})
	return err
}
func (d *DingDept) Delete() (err error) {
	return global.GLOAB_DB.Unscoped().Select(clause.Associations).Delete(d).Error
}

// 获取用户的考勤信息
func (d *DingDept) GetAttendanceData(userids []string, curTime *localTime.MySelfTime, OnDutyTime []string, OffDutyTime []string, isInSchool bool) (result map[string][]DingAttendance, attendanceList []DingAttendance, NotRecordUserIdList []string, err error) {
	result = make(map[string][]DingAttendance, 0)
	attendanceList = make([]DingAttendance, 0)
	a := DingAttendance{DingToken: DingToken{Token: d.Token}}
	if userids != nil || len(userids) != 0 {
		for i := 0; i <= len(userids)/50; i++ {
			var split []string
			if len(userids) <= (i+1)*50 {
				split = userids[i*50:]
			} else {
				split = userids[i*50 : (i+1)*50]
			}
			var list []DingAttendance
			zap.L().Info(fmt.Sprintf("接下来开始获取考勤数据，当前时间为：%v %s", curTime.Duration, curTime.Time))
			if len(OnDutyTime) == 3 {
				if curTime.Duration == 1 {
					zap.L().Info(fmt.Sprintf("获取上午考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, curTime.Format[:10]+" 00:00:00", OnDutyTime[0]))
					list, err = a.GetAttendanceList(split, curTime.Format[:10]+" 00:00:00", OnDutyTime[0])
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, curTime.Format[:10]+" 00:00:00", OnDutyTime[0]), zap.Error(err))
						continue
					}
				} else if curTime.Duration == 2 {
					zap.L().Info(fmt.Sprintf("获取下午考勤数据,userIds:%v,开始时间%s,结束时间：%s ", split, OffDutyTime[0], OnDutyTime[1]))
					list, err = a.GetAttendanceList(split, OffDutyTime[0], OnDutyTime[1])
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[0], OnDutyTime[1]), zap.Error(err))
						continue
					}
				} else if curTime.Duration == 3 {
					zap.L().Info(fmt.Sprintf("获取晚上考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, OffDutyTime[1], OnDutyTime[2]))
					list, err = a.GetAttendanceList(split, OffDutyTime[1], OnDutyTime[2])
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[1], OnDutyTime[2]), zap.Error(err))
						continue
					}
				}
				zap.L().Info(fmt.Sprintf("部门：%s,成功获取%v,考勤数据，具体数据:%v", d.Name, curTime.Duration, list))
			} else if len(OnDutyTime) == 5 {
				//如果是第二节课考勤
				if curTime.Duration == 1 {
					if curTime.ClassNumber == 1 {
						zap.L().Info(fmt.Sprintf("获取上午第一节课考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, curTime.Format[:10]+" 00:00:00", OnDutyTime[0]))
						list, err = a.GetAttendanceList(split, curTime.Format[:10]+" 00:00:00", OnDutyTime[0])
					} else if curTime.ClassNumber == 2 {
						//如果连续的两节课都是没课，第一节课打过卡的不用再打了
						zap.L().Info(fmt.Sprintf("获取上午第二节课考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, curTime.Format[:10]+" 00:00:00", OnDutyTime[1]))
						list, err = a.GetAttendanceList(split, curTime.Format[:10]+" 00:00:00", OnDutyTime[1])
					}
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, curTime.Format[:10]+" 00:00:00", OnDutyTime[0]), zap.Error(err))
						continue
					}
				} else if curTime.Duration == 2 {
					if curTime.ClassNumber == 1 {
						zap.L().Info(fmt.Sprintf("获取下午考勤数据,userIds:%v,开始时间%s,结束时间：%s ", split, OffDutyTime[1], OnDutyTime[2])) //第二个下班时间到第三个下班时间
						list, err = a.GetAttendanceList(split, OffDutyTime[1], OnDutyTime[2])
					} else if curTime.ClassNumber == 2 {
						zap.L().Info(fmt.Sprintf("获取下午考勤数据,userIds:%v,开始时间%s,结束时间：%s ", split, OffDutyTime[1], OnDutyTime[3])) //第二个下班时间到第三个下班时间
						//如果连续的两节课都是没课，第一节课打过卡的不用再打了
						list, err = a.GetAttendanceList(split, OffDutyTime[1], OnDutyTime[3])
					}

					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[0], OnDutyTime[1]), zap.Error(err))
						continue
					}
				} else if curTime.Duration == 3 {
					zap.L().Info(fmt.Sprintf("获取晚上考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, OffDutyTime[3], OnDutyTime[4]))
					list, err = a.GetAttendanceList(split, OffDutyTime[3], OnDutyTime[4])
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[3], OnDutyTime[4]), zap.Error(err))
						continue
					}
				}
			}
			attendanceList = append(attendanceList, list...)
			//处理该部门获取到的考勤记录，只保留上班打卡的记录
		}
		for i := 0; i < len(attendanceList); i++ {
			if attendanceList[i].CheckType == "OffDuty" {
				attendanceList = append(attendanceList[:i], attendanceList[i+1:]...)
				//数组变了，把下标往后移动回去
				i--
			}
		}
		zap.L().Info("只保留上班打卡数据成功")
	}
	HasAttendanceDateUser := make(map[string]int64, 0)
	for i := 0; i < len(attendanceList); i++ {
		u := DingUser{
			DingToken: DingToken{
				Token: d.Token,
			},
			UserId: attendanceList[i].UserID,
		}
		err := u.GetUserDetailByUserId()

		if err != nil {
			zap.L().Error(fmt.Sprintf("考勤数据中的成员id:%s 转化为详细信息失败", attendanceList[i].UserID), zap.Error(err))
			continue
		}
		attendanceList[i].UserName = u.Name //完善考勤记录
		HasAttendanceDateUser[attendanceList[i].UserID] = attendanceList[i].UserCheckTime
	}
	zap.L().Info(fmt.Sprintf("打卡机数据获取完毕，完整数据如下：%v", attendanceList))
	NotRecordUserIdList = make([]string, 0)
	for _, UserId := range userids {
		//找到没有考勤记录的人
		_, ok := HasAttendanceDateUser[UserId]
		if !ok {
			NotRecordUserIdList = append(NotRecordUserIdList, UserId)
		}
	}
	for _, attendance := range attendanceList {
		flag := false
		//查一下课表，有课且打卡的话，判定为有课
		if isInSchool {
			course, _ := classCourse.GetIsHasCourse(curTime.ClassNumber, curTime.StartWeek, 0, []string{attendance.UserID}, curTime.Week)
			for _, Byclass := range course {
				if Byclass.Userid == attendance.UserID {
					result["HasCourse"] = append(result["HasCourse"], attendance)
					flag = true
					break
				}
			}
		}
		if flag == false {
			if attendance.TimeResult == "Normal" {
				result["Normal"] = append(result["Normal"], attendance)
			}
		}
	}
	return
}
func (d *DingDept) SendFrequencyLeave(startWeek int) (err error) {

	key := redis.KeyDeptAveLeave + strconv.Itoa(startWeek) + ":dept:" + d.Name + ":detail:"
	results, err := global.GLOBAL_REDIS.ZRangeWithScores(context.Background(), key, 0, -1).Result()
	if err != nil {
		return
	}
	msg := d.Name + "请假情况如下：\n"
	for i := 0; i < len(results); i++ {
		name := results[i].Member.(string)
		time := int(results[i].Score)
		msg += name + "请假次数：" + strconv.Itoa(time) + "\n"
	}

	p := &ParamCronTask{
		MsgText: &common.MsgText{
			Msgtype: "text",
			Text:    common.Text{Content: msg},
		},
		RepeatTime: "立即发送",
	}
	err = (&DingRobot{RobotId: viper.Conf.MiniProgramConfig.RobotCode}).SendMessage(p)
	if err != nil {
		return err
	}
	return
}

// GetLeaveStatus 获取部门请假状态
func GetLeaveStatus(lea leave.RequestDingLeave) ([]leave.DingLeaveStatus, error) {
	var res []leave.DingLeaveStatus
	// 将json数据编码为字节数组
	var send func(lea leave.RequestDingLeave) error
	send = func(lea leave.RequestDingLeave) error {
		jsonLeave, err := json.Marshal(lea)
		if err != nil {
			fmt.Println("json.Marshal(response) failed:", err)
			return err
		}
		url := fmt.Sprintf("https://oapi.dingtalk.com/topapi/attendance/getleavestatus?%s", "access_token=a9450b2c107d3c67938367cd28dd8825")
		buffer := bytes.NewBuffer(jsonLeave)
		response, err := http.Post(url, "application/json", buffer)
		if err != nil {
			fmt.Println("http.Post(\"https://oapi.dingtalk.com/topapi/attendance/getleavestatus\", \"application/json\", buffer) failed:", err)
			return err
		}
		var dingResp leave.DingResponse
		err = json.NewDecoder(response.Body).Decode(&dingResp)
		if err != nil {
			return err
		}
		res = append(res, *dingResp.Result.LeaveStatus...)
		if dingResp.Result.HasMore {
			lea.Offset += lea.Size
			send(lea)
		}
		return nil
	}
	err := send(lea)
	return res, err
}

// GetDeptLeave 获取部门请假
func (d *DingDept) GetDeptLeave() (map[string]int, map[string]string) { // 周日调用此函数
	userList := make([]string, 0)
	userMap := make(map[string]string, len(d.UserList)) // map[id]姓名 返回消息时要用
	for i, _ := range d.UserList {                      // 拿到部门所有用户id
		userList = append(userList, d.UserList[i].UserId)
		userMap[d.UserList[i].UserId] = d.UserList[i].Name
	}
	// 获取当前时间  0 表示周日
	now := time.Now()
	monday := now.AddDate(0, 0, -int(now.Weekday())+1-7)                                                    // 本周一
	sunday := now.AddDate(0, 0, -int(now.Weekday()))                                                        // 周日调用
	startTime := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.Local).UnixMilli() // 获取本周一00：00:00的时间戳
	endTime := time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 18, 0, 0, 0, time.Local).UnixMilli()  // 获取本周日18：00:00的时间戳
	lea := leave.RequestDingLeave{
		UseridList: strings.Join(userList, ","),
		StartTime:  startTime,
		EndTime:    endTime,
		Offset:     0,
		Size:       20,
	}
	data, err := GetLeaveStatus(lea)
	if err != nil {
		zap.L().Error("GetLeaveStatus(lea) failed", zap.Error(err))
	}
	code := map[string]string{"个人事假": "d4edf257-e581-45f9-b9b9-35755b598952"}
	resMap := map[string]int{} // map[id]个人假次数
	// 统计个人事假请假次数
	for i, _ := range data {
		if data[i].LeaveCode == code["个人事假"] {
			resMap[data[i].UserID]++
		}
	}
	return resMap, userMap
}

// SendFrequencyPrivateLeave 发送个人假次数
func (d *DingDept) SendFrequencyPrivateLeave(startWeek int) error {

	resMap, userMap := d.GetDeptLeave()
	// 从大到小进行排序
	res := []leave.DingUser{}
	for k, v := range resMap {
		tmp := leave.DingUser{
			Id:   k,
			Name: userMap[k],
			Type: map[string]int{"个人事假": v},
		}
		res = append(res, tmp)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Type["个人事假"] > res[j].Type["个人事假"]
	})

	// 发送消息
	msg := d.Name + "第" + strconv.Itoa(startWeek) + "周" + "个人事假统计如下[算账] ：\n"
	for i := 0; i < len(res); i++ {
		msg += res[i].Name + "请假次数：" + strconv.Itoa(res[i].Type["个人事假"]) + "\n"
	}

	p := &ParamCronTask{
		MsgText: &common.MsgText{
			Msgtype: "text",
			Text:    common.Text{Content: msg},
		},
		RepeatTime: "立即发送",
	}
	(&DingRobot{RobotId: "4e1aecbc81c1d673a3817001b960a898e4b4efa61d1080757eb1d683685f0e8e"}).CronSend(nil, p)
	return nil
}

// SendSubSectorPrivateLeave 发送子部门个人请假次数
func (d *DingDept) SendSubSectorPrivateLeave(startWeek int) error {

	dataMap, _ := d.GetDeptLeave()

	deptList, err := d.GetDepartmentListByID()
	if err != nil {
		zap.L().Error(" d.GetDepartmentListByID() failed", zap.Error(err))
	}
	deptLen := map[string]int{}  // 部门人员数
	deptName := map[int]string{} // 子部门id：部门名字
	for i, _ := range deptList {
		deptName[deptList[i].DeptId] = deptList[i].Name
		deptLen[deptName[deptList[i].DeptId]] = 0
	}
	for _, user := range d.UserList {
		for _, v := range user.DeptIdList {
			if _, exit := deptName[v]; exit {
				deptLen[deptName[v]]++
			}
		}
	}

	resMap := map[string]int{} // 子部门请假次数
	for _, user := range d.UserList {
		for _, deptId := range user.DeptIdList { // 遍历用户部门列表
			if _, exit := deptName[deptId]; exit { // 新生只有二个部门
				resMap[deptName[deptId]] += dataMap[user.UserId]
			}
		}
	}
	// 排序 将子部门看成一个用户
	res := []leave.DingUser{}
	for name, count := range resMap {
		tmp := leave.DingUser{Name: name, Type: map[string]int{"请假总次数": count}, AverageLeave: float64(count) / float64(deptLen[name])}
		res = append(res, tmp)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].AverageLeave > res[j].AverageLeave
	})

	// 发送消息
	msg := d.Name + "第" + strconv.Itoa(startWeek) + "周" + "各组统计如下[算账] ：\n"
	for i := 0; i < len(res); i++ {
		msg += fmt.Sprintf("  %s请假总次数为：%d，平均请假次数为：%.2f \n", res[i].Name, res[i].Type["请假总次数"], res[i].AverageLeave)
	}

	p := &ParamCronTask{
		MsgText: &common.MsgText{
			Msgtype: "text",
			Text:    common.Text{Content: msg},
		},
		RepeatTime: "立即发送",
	}
	(&DingRobot{RobotId: "4e1aecbc81c1d673a3817001b960a898e4b4efa61d1080757eb1d683685f0e8e"}).CronSend(nil, p)
	return nil
}

func (d *DingDept) SendFrequencyLate(startWeek int) (err error) {
	//从redis中取数据，封装，调用钉钉接口，发送即可
	key := redis.KeyDeptAveLate + strconv.Itoa(startWeek) + ":dept:" + d.Name + ":detail:"
	results, err := global.GLOBAL_REDIS.ZRangeWithScores(context.Background(), key, 0, -1).Result()
	if err != nil {
		return
	}
	msg := d.Name + "迟到次数如下：\n"
	for i := 0; i < len(results); i++ {
		name := results[i].Member.(string)
		time := int(results[i].Score)
		msg += name + "迟到次数：" + strconv.Itoa(time) + "\n"
	}
	//fmt.Println("发送迟到频率了")
	p := &ParamCronTask{
		MsgText: &common.MsgText{
			Msgtype: "text",
			Text:    common.Text{Content: msg},
		},
		RepeatTime: "立即发送",
	}
	(&DingRobot{RobotId: viper.Conf.MiniProgramConfig.RobotCode}).CronSend(nil, p)
	return nil
}

// 通过部门id获取部门用户idList https://open.dingtalk.com/document/isvapp/query-the-list-of-department-userids
func (d *DingDept) GetUserIDListByDepartmentID() (useridList []string, err error) {
	token, _ := (&DingToken{}).GetAccessToken()
	d.Token = token
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/user/listid?access_token=" + d.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		DeptID int `json:"dept_id"`
	}{
		DeptID: d.DeptId,
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
			UserIdList []string `json:"userid_list"`
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
	useridList = r.Result.UserIdList
	return
}

// 通过部门id获取部门用户详情（无法获取到额外信息） https://open.dingtalk.com/document/isvapp/queries-the-complete-information-of-a-department-user
func (d *DingDept) GetUserListByDepartmentID(cursor, size int) (userList []DingUser, hasMore bool, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/list?access_token=" + d.DingToken.Token
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
		DeptID: d.DeptId,
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
		return nil, false, errors.New(r.Errmsg)
	}
	userList = r.Result.List

	// 此处举行具体的逻辑判断，然后返回即可
	return r.Result.List, r.Result.HasMore, nil
}

func (d *DingDept) GetUserListByIdList() (userList []DingUser, err error) {
	userList = make([]DingUser, len(userList))
	UserIdList, err := d.GetUserIDListByDepartmentID()
	if err != nil {
		return
	}
	for i := 0; i < len(UserIdList); i++ {
		user := &DingUser{UserId: UserIdList[i], DingToken: DingToken{Token: d.Token}}
		err = user.GetUserDetailByUserId()
		if err != nil {
			zap.L().Error("获取用户详情失败", zap.Error(err))
		}
		userList = append(userList, *user)
	}
	for i := 0; i < len(userList); i++ {
		for j := 0; j < len(userList[i].ExtAttrs); j++ {
			if userList[i].ExtAttrs[j].Code == "1263467522" {
				userList[i].JianshuAddr = userList[i].ExtAttrs[j].Value.Text
			} else if userList[i].ExtAttrs[j].Code == "1263534303" {
				userList[i].BlogAddr = userList[i].ExtAttrs[j].Value.Text
			} else if userList[i].ExtAttrs[j].Code == "1263581295" {
				userList[i].LeetcodeAddr = userList[i].ExtAttrs[j].Value.Text
			}
		}
	}
	return
}

// 两个数组取差集
func DiffArray(a []DingDept, b []DingDept) []DingDept {
	var diffArray []DingDept
	temp := map[int]struct{}{}

	for _, val := range b {
		if _, ok := temp[val.DeptId]; !ok {
			temp[val.DeptId] = struct{}{}
		}
	}

	for _, val := range a {
		if _, ok := temp[val.DeptId]; !ok {
			diffArray = append(diffArray, val)
		}
	}

	return diffArray
}
func DiffSliceDept(a []DingDept, b []DingDept) []DingDept {
	var diffArray []DingDept
	temp := map[int]struct{}{}

	for _, val := range b {
		if _, ok := temp[val.DeptId]; !ok {
			temp[val.DeptId] = struct{}{}
		}
	}

	for _, val := range a {
		if _, ok := temp[val.DeptId]; !ok {
			diffArray = append(diffArray, val)
		}
	}

	return diffArray
}
func DiffSliceUser(a []DingUser, b []DingUser) []DingUser {
	var diffArray []DingUser
	temp := map[string]struct{}{}

	for _, val := range b {
		if _, ok := temp[val.UserId]; !ok {
			temp[val.UserId] = struct{}{}
		}
	}

	for _, val := range a {
		if _, ok := temp[val.UserId]; !ok {
			diffArray = append(diffArray, val)
		}
	}

	return diffArray
}

// 递归查询部门并存储到数据库中
func (d *DingDept) ImportDeptData() (DepartmentList []DingDept, err error) {
	var oldDept []DingDept
	err = global.GLOAB_DB.Find(&oldDept).Error
	if err != nil {
		return
	}
	var dfs func(string, int) (err error)
	dfs = func(token string, id int) (err error) {
		var client *http.Client
		var request *http.Request
		var resp *http.Response
		var body []byte
		URL := "https://oapi.dingtalk.com/topapi/v2/department/listsub?access_token=" + token
		client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}, Timeout: time.Duration(time.Second * 5)}
		//此处是post请求的请求题，我们先初始化一个对象
		b := struct {
			DeptID int `json:"dept_id"`
		}{
			DeptID: id,
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
			Result []DingDept `json:"result"`
		}{}
		//把请求到的结构反序列化到专门接受返回值的对象上面
		err = json.Unmarshal(body, &r)
		if err != nil {
			return
		}
		if r.Errcode != 0 {
			return errors.New(r.Errmsg)
		}
		// 此处举行具体的逻辑判断，然后返回即可
		subDepartments := r.Result
		DepartmentList = append(DepartmentList, subDepartments...)
		if len(subDepartments) > 0 {
			for i := range subDepartments {
				departmentList := make([]DingDept, 0)
				dfs(token, subDepartments[i].DeptId)
				if err != nil {
					return
				}
				DepartmentList = append(DepartmentList, departmentList...)
			}
		}
		return
	}
	err = dfs(d.DingToken.Token, 1)
	if err != nil {
		return
	}
	//取差集查看一下那些部门已经不在来了，进行软删除
	err = global.GLOAB_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "dept_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "parent_id"}),
	}).Create(&DepartmentList).Error
	//找到不存在的部门进行软删除,同时删除其关系
	Deleted := DiffSliceDept(oldDept, DepartmentList)
	if Deleted != nil {
		err = global.GLOAB_DB.Select(clause.Associations).Delete(&Deleted).Error
	}

	//根据部门id存储一下部门用户
	for i := 0; i < len(DepartmentList); i++ {
		d.DeptId = DepartmentList[i].DeptId
		newUserList, err := d.GetUserListByIdList()
		if err != nil {
			zap.L().Error("获取部门用户详情失败", zap.Error(err))
		}

		//查到用户后，同步到数据库里面，把不在组织架构里面直接删除掉
		//先查一下老的
		var oldUserList []DingUser
		err = global.GLOAB_DB.Model(&DingDept{DeptId: DepartmentList[i].DeptId}).Association("UserList").Find(&oldUserList)
		if err != nil {
			return nil, err
		}
		//取差集找到需要删除的名单
		userDeleted := DiffSliceUser(oldUserList, newUserList)
		if userDeleted != nil {
			err = global.GLOAB_DB.Select(clause.Associations).Delete(&userDeleted).Error
		}

		err = global.GLOAB_DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "userid"}},
			DoUpdates: clause.AssignmentColumns([]string{"jianshu_addr", "leetcode_addr", "blog_addr"}),
		}).Create(&newUserList).Error
		//更新用户部门关系，更新的原理是：先把之前该部门的关系全部删除，然后重新添加
		err = global.GLOAB_DB.Model(&DepartmentList[i]).Association("UserList").Replace(newUserList)
	}
	return
}

// 根据id获取子部门列表详情
func (d *DingDept) GetDepartmentListByID() (subDepartments []DingDept, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/department/listsub?access_token=" + d.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		DeptID int `json:"dept_id"`
	}{
		DeptID: d.DeptId,
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
		Result []DingDept `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, errors.New("token有误，尝试输入新token")
	}
	// 此处举行具体的逻辑判断，然后返回即可
	subDepartments = r.Result
	return subDepartments, nil
}

// 根据id获取子部门列表详情（从数据库查）
func (d *DingDept) GetDepartmentListByID2() (subDepartments []DingDept, err error) {
	err = global.GLOAB_DB.Where("parent_id = ?", d.DeptId).Find(&subDepartments).Error
	return
}
func (d *DingDept) GetDeptByIDFromMysql() (err error) {
	err = global.GLOAB_DB.First(d, d.DeptId).Error
	return
}

// 获得需要进行周报检测的部门
func (d *DingDept) GetDeptByWeekPaper(num int) (depts []DingDept, err error) {
	err = global.GLOAB_DB.Where("is_week_paper = ?", num).Find(&depts).Error
	return
}
func (d *DingDept) GetDeptByListFromMysql(p *params.ParamGetDeptListFromMysql) (deptList []DingDept, total int64, err error) {
	limit := p.PageSize
	offset := p.PageSize * (p.Page - 1)
	err = global.GLOAB_DB.Limit(limit).Offset(offset).Preload("UserList", "user_dept.is_responsible = 1").Find(&deptList).Error
	if err != nil {
		zap.L().Error("查询部门列表失败", zap.Error(err))
	}
	err = global.GLOAB_DB.Model(&DingDept{}).Count(&total).Error
	if err != nil {
		zap.L().Error("查询部门列表失败", zap.Error(err))
	}
	return
}
func (d *DingDept) GetDepartmentRecursively() (list []DingDept, total int64, err error) {

	db := global.GLOAB_DB.Model(&DingDept{})
	err = db.Where("parent_id = ?", 1).Count(&total).Error
	var department []DingDept
	err = db.Where("parent_id = ?", 1).Find(&department).Error

	if len(department) > 0 {
		for i := range department {
			//err = global.GVA_REDIS.HSet("deptCache", strconv.Itoa(int(department[i].Children[i].ID)), "").Err()
			err = d.findChildrenDepartment(&department[i])
		}
	}

	return department, total, err
}
func (d *DingDept) findChildrenDepartment(department *DingDept) (err error) {
	err = global.GLOAB_DB.Where("parent_id = ?", department.DeptId).Find(&department.Children).Order("sort").Error

	if len(department.Children) > 0 {
		for k := range department.Children {
			err = d.findChildrenDepartment(&department.Children[k])
		}
	}
	return err
}

// 查看部门推送情况开启推送情况
func (d *DingDept) SendFirstPerson(cursor, size int) {
	var depts []DingDept
	global.GLOAB_DB.Select("Name").Find(&depts)
}

// 通过部门id获取部门详细信息（取钉钉接口）  https://open.dingtalk.com/document/isvapp-server/industry-address-book-api-for-obtaining-department-information
func (d *DingDept) GetDeptDetailByDeptId() (err error) {
	token, _ := (&DingToken{}).GetAccessToken()
	d.Token = token
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/department/get?access_token=" + d.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		DeptID int `json:"dept_id"`
	}{
		DeptID: d.DeptId,
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
		Dept DingDept `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	*d = r.Dept // 此处不可使用 d = &(r.Dept)
	return
}

// 更新部门信息
func (d *DingDept) UpdateDept(p *ding.ParamUpdateDept) (err error) {
	dept := &DingDept{DeptId: p.DeptID, IsSendFirstPerson: p.IsSendFirstPerson, IsRobotAttendance: p.IsRobotAttendance, RobotToken: p.RobotToken, IsJianShuOrBlog: p.IsJianshuOrBlog, IsLeetCode: p.IsLeetCode}
	// 使用select更新选中的字段
	err = global.GLOAB_DB.Select("IsSendFirstPerson", "IsRobotAttendance", "RobotToken", "IsJianShuOrBlog", "IsLeetCode", "ResponsibleUserIds").Updates(dept).Error
	if len(p.ResponsibleUserIds) > 0 {
		err = global.GLOAB_DB.Table("user_dept").Where("ding_dept_dept_id = ?", p.DeptID).Update("is_responsible", false).Error
		err = global.GLOAB_DB.Table("user_dept").Where("ding_user_user_id IN ? AND ding_dept_dept_id = ?", p.ResponsibleUserIds, p.DeptID).Update("is_responsible", true).Error
	}
	return err
}

// 查询部门周报检测状态
func (d *DingDept) GetDeptWeekCheckStatus() (depts []DingDept, err error) {
	err = global.GLOAB_DB.Model(d).Select("dept_id", "name", "is_week_paper").Find(&depts).Error
	return
}

// 更新部门周报检测状态
func (d *DingDept) UpdateDeptWeekCheckStatus() (err error) {
	err = global.GLOAB_DB.Model(d).Where("dept_id = ?", d.DeptId).Update("is_week_paper", d.IsStudyWeekPaper).Error
	return
}
