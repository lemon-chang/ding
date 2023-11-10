package dingding

import (
	"crypto/tls"
	"ding/global"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type DingLeave struct {
	DurationUnit string `json:"duration_unit"` //请假单位，小时或者天
	EndTime      int64  `json:"end_time"`
	StartTime    int64  `json:"start_time"`
	Userid       string `json:"userid"`
	UserName     string `json:"user_name"`
	DingToken
}
type SubscriptionRelationship struct {
	Subscriber   string //订阅人
	Subscribee   string //被订阅人
	IsCurriculum bool   //是否订阅课表
}

type DingUserRequest struct {
	UseridList []int `json:"userid_list"` // 带查询用户的id列表，每次最多100个
	StartTime  int   `json:"start_time"`  // 开始时间 ，Unix时间戳，支持最多180天的查询。
	EndTime    int   `json:"end_time"`    // 结束时间，Unix时间戳，支持最多180天的查询。
	Offset     int   `json:"offset"`      // 支持分页查询，与size参数同时设置时才生效，此参数代表偏移量，偏移量从0开始。
	Size       int   `json:"size"`        // 支持分页查询，与offset参数同时设置时才生效，此参数代表分页大小，最大20。
}

// DingUserLeaveState 请假状态
type DingUserLeaveState struct {
	DurationUnit    string `json:"duration_unit"`    // 请假单位：percent_day:天  percent_hour:小时
	DurationPercent int    `json:"duration_percent"` // 假期时长*100，例如用户请假时长为1天，该值就等于100。
	LeaveCode       string `json:"leave_code"`       // 请假类型 个人事假：d4edf257-e581-45f9-b9b9-35755b598952  非个人事假：baf811bc-3daa-4988-9604-d68ec1edaf50  病假：a7ffa2e6-872a-498d-aca7-4554c56fbb52
	EndTime         int64  `json:"end_time"`         // 请假结束时间，Unix时间戳。
	StartTime       int64  `json:"start_time"`       //请假开始时间，Unix时间戳。
}

// DingLeaveStatusList 请假列表
type DingLeaveStatusList struct {
	HasMore            bool   `json:"has_more"` // 是否有更多数据
	Success            bool   `json:"success"`  // 请求是否成功
	RequestId          string // 请求ID
	DingUserLeaveState *[]DingLeaveStatusList
}

func (a *DingLeave) GetLeaveStatus(StartTime, EndTime int64, Offset, Size int, UseridList string) (leaveStatus []DingLeave, hasMore bool, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/attendance/getleavestatus?access_token=" + a.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		EndTime    int64  `json:"end_time"`
		StartTime  int64  `json:"start_time"`
		Offset     int    `json:"offset"`
		Size       int    `json:"size"`
		UseridList string `json:"userid_list"`
	}{
		EndTime:    EndTime,
		StartTime:  StartTime,
		Offset:     Offset,
		Size:       Size,
		UseridList: UseridList,
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
			HasMore   bool        `json:"has_more"`
			DingLeave []DingLeave `json:"leave_status"`
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
	hasMore = r.Result.HasMore
	// 此处举行具体的逻辑判断，然后返回即可

	return r.Result.DingLeave, hasMore, err
}

func (a *SubscriptionRelationship) SubscribeSomeone() (err error) {
	//获取请假人姓名
	user := DingUser{}
	err = global.GLOAB_DB.Where("user_id = ?", a.Subscriber).First(&user).Error
	err = global.GLOAB_DB.Where("user_id = ?", a.Subscribee).First(&user).Error
	if err != nil {
		return
	}
	err = global.GLOAB_DB.Create(a).Error
	return
}
func (a *SubscriptionRelationship) UnsubscribeSomeone() (err error) {
	sr := SubscriptionRelationship{}
	err = global.GLOAB_DB.Where("subscriber = ?", a.Subscriber).First(&sr).Error
	err = global.GLOAB_DB.Where("subscribee = ?", a.Subscribee).First(&sr).Error
	if err != nil {
		return
	}
	err = global.GLOAB_DB.Where("subscriber = ? And subscribee = ?", a.Subscriber, a.Subscribee).Delete(a).Error
	return
}

func (a *SubscriptionRelationship) QuerySubscribed() (sr []SubscriptionRelationship, err error) {
	err = global.GLOAB_DB.Where("is_curriculum = ?", true).Find(&sr).Error
	if err != nil {
		return
	}
	return
}
