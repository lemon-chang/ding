package v2

import (
	"strconv"
	"strings"
)

//func GetAttendances(p *params.ParamGetAttendances) (atts []params.ParamGetAttendances, err error) {
//	atts, err = v2.GetAttendances(p)
//	return
//}

func SplitTime(target string) []int {
	split := strings.Split(target, ":")
	res := make([]int, 0)
	for _, s := range split {
		atoi, _ := strconv.Atoi(s)
		res = append(res, atoi)
	}
	return res
}

/*
target 目标时间段，进行考勤查询
Duration 上午，下午，晚上
FirstTime
*/

func UserIdListSplit(userList []string, k int) (res [][]string) {
	n := len(userList) / k
	for i := 0; i <= n; i++ {
		if i == n {
			//说明到了最后一轮
			res = append(res, userList[k*i:])
		} else {
			res = append(res, userList[i*k:(i+1)*k])
		}
	}
	return res
}

//过滤出来成员id列表
