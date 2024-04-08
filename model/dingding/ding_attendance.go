package dingding

import (
	"crypto/tls"
	"ding/initialize/viper"
	"ding/model/common"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DingAttendance struct {
	UserCheckTime int64  `json:"userCheckTime"` //时间戳实际打卡时间。
	TimeResult    string `json:"timeResult"`    //打卡结果Normal：正常 Early：早退 Late：迟到 SeriousLate：严重迟到 Absenteeism：旷工迟到 NotSigned：未打卡
	CheckType     string `json:"checkType"`     //OnDuty 上班，OffDuty下班
	UserID        string `json:"userId"`
	UserName      string `json:"user_name"`
	DingToken
}

// 获取考勤数据//获取考勤结果（可以根据userid批量查询） https://open.dingtalk.com/document/orgapp/attendance-clock-in-record-is-open
func (a *DingAttendance) GetAttendanceList(userIds []string, CheckDateFrom string, CheckDateTo string) (RecordResult []DingAttendance, err error) {

	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/attendance/listRecord?access_token=" + a.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		CheckDateFrom string   `json:"checkDateFrom"`
		CheckDateTo   string   `json:"checkDateTo"`
		UserIds       []string `json:"userIds"`
	}{
		CheckDateFrom: CheckDateFrom,
		CheckDateTo:   CheckDateTo,
		UserIds:       userIds,
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
		RecordResult []DingAttendance `json:"recordresult"`
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
	return r.RecordResult, nil
}
func (g *DingAttendGroup) SendWeekPaper(semester string, startWeek, weekDay, MNE int) {
	// 获取部门中的每一位开启考勤的用户，该数据从钉钉接口获取
	deptAttendanceUser, err := g.GetGroupDeptNumber()
	for deptId, users := range deptAttendanceUser {
		deptIdInt, _ := strconv.Atoi(deptId)
		dept := (&DingDept{DeptId: deptIdInt})
		err = dept.GetDeptByIDFromMysql()
		if err != nil {
			return
		}
		message := ""
		if !dept.IsAttendanceWeekPaper {
			message = fmt.Sprintf("该部门:%s 未开启考勤周报", dept.Name)
			continue
		} else {
			message = dept.Name + semester + "第" + strconv.Itoa(startWeek) + "周考勤周报如下：\n"
			for i := 0; i < len(users); i++ {
				consecutiveSignNum, err := users[i].GetConsecutiveSignNum(semester, startWeek, weekDay, MNE)
				if err != nil {
					return
				}
				num, err := users[i].GetWeekSignNum(semester, startWeek)
				if err != nil {
					return
				}
				signDetail, err := users[i].GetWeekSignDetail(semester, startWeek)
				if err != nil {
					return
				}
				message += fmt.Sprintf("%v 连续签到次数 : %v 签到总次数：%v 签到详情：%v\n", users[i].Name, consecutiveSignNum, num, signDetail)
			}

		}
		p := &ParamCronTask{
			MsgText:    &common.MsgText{Text: common.Text{Content: message}, At: common.At{}, Msgtype: "text"},
			RepeatTime: "立即发送",
			RobotId:    viper.Conf.MiniProgramConfig.RobotToken,
		}
		err, _ := (&DingRobot{RobotId: p.RobotId}).CronSend(nil, p)
		if err != nil {
			return
		}
	}
}
