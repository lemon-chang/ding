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

type SubscriptionRelationship struct {
	Subscriber string //订阅人
	Subscribee string //被订阅人
}

func (a *SubscriptionRelationship) SubscribeSomeone() {
	global.GLOAB_DB.Create(a)
}
func (a *SubscriptionRelationship) UnsubscribeSomeone() {
	global.GLOAB_DB.Delete(a)
}
