package initialize

func RegularlySendCourses() (err error) {
	//	var p dingding.ParamChat
	//	c := cron.New(cron.WithSeconds())
	//	c.Start()
	//	c.AddFunc("0 0 8 * * ?", func() {
	//		sr, err := (&dingding.SubscriptionRelationship{}).QuerySubscribed()
	//		if err != nil {
	//			return
	//		}
	//		for _, value := range sr {
	//			var userids []string
	//			userids = append(userids, value.Subscriber)
	//			username, _ := (&dingding.DingUser{UserId: value.Subscribee}).GetUserByUserId()
	//			p.UserIds = userids
	//			p.MsgKey = "sampleText"
	//			p.MsgParam = fmt.Sprintf("姓名:%v\n", username)
	//			//获取被订阅人的今日课程情况，并拼接到一起
	//			URL := "http://127.0.0.1:20080/course/getCourseOfWeek"
	//			client := &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
	//				TLSClientConfig: &tls.Config{
	//					InsecureSkipVerify: true,
	//				},
	//			}, Timeout: time.Duration(time.Second * 5)}
	//			b := struct {
	//				userid string `json:"userid"`
	//				week   int    `json:"week"`
	//			}{
	//				userid: value.Subscribee,
	//				week:   10,
	//			}
	//			//然后把结构体对象序列化一下
	//			bodymarshal, err := json.Marshal(&b)
	//			if err != nil {
	//				return
	//			}
	//			//再处理一下
	//			reqBody := strings.NewReader(string(bodymarshal))
	//			request, _ := http.NewRequest("GET", URL, reqBody)
	//			resp, _ := client.Do(request)
	//
	//			//发送被订阅人课程信息给订阅人
	//			token, _ := (&dingding.DingToken{}).GetAccessToken()
	//			_ = (&dingding.DingRobot{DingToken: dingding.DingToken{token}}).ChatSendMessage(&p)
	//		}
	//	})
	//
	return
}
