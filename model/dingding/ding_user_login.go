package dingding

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"ding/global"
	"ding/utils"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Code struct {
	AuthCode string `json:"authCode"`
}

func (d *DingUser) Login() (user *DingUser, err error) {
	user = &DingUser{
		Mobile:   d.Mobile,
		Password: d.Password,
	}
	//判断该用户是否存在
	err = global.GLOAB_DB.Model(DingUser{}).Where("mobile", d.Mobile).First(user).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	//判断密码是否正确
	if user.Password != d.Password {
		return nil, errors.New("密码错误")
	}
	//此处的Login函数传递的是一个指针类型的数据
	opassword := user.Password //此处是用户输入的密码，不一定是对的
	err = global.GLOAB_DB.Where(&DingUser{Mobile: user.Mobile}).Preload("Authorities").Preload("Authority").First(user).Error
	if err != nil {
		zap.L().Error("登录时查询数据库失败", zap.Error(err))
		return
	}
	//如果到了这里还没有结束的话，那就说明该用户至少是存在的，于是我们解析一下密码
	//password := encryptPassword(opassword)
	password := opassword
	//拿到解析后的密码，我们看看是否正确
	if password != user.Password {
		return nil, errors.New("密码错误")
	}
	d.UserAuthorityDefaultRouter(user)
	//如果能到这里的话，那就登录成功了
	return

}

func (code *Code) SweepLogin(c *gin.Context) (user DingUser, err error) {
	accessToken, err := GetAccessToken(code)
	if err != nil {
		return
	}
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://api.dingtalk.com/v1.0/contact/users/me"
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	request, err = http.NewRequest("GET", URL, nil)
	if err != nil {
		return
	}
	// Set the header parameter
	request.Header.Set("x-acs-dingtalk-access-token", accessToken)
	request.Header.Set("Content-Type", "application/json")
	// Send the request
	resp, err = client.Do(request)
	if err != nil {
		// Handle error
	}
	defer resp.Body.Close()
	// Read the response body
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// Handle error
	}
	r := struct {
		AvatarURL string `json:"avatarUrl"`
		Mobile    string `json:"mobile"`
		Nick      string `json:"nick"`
		OpenID    string `json:"openId"`
		StateCode string `json:"stateCode"`
		UnionID   string `json:"unionId"` //不同应用中用户唯一ID
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	err = global.GLOAB_DB.Where("mobile = ?", r.Mobile).Preload("Authorities").Preload("Authority").First(&user).Error
	if err != nil {
		return
	}
	return
}
func GetAccessToken(code *Code) (accessToken string, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://api.dingtalk.com/v1.0/oauth2/userAccessToken"
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	appKey, _ := global.GLOBAL_REDIS.Get(context.Background(), utils.AppKey).Result()
	appSecret, _ := global.GLOBAL_REDIS.Get(context.Background(), utils.AppSecret).Result()
	b := struct {
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		Code         string `json:"code"`
		GrantType    string `json:"grantType"`
	}{
		ClientID:     appKey,
		ClientSecret: appSecret,
		Code:         code.AuthCode,
		GrantType:    "authorization_code",
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
	request.Header.Set("Content-Type", "application/json")
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
		AccessToken  string `json:"accessToken"`
		Code         string `json:"code"`
		ExpireIn     int64  `json:"expireIn"`
		Message      string `json:"message"`
		RefreshToken string `json:"refreshToken"`
		Requestid    string `json:"requestid"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Code != "" {
		return "", errors.New(r.Message)
	}
	return r.AccessToken, nil
}
func encryptPassword(oPassword string) string {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum([]byte(oPassword)))
}
