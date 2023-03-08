package dingding

import (
	"crypto/tls"
	"ding/model/dingding"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type GetUserIdByMobileBody struct {
	Mobile string `json:"mobile"`
}
type ResponseGetUserIdByMobile struct {
	UserId string `json:"userid"`
}

//通过电话号码获取钉钉用户userid
func GetUserIdByMobile(token string, mobile string) (userId string, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/getbymobile?access_token=" + token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := GetUserIdByMobileBody{
		Mobile: mobile,
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
	r := ResponseGetUserIdByMobile{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}

	// 此处举行具体的逻辑判断，然后返回即可

	return r.UserId, nil
}

type GetGetUserDetailByUserIdBody struct {
	UserId string `json:"userid"`
}

type ResponseGetUserDetailByUserId struct {
	dingding.DingResponseCommon
	User dingding.DingUser `json:"result"` //必须大写，不然的话，会被忽略，从而反序列化不上
}

//通过userid获取用户详细信息 https://open.dingtalk.com/document/orgapp-server/query-user-details
func GetUserDetailByUserId(token string, userId string) (user dingding.DingUser, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/get?access_token=" + token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := GetGetUserDetailByUserIdBody{
		UserId: userId,
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
	r := ResponseGetUserDetailByUserId{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}

	// 此处举行具体的逻辑判断，然后返回即可

	return r.User, nil
}
