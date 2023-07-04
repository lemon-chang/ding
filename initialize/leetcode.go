package initialize

import (
	"ding/global"
	"ding/model/dingding"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
)

func SendLeetCode() (err error) {
	//spec:="0 0 0 ? * 1 "
	//spec:="0 0 0 ? * * "
	//开启定时器，定时周日下午6：00(cron定时任务的创建)
	//global.GLOAB_CORN.AddFunc(spec, func() {})
	// 加载数据库部门表，找到需要查力扣的部门（gorm预加载https://gorm.io/zh_CN/docs/preload.html）
	var dept []dingding.DingDept
	err = global.GLOAB_DB.Where("is_leet_code=?", 1).Find(&dept).Error
	if err != nil {
		zap.L().Error("获取需要查询力扣的组织错误", zap.Error(err))
	}
	fmt.Println(dept)
	//遍历部门
	for i := 0; i < len(dept); i++ {

	}
	//遍历某部门的同学，拿到力扣主页地址题目数据

	// 爬取本周数据，并存储
	// 通过课表小程序查找是哪一周，查找上周的数据
	// 对比两周数据
	//
	//err := (&dingding.DingDept{DeptId: }).SendLeetCode(,)
	//if err != nil {
	//	return err
	//}
	//return err

	return
}
func getLeetCodeNum(leetCodeAddress string) {
	url := "https://leetcode.cn/graphql/"
	method := "POST"
	payload := strings.NewReader(`{"query":"\n    query userQuestionProgress($userSlug: String!) {\n  
		userProfileUserQuestionProgress(userSlug: $userSlug) {\n    
		numAcceptedQuestions {\n      difficulty\n      count\n    }\n    
		numFailedQuestions {\n      difficulty\n      count\n    }\n    
		numUntouchedQuestions {\n      difficulty\n      count\n    }\n  }\n}\n    
		","variables":{"userSlug":"` + leetCodeAddress + `"},
		"operationName":"userQuestionProgress"}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Origin", "https://leetcode.cn")
	req.Header.Add("Referer", "https://leetcode.cn/u/mgy001/")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")
	req.Header.Add("baggage", "sentry-environment=production,sentry-release=8Vss9UXY58jLaWvYWwsq8,sentry-transaction=%2Fu%2F%5Busername%5D,sentry-public_key=767ac77cf33a41e7832c778204c98c38,sentry-trace_id=bae3f405963d40deb861b9aad7e6d694,sentry-sample_rate=0.03")
	req.Header.Add("random-uuid", "7e09b8f9-f22e-a184-a6f2-73ab1023d23b")
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"113\", \"Chromium\";v=\"113\", \"Not-A.Brand\";v=\"24\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sentry-trace", "bae3f405963d40deb861b9aad7e6d694-bec854a9af779d74-0")
	req.Header.Add("x-csrftoken", "vXeFshgEi0fvllNmBSlFuUmK6g9wnayGieKJNMnavdcO9DQ4cniPP1u003AS5SG6")
	req.Header.Add("Cookie", "csrftoken=vXeFshgEi0fvllNmBSlFuUmK6g9wnayGieKJNMnavdcO9DQ4cniPP1u003AS5SG6; gr_user_id=54ba0057-ac8a-4e52-84ec-2ad5812054b0; a2873925c34ecbd2_gr_last_sent_cs1=mgy001; __atuvc=1%7C45%2C1%7C46%2C1%7C47; Hm_lvt_fa218a3ff7179639febdb15e372f411c=1677827295; _bl_uid=6mlqth5jbX5sRdfbn24nsyslkdze; Hm_lvt_f0faad39bcf8471e3ab3ef70125152c3=1688461039; _gid=GA1.2.1592313913.1688461039; wechat_state=yC9Xw46I; LEETCODE_SESSION=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJuZXh0X2FmdGVyX29hdXRoIjoiLyIsIl9hdXRoX3VzZXJfaWQiOiI0NzQwNTg1IiwiX2F1dGhfdXNlcl9iYWNrZW5kIjoiZGphbmdvLmNvbnRyaWIuYXV0aC5iYWNrZW5kcy5Nb2RlbEJhY2tlbmQiLCJfYXV0aF91c2VyX2hhc2giOiJiZWU5ZDJhMWJjZmY1NWIzZTU3MDI0Y2Y3ZDA2OTlhNjlmY2MzNjZhNTdhMDE2YjljZDVhZjY4ZmFkZDgwNjdkIiwiaWQiOjQ3NDA1ODUsImVtYWlsIjoiIiwidXNlcm5hbWUiOiJtZ3kwMDEiLCJ1c2VyX3NsdWciOiJtZ3kwMDEiLCJhdmF0YXIiOiJodHRwczovL2Fzc2V0cy5sZWV0Y29kZS5jbi9hbGl5dW4tbGMtdXBsb2FkL3VzZXJzL29qTGpqSDdxdGUvYXZhdGFyXzE2NDk5MjQ0NTgucG5nIiwicGhvbmVfdmVyaWZpZWQiOnRydWUsIl90aW1lc3RhbXAiOjE2ODg0NjEwNTUuNTI4MzA3MiwiZXhwaXJlZF90aW1lXyI6MTY5MTAwMjgwMCwidmVyc2lvbl9rZXlfIjowfQ.nS3yo6QO2sfINedRyEGakLsK4Y3Ssi-1BOLX8bgVBeg; NEW_QUESTION_DETAIL_PAGE_V2=1; Hm_lpvt_f0faad39bcf8471e3ab3ef70125152c3=1688461115; _ga=GA1.1.300765497.1667698342; a2873925c34ecbd2_gr_session_id=7a5bff04-a804-4b40-84c8-e1d210dd5be8; a2873925c34ecbd2_gr_last_sent_sid_with_cs1=7a5bff04-a804-4b40-84c8-e1d210dd5be8; a2873925c34ecbd2_gr_cs1=mgy001; a2873925c34ecbd2_gr_session_id_sent_vst=7a5bff04-a804-4b40-84c8-e1d210dd5be8; _ga_PDVPZYN3CW=GS1.1.1688462978.28.0.1688462981.57.0.0; csrftoken=vXeFshgEi0fvllNmBSlFuUmK6g9wnayGieKJNMnavdcO9DQ4cniPP1u003AS5SG6")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNDEzNTUwNjIyOTM3NTUzMjU1IiwidXNlcl9uYW1lIjoi6Zer5L2z6bmPIiwiYXV0aG9yaXR5X2lkIjo4ODgsImV4cCI6MTcxODA3NTEyNiwiaXNzIjoieWpwIn0.QTKq6dTkm2xEf0q2DO09QSdJcd6q6l6mDJ1BH6AUAWI")
	req.Header.Add("Host", "leetcode.cn")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))

}
