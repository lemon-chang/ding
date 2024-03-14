package localTime

import (
	"ding/model/classCourse"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MySelfTime struct {
	TimeStamp   int64     // 时间戳
	Format      string    // 完整的时间字符串
	Time        time.Time // time类型
	Duration    int       //上午 下午 晚上 1 2 3
	ClassNumber int       //当前是第几课节
	Week        int       //周几
	StartWeek   int       // 课表小程序中的第几周
	Semester    string    // 学期
}

// 根据考勤组判断当前时间（时间戳，字符串，time.Time,上午还是下午（根据考勤组规则制定））
func (t *MySelfTime) GetCurTime(commutingTime map[string][]string) (err error) {
	err = t.GetSemester()
	if err != nil || t.Semester == "" {
		zap.L().Error("获取学期学年失败", zap.Error(err))
		t.Semester = "2023-2024学年第二学期"
	}
	m1 := map[string]int{"Sunday": 7, "Monday": 1, "Tuesday": 2, "Wednesday": 3, "Thursday": 4, "Friday": 5, "Saturday": 6}
	now := time.Now()
	weekEnglish := t.GetWeek(&now)
	//周几
	week := m1[weekEnglish]
	t.Week = week
	startWeek, err := (&classCourse.Calendar{}).GetWeek()
	t.StartWeek = startWeek
	if err != nil {
		zap.L().Error("通过课表小程序获取当前第几周失败", zap.Error(err))
	}
	timeStamp := time.Now()
	//获取到时间戳
	t.TimeStamp = timeStamp.UnixMilli()
	//time.Time转成字符串
	StringCurTime := timeStamp.Format("2006-01-02 15:04:05")
	t.Format = StringCurTime
	//字符串转成时间格式
	CurTime, _ := time.Parse("2006-01-02 15:04:05", StringCurTime)
	t.Time = CurTime
	zap.L().Info(fmt.Sprintf("当前时间的时间戳：%v,time.Time：%v,字符串格式：%s", t.TimeStamp, t.Time, t.Format))
	if commutingTime == nil {
		atoi, _ := strconv.Atoi(strings.Split(strings.Split(t.Format, " ")[1], ":")[0])
		zap.L().Info(fmt.Sprintf("截取到的小时为%v", atoi))
		if atoi < 12 {
			zap.L().Info("小于12，是上午")
			t.Duration = 1
		} else if atoi > 12 && atoi < 18 {
			zap.L().Info("大于12&&小于18，是下午")
			t.Duration = 2
		} else if atoi > 18 {
			zap.L().Info("大于18，是晚上")
			t.Duration = 3
		}
		return
	}

	OnDuty := commutingTime["OnDuty"]
	if len(OnDuty) == 3 {
		AfternoonStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[1])
		//AfternoonEnd, _ := time.Parse("2006-01-02 15:04:05", OffDuty[1])
		EveningStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[2])
		//EveningEnd, _ := time.Parse("2006-01-02 15:04:05", OffDuty[2])
		zap.L().Info(fmt.Sprintf("上午下午时间分界点为：%s 下午晚上时间分界点为：%s", AfternoonStart, EveningStart))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.Before(AfternoonStart) 的值为:%v", CurTime, CurTime.Before(AfternoonStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(AfternoonStart) && CurTime.Before(EveningStart):%v", CurTime, CurTime.After(AfternoonStart) && CurTime.Before(EveningStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(EveningStart) 的值为:%v", CurTime, CurTime.After(EveningStart)))
		if CurTime.Before(AfternoonStart) {
			t.Duration = 1 //上午
		} else if CurTime.After(AfternoonStart) && CurTime.Before(EveningStart) {
			t.Duration = 2 //下午
		} else if CurTime.After(EveningStart) {
			t.Duration = 3
		}
		if t.Duration == 0 {
			zap.L().Info("直接用时间对比，判断现在是上午还是下午失败，我们使用时间字符串，截取到小时，来判断")
			atoi, _ := strconv.Atoi(strings.Split(strings.Split(t.Format, " ")[1], ":")[0])
			zap.L().Info(fmt.Sprintf("截取到的小时为%v", atoi))
			if atoi < 12 {
				zap.L().Info("小于12，是上午")
				t.Duration = 1
			} else if atoi > 12 && atoi < 18 {
				zap.L().Info("大于12&&小于18，是下午")
				t.Duration = 2
			} else if atoi > 18 {
				zap.L().Info("大于18，是晚上")
				t.Duration = 3
			}
		}
		t.ClassNumber = 1 //直接判定成第一节课
	} else if len(OnDuty) == 5 {
		zap.L().Info("len(OnDuty) == 5，第二节课也进行考勤")
		//上午第二节课开始
		MorningSecondClassStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[1])
		//下午第二节课开始
		AfternoonSecondClassStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[3])
		AfternoonStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[2])
		EveningStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[4]) //晚上上班
		//EveningEnd, _ := time.Parse("2006-01-02 15:04:05", OffDuty[2])
		zap.L().Info(fmt.Sprintf("上午下午时间分界点为：%s", AfternoonStart))
		zap.L().Info(fmt.Sprintf("下午晚上时间分界点为：%s", EveningStart))
		zap.L().Info(fmt.Sprintf("上午第一节课和第二节课的分界点为：%v", MorningSecondClassStart))
		zap.L().Info(fmt.Sprintf("下午第一节课和第二节课的分界点为：%v", AfternoonSecondClassStart))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.Before(AfternoonStart) 的值为:%v", CurTime, CurTime.Before(AfternoonStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(AfternoonStart) && CurTime.Before(EveningStart):%v", CurTime, CurTime.After(AfternoonStart) && CurTime.Before(EveningStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(EveningStart) 的值为:%v", CurTime, CurTime.After(EveningStart)))
		t.ClassNumber = 1
		if CurTime.Before(AfternoonStart) {
			zap.L().Info("现在是上午时间")
			t.Duration = 1
			zap.L().Info(fmt.Sprintf("CurTime.After(MorningSecondClassStart)为 %v", CurTime.After(MorningSecondClassStart)))
			if CurTime.After(MorningSecondClassStart) {
				zap.L().Info("CurTime.After(MorningSecondClassStart) 为true,现在是上午第二节课")
				t.ClassNumber = 2
			}
		} else if CurTime.After(AfternoonStart) && CurTime.Before(EveningStart) {
			zap.L().Info("成功判定当前为下午，t.Duration = 2")
			t.Duration = 2
			zap.L().Info(fmt.Sprintf("CurTime为%v,AfternoonSecondClassStart为%v,CurTime.After(AfternoonSecondClassStart)的值为%v", CurTime, AfternoonSecondClassStart, CurTime.After(AfternoonSecondClassStart)))
			if CurTime.After(AfternoonSecondClassStart) {
				zap.L().Info("成功判定当前是下午第二节课")
				t.ClassNumber = 2
			}
		} else if CurTime.After(EveningStart) {
			t.Duration = 3
		}
		if t.Duration == 0 {
			zap.L().Info("直接用时间对比，判断现在是上午还是下午失败，我们使用时间字符串，截取到小时，来判断")
			hour, _ := strconv.Atoi(strings.Split(strings.Split(t.Format, " ")[1], ":")[0])
			zap.L().Info(fmt.Sprintf("截取到的小时为%v", hour))
			if hour < 12 {
				zap.L().Info("小于12，是上午")
				t.Duration = 1

				atoi, _ := strconv.Atoi(OnDuty[2])
				if hour > atoi {
					t.ClassNumber = 2
				}
			} else if hour > 12 && hour < 18 {
				zap.L().Info("大于12&&小于18，是下午")
				t.Duration = 2
				atoi, _ := strconv.Atoi(OnDuty[4])
				if hour > atoi {
					t.ClassNumber = 2
				}
			} else if hour > 18 {
				zap.L().Info("大于18，是晚上")
				t.Duration = 3
			}
		}
	}
	//获取当前具体是第几节课，t.ClassNumber是为了后面调用课表小程序
	if t.Duration == 1 {
		if t.ClassNumber == 1 {
			t.ClassNumber = 1
		} else if t.ClassNumber == 2 {
			t.ClassNumber = 2
		}
	} else if t.Duration == 2 {
		if t.ClassNumber == 1 {
			zap.L().Info("curT.Duration == 2 ,现在是下午，所以我们查第三课考勤")
			t.ClassNumber = 3
		} else if t.ClassNumber == 2 {
			zap.L().Info("curT.Duration == 2 ,现在是下午，所以我们查第四课考勤")
			t.ClassNumber = 4
		}
	} else if t.Duration == 3 {
		zap.L().Info("curT.Duration == 3 ,现在是晚上，所以我们查第五课考勤")
		t.ClassNumber = 5
	}
	// 获取学期

	return
}

