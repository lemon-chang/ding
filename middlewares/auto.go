package middlewares

import (
	"crypto/tls"
	"ding/response"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

//当在中间件或 handler 中启动新的 Goroutine 时，不能使用原始的上下文，必须使用只读副本。
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		//需要去oss系统进行一下统一的判断认证
		//调用oss的接口，来进行登录认证
		var client *http.Client
		var request *http.Request
		var resp *http.Response
		var body []byte
		//URL := "https://oapi.dingtalk.com/attendance/listRecord?access_token=" + a.DingToken.Token
		URL := "http://121.43.119.224:8890/marchsoft/getUserInfo"
		client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}, Timeout: time.Duration(time.Second * 5)}

		//然后把结构体对象序列化一下
		//然后就可以放入具体的request中的
		request, err := http.NewRequest(http.MethodPost, URL, nil)
		request.Header.Set("Authorization", authHeader)
		if err != nil {
			return
		}
		resp, err = client.Do(request)
		if err != nil {
			return
		}
		zap.L().Info(fmt.Sprintf("发送请求成功，原始resp为:%v", resp))
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
		if err != nil {
			return
		}
		r := struct {
			Code int `json:"code"`
			Data struct {
				UserId string `json:"userid"`
				Name   string `json:"name"`
				Mobile string `json:"mobile"`
			} `json:"data"`
			Msg string `json:"msg"`
		}{}

		//把请求到的结构反序列化到专门接受返回值的对象上面
		err = json.Unmarshal(body, &r)
		if err != nil {
			return
		}
		if r.Code != 0 {
			response.FailWithMessage("登录失效", c)
			c.Abort()
			return
		}
		c.Set("user_id", r.Data.UserId)
		c.Next()
	}
}
