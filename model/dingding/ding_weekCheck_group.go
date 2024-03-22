package dingding

import (
	"crypto/tls"
	"ding/model/common"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type ReportDescribe struct {
	DataList   []ReportOapiVo `json:"data_list"`   //日志列表
	Size       int            `json:"size"`        //分页大小
	NextCursor int64          `json:"next_cursor"` //下一游标
	HasMore    bool           `json:"has_more"`
	DingToken  `gorm:"-"`
}
type ReportOapiVo struct {
	Contents     []Content `json:"contents,omitempty" `     //日志内容
	Remark       string    `json:"remark,omitempty" `       //备注
	TemplateName string    `json:"template_name,omitempty"` //模板名称
	DeptName     string    `json:"dept_name,omitempty"`     //部分名称
	CreatorName  string    `json:"creator_name,omitempty"`  //创建名称
	CreatorId    string    `json:"creator_id,omitempty"`    //创建人用用户userid
	CreateTime   int       `json:"create_time,omitempty"`   //日志创建时间
	Report       string    `json:"report,omitempty"`        //日志id
	ModifiedTime int       `json:"modified_time,omitempty"` //日志修改时间
}

type Content struct {
	Sort  string `json:"sort,omitempty"`  //排序
	Type  string `json:"type,omitempty"`  //日志类型
	Value string `json:"value,omitempty"` //用户填写内容
	Key   string `json:"key,omitempty"`   //模板内容
}

// 周报检测用的用户
type ReportCheckUser struct {
	userid string //
	Name   string //
	Num    int    //规定时间段内检测到了
}
type WeekPaperDetail struct {
	DeptId     int    //部门id
	DeptName   string //部门名称
	TempName   string //模板名称
	UserMap    map[string][]WeekUser
	RobotToken string `json:"robot_token"` //该部门发送消息的robot
}
type WeekUser struct {
	userId string //userid
	Name   string //姓名
	Num    int    //周报数量
}

func (r *ReportDescribe) WeekCheckByRobot() {
	token, err := (&DingToken{}).GetAccessToken()
	if err != nil || token == "" {
		zap.L().Error("从redis中取出token失败", zap.Error(err))
		return
	}
	r.Token = token
	//获取指定时间内全部的日志文件
	//获取这周周报情况，目前只有四期五期写周报
	if err = r.GetAllReportListByDingAPI(); err != nil {
		return
	}
	//将需要周报检测的部门成员全部设置为开启
	//weekUserInit(r.DingToken)

	dingDept := DingDept{}
	//获取需要检测的部门
	WeekCheckDepts, err := dingDept.GetDeptByWeekPaper(1)
	if err != nil {
		zap.L().Error("mysql查询日报部门失败", zap.Error(err))
		return
	}
	//部门中遍历
	for i := 0; i < len(WeekCheckDepts); i++ {
		var WeekCheckUser DingUser
		users, err := WeekCheckUser.GetIsWeekPaperUsersByDeptId(WeekCheckDepts[i].DeptId, 1)
		if err != nil {
			zap.L().Error("mysql查询日报部门失败", zap.Error(err))
			return
		}
		WeekCheckDepts[i].UserList = users
	}
	//筛选出部门中需要周报检测的成员
	weekPaperDetails := filterWeekPaperUser(*r, WeekCheckDepts)
	//周报检测消息进行发送
	sendWeekPaperDetail(weekPaperDetails)
}

func sendWeekPaperDetail(details []WeekPaperDetail) {
	for _, detail := range details {
		if detail.DeptName == "家族4期" {
			detail.RobotToken = "b97f17579f3b510c6cf6cc47bd600b7b823e89525eebfa8401ce5c88840c90a2"
		}
		if detail.DeptName == "家族5期" {
			detail.RobotToken = "b97f17579f3b510c6cf6cc47bd600b7b823e89525eebfa8401ce5c88840c90a2"
		}
		//发送
		sendDeptWeekPaper(detail)
	}
}

func sendDeptWeekPaper(detail WeekPaperDetail) {
	openMessage := handlerMessage(detail, detail.UserMap)
	pSend := &ParamCronTask{
		MsgText: &common.MsgText{
			At: common.At{IsAtAll: false},
			Text: common.Text{
				Content: openMessage,
			},
			Msgtype: "text",
		},
		RobotId: detail.RobotToken,
	}
	err := (&DingRobot{RobotId: pSend.RobotId}).SendMessage(pSend)
	if err != nil {
		zap.L().Error(fmt.Sprintf("发送信息失败，信息参数为%v", pSend), zap.Error(err))
	}
}

// 拼装机器人发送信息
func handlerMessage(detail WeekPaperDetail, userMap map[string][]WeekUser) string {
	openMessage := detail.DeptName + "周报检测如下：\n"
	openMessage += "完成成员："
	for _, user := range userMap["done"] {
		openMessage += user.Name + " "
	}
	openMessage += "\n"
	openMessage += "未完成成员："
	for _, user := range userMap["undone"] {
		openMessage += user.Name + " "
	}
	return openMessage
}

func filterWeekPaperUser(r ReportDescribe, depts []DingDept) (weekPaperDetails []WeekPaperDetail) {
	weekPaperDetails = make([]WeekPaperDetail, 0)
	for _, dept := range depts {
		weekPaperDetail := WeekPaperDetail{}
		weekUserMap := map[string][]WeekUser{}
		for i := 0; i < len(dept.UserList); i++ {
			userId := dept.UserList[i].UserId
			Num := 0
			weekUser := WeekUser{userId: dept.UserList[i].UserId, Name: dept.UserList[i].Name}
			weekPaperDetail.DeptId = dept.DeptId
			weekPaperDetail.DeptName = dept.Name
			for j := 0; j < len(r.DataList); j++ {
				if r.DataList[j].CreatorId == userId {
					Num++
					weekPaperDetail.TempName = r.DataList[j].TemplateName
				}
			}
			weekUser.Num = Num
			if Num != 0 {
				weekUserMap["done"] = append(weekUserMap["done"], weekUser)
			} else {
				weekUserMap["undone"] = append(weekUserMap["undone"], weekUser)
			}
			weekPaperDetail.UserMap = weekUserMap
		}
		weekPaperDetails = append(weekPaperDetails, weekPaperDetail)
	}
	return weekPaperDetails
}

func (r *ReportDescribe) GetAllReportListByDingAPI() error {
	var client *http.Client
	var resp *http.Response
	var respBody []byte
	URL := "https://oapi.dingtalk.com/topapi/report/list?access_token=" + r.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	nowTime := time.Now()
	endTime := nowTime.UnixMilli()
	startTime := nowTime.Add(-5 * time.Hour * 24).UnixMilli()

	var cursor, size int64 = 0, 20
	r.HasMore = true
	ReportOapiVos := make([]ReportOapiVo, 0)
	for r.HasMore {
		boy := struct {
			StartTime int64  `json:"start_time"`
			EndTime   int64  `json:"end_time"`
			UserId    string `json:"userid"`
			Cursor    int64  `json:"cursor"`
			Size      int64  `json:"size"`
		}{
			StartTime: startTime,
			EndTime:   endTime,
			Cursor:    cursor,
			Size:      size,
		}
		//然后把结构体对象序列化一下
		bodymarshal, _ := json.Marshal(&boy)

		//再处理一下
		reqBody := strings.NewReader(string(bodymarshal))
		//然后就可以放入具体的request中的
		request, err := http.NewRequest(http.MethodPost, URL, reqBody)
		if err != nil {
			return errors.New(err.Error())
		}
		resp, err = client.Do(request)
		if err != nil {
			return errors.New(err.Error())
		}
		defer resp.Body.Close()
		respBody, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
		if err != nil {
			return errors.New(err.Error())
		}
		allResp := struct {
			DingResponseCommon
			ReportDescribe `json:"result"`
		}{}
		//把请求到的结构反序列化到专门接受返回值的对象上面
		err = json.Unmarshal(respBody, &allResp)

		if allResp.Errcode != 0 {
			return errors.New(allResp.Errmsg)
		}
		ReportOapiVos = append(ReportOapiVos, allResp.ReportDescribe.DataList...)
		r.HasMore = allResp.ReportDescribe.HasMore
		cursor = allResp.ReportDescribe.NextCursor
	}
	// ReportDescribe中
	r.DataList = ReportOapiVos
	return nil
}
func weekUserInit(token DingToken) {
	//根据部门将其中人员全部设置为需检测周报
	dingDept := DingDept{}
	depts, err := dingDept.GetDeptByWeekPaper(1)
	if err != nil {
		zap.L().Error("mysql查询日报部门失败", zap.Error(err))
		return
	}
	for _, dept := range depts {
		dept.DingToken = token
		users, _, _ := dept.GetUserListByDepartmentID(0, 100)
		for _, user := range users {
			if err = user.InitUserWeekPaper(1); err != nil {
				zap.L().Error("mysql查询日报部门失败", zap.Error(err))
				return
			}
		}
	}
}