func (t *MySelfTime) StringToStamp(s string) (int64, error) {
	if s == "" && t.Format != "" {
		timeT, err := time.ParseInLocation("2006-01-02 15:04:05", t.Format, time.Local)
		if err != nil {
			return 0, errors.New("时间转化成时间戳失败")
		}
		return timeT.Unix() * 1000, err
	}
	timeT, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		return 0, errors.New("时间转化成时间戳失败")
	}
	return timeT.Unix() * 1000, err
}
func (t *MySelfTime) StampToString(s int64) string {
	//如果传递参数是0，且t.TimeStamp != 0 ,默认获取当前时间的时间戳
	if s == 0 && t.TimeStamp != 0 {
		return time.Unix(t.TimeStamp/1000, 0).Format("2006-01-02 15:04:05")
	}
	return time.Unix(s/1000, 0).Format("2006-01-02 15:04:05")

}
func (t *MySelfTime) GetWeek(T *time.Time) string {
	if T != nil {
		return T.Weekday().String()
	}
	return t.Time.Weekday().String()
}

func (t *MySelfTime) GetSemester() (err error) {
	url := "http://jwgl.hist.edu.cn/frame/droplist/getDropLists.action"
	method := "POST"
	payload := strings.NewReader("comboBoxName=StMsXnxqDxDesc&paramValue=&isYXB=0&isCDDW=0&isXQ=0&isDJKSLB=0&isZY=0")
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Add("Origin", "http://jwgl.hist.edu.cn")
	req.Header.Add("Proxy-Connection", "keep-alive")
	req.Header.Add("Referer", "http://jwgl.hist.edu.cn/kbbp/dykb.bjkb.html?menucode=SB03")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Cookie", "JSESSIONID=D9BE73709053CF16D408ECF0967A4429")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNDEzNTUwNjIyOTM3NTUzMjU1IiwidXNlcl9uYW1lIjoi6Zer5L2z6bmPIiwiYXV0aG9yaXR5X2lkIjo4ODgsImV4cCI6MTcxODA3NTEyNiwiaXNzIjoieWpwIn0.QTKq6dTkm2xEf0q2DO09QSdJcd6q6l6mDJ1BH6AUAWI")
	req.Header.Add("Host", "jwgl.hist.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	var r []struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	t.Semester = r[0].Name
	return
}
