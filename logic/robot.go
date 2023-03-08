package logic

import (
	"ding/dao/mysql"
	"ding/global"
	"ding/model/common"
	"ding/model/dingding"
	"ding/model/params"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

func HandleSpec(p *dingding.ParamCronSend) (spec, detailTimeForUser string, err error) {
	spec = ""
	detailTimeForUser = ""
	n := len(p.DetailTime)
	if p.RepeateTime == "仅发送一次" {
		second := p.DetailTime[n-2:]
		minute := p.DetailTime[n-5 : n-3]
		hour := p.DetailTime[n-8 : n-6]
		//year := p.DetailTime[:4]
		month := p.DetailTime[5:7]
		day := p.DetailTime[8:10]
		week := "?" //问号代表放弃周
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
		detailTimeForUser = "仅在" + p.DetailTime + "发送一次"
	}
	if string([]rune(p.RepeateTime)[0:3]) == "周重复" {
		M := map[string]string{"0": "周日", "1": "周一", "2": "周二", "3": "周三", "4": "周四", "5": "周五", "6": "周六"}
		detailTimeForUser = "周重复 ："
		weeks := strings.Split(p.RepeateTime, "/")[1:]
		week := ""
		for i := 0; i < len(weeks); i++ {
			detailTimeForUser += M[weeks[i]]
			week += weeks[i] + ","
		}
		week = week[0 : len(week)-1]
		HMS := strings.Split(p.DetailTime, ":")
		second := HMS[2]
		minute := HMS[1]
		hour := HMS[0]
		month := "*" //每个月的每个星期都发送
		day := "?"   //选了星期就要放弃具体的某一天
		detailTimeForUser += hour + "：" + minute + "：" + second
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
	}

	if string([]rune(p.RepeateTime)[0:3]) == "月重复" {
		var daymap map[int]string
		daymap = make(map[int]string)
		for i := 1; i <= 31; i++ {
			daymap[i] += strconv.Itoa(i) + "号"
		}
		//字符串数组
		days := strings.Split(p.RepeateTime, "/")[1:]
		detailTimeForUser = "月重复 ："
		day := ""
		for i := 0; i < len(days); i++ {
			atoi, _ := strconv.Atoi(days[i])
			detailTimeForUser += daymap[atoi]
			day += days[i] + ","
		}
		day = day[0 : len(day)-1]
		HMS := strings.Split(p.DetailTime, ":")
		second := HMS[2]
		minute := HMS[1]
		hour := HMS[0]
		month := "*" //每个月的每个星期都发送
		week := "?"
		detailTimeForUser += hour + ":" + minute + ":" + second
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
	}

	if spec == "" || detailTimeForUser == "" {
		return spec, detailTimeForUser, errors.New("cron定时规则转化错误")
	}
	return spec, detailTimeForUser, nil
}

//func Send(c *gin.Context, p *params.ParamCronSend) (err error, task dingding.Task) {
//	spec, detailTimeForUser, err := HandleSpec(p)
//	tid := "0"
//	uid, err := global.GetCurrentID(c)
//	if err != nil {
//		uid = 0
//	}
//	user, err := (&dingding.DingUser{Model: gorm.Model{
//		ID: uid,
//	}}).GetUserByIDOrUserID()
//	if err != nil {
//		user = dingding.DingUser{}
//	}
//	//先来判断一下此用户是否有这个小机器人
//	robot, err := mysql.GetRobotByRobotId(p.RobotId)
//	if err != nil {
//		zap.L().Error("通过机器人的robot_id获取机器人失败", zap.Error(err))
//
//	}
//	//到了这里就说明这个用户有这个小机器人
//	d := dingding.DingRobot{
//		RobotId: p.RobotId,
//		Secret:  robot.Secret,
//	}
//	//crontab := cron.New(cron.WithSeconds()) //精确到秒
//	//spec := "* 30 22 * * ?" //cron表达式，每五秒一次
//	if p.MsgText.Msgtype == "text" {
//		//err = AtMobiles(p, &robot, &user)
//		//if err != nil {
//		//	zap.L().Error("通过人名查询电话号码失败", zap.Error(err))
//		//	return
//		//}
//		if (p.RepeateTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
//			zap.L().Info("进入即时发送消息模式")
//			err := d.SendMessage(p)
//			if err != nil {
//				return err, dingding.Task{}
//			} else {
//				zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", user.Name, robot.Title))
//			}
//			//定时任务
//			task = dingding.Task{
//				TaskID:            tid,
//				TaskName:          p.TaskName,
//				DingUserID:        user.ID,
//				RobotId:           robot.RobotId,
//				RobotName:         robot.Title,
//				RobotSecret:       robot.Secret,
//				DetailTimeForUser: detailTimeForUser, //给用户看的
//				Spec:              spec,              //cron后端定时规则
//				FrontRepateTime:   p.RepeateTime,     // 前端给的原始数据
//				FrontDetailTime:   p.DetailTime,
//				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
//				//MsgLink:           p.MsgLink,
//				//MsgMarkDown:       p.MsgMarkDown,
//			}
//			return err, task
//		} else { //我要做定时任务
//			tasker := func() {}
//			zap.L().Info("进入定时任务模式")
//			tasker = func() {
//				err := d.SendMessage(p)
//				if err != nil {
//					return
//				} else {
//					zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", user.Name, robot.Title))
//				}
//			}
//			TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
//			tid = strconv.Itoa(int(TaskID))
//			if err != nil {
//				err = ErrorSpecInvalid
//				return err, dingding.Task{}
//			}
//			//把定时任务添加到数据库中
//			task = dingding.Task{
//				TaskID:            tid,               //cron给的taskid
//				TaskName:          p.TaskName,        //任务名称
//				DingUserID:        user.ID,           //所属用户id
//				UserName:          user.Name,         //所属用户姓名
//				RobotId:           robot.RobotId,     //机器人id
//				RobotName:         robot.Title,       //机器人name
//				RobotSecret:       robot.Secret,      //机器人Secret
//				DetailTimeForUser: detailTimeForUser, //给用户看的
//				Spec:              spec,              //cron后端定时规则
//				FrontRepateTime:   p.RepeateTime,     // 前端给的原始数据
//				FrontDetailTime:   p.DetailTime,
//				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
//				//MsgLink:           p.MsgLink,
//				//MsgMarkDown:       p.MsgMarkDown,
//			}
//			err = mysql.InsertTask(task)
//			if err != nil {
//				zap.L().Info(fmt.Sprintf("定时任务插入数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", user.Name, robot.Title, p.DetailTime, zap.Error(err)))
//				return err, dingding.Task{}
//			}
//			zap.L().Info(fmt.Sprintf("定时任务插入数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", user.Name, robot.Title, p.DetailTime))
//		}
//	} else if p.MsgLink.Msgtype == "link" {
//		if (p.RepeateTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
//			zap.L().Info("进入即时发送消息模式")
//			err := d.SendMessage(p)
//			if err != nil {
//				return err, dingding.Task{}
//			} else {
//				zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", user.Name, robot.Title))
//			}
//			//定时任务
//			task = dingding.Task{}
//			return err, task
//		} else { //我要做定时任务
//			tasker := func() {}
//			zap.L().Info("进入定时任务模式")
//			tasker = func() {
//				err := d.SendMessage(p)
//				if err != nil {
//					return
//				} else {
//					zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", user.Name, robot.Title))
//				}
//			}
//			TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
//			tid = strconv.Itoa(int(TaskID))
//			if err != nil {
//				err = mysql.ErrorSpecInvalid
//				return err, dingding.Task{}
//			}
//			//把定时任务添加到数据库中
//			task = dingding.Task{
//
//				TaskID:            tid,               //cron给的taskid
//				TaskName:          p.TaskName,        //任务名称
//				DingUserID:        user.ID,           //所属用户id
//				UserName:          user.Name,         //所属用户姓名
//				RobotId:           robot.RobotId,     //机器人id
//				RobotName:         robot.Title,       //机器人name
//				RobotSecret:       robot.Secret,      //机器人Secret
//				DetailTimeForUser: detailTimeForUser, //给用户看的
//				Spec:              spec,              //cron后端定时规则
//				FrontRepateTime:   p.RepeateTime,     // 前端给的原始数据
//				FrontDetailTime:   p.DetailTime,
//				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
//				//MsgLink:           p.MsgLink,
//				//MsgMarkDown:       p.MsgMarkDown,
//			}
//			err = mysql.InsertTask(task)
//			if err != nil {
//				zap.L().Info(fmt.Sprintf("定时任务插入数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", user.Name, robot.Title, p.DetailTime, zap.Error(err)))
//				return err, dingding.Task{}
//			}
//			zap.L().Info(fmt.Sprintf("定时任务插入数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", user.Name, robot.Title, p.DetailTime))
//		}
//	} else if p.MsgMarkDown.Msgtype == "markdown" {
//		//err = AtMobiles(p, &robot, &user)
//		//if err != nil {
//		//	zap.L().Error("通过人名查询电话号码失败", zap.Error(err))
//		//	return
//		//}
//		//if (p.RepeateTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
//		//	zap.L().Info("进入即时发送消息模式")
//		//	err := d.SendMessage(p)
//		//	if err != nil {
//		//		return err, dingding.Task{}
//		//	} else {
//		//		zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", user.Username, robot.RobotName))
//		//	}
//		//	//定时任务
//		//	task = dingding.Task{
//		//		Version:           p.Version,
//		//		TaskID:            tid,
//		//		TaskName:          p.TaskName,
//		//		UserID:            user.UserId,
//		//		UserName:          user.Username,
//		//		RobotId:           robot.RobotId,
//		//		RobotName:         robot.RobotName,
//		//		RobotSecret:       robot.Secret,
//		//		DetailTimeForUser: detailTimeForUser, //给用户看的
//		//		Spec:              spec,              //cron后端定时规则
//		//		FrontRepateTime:   p.RepeateTime,     // 前端给的原始数据
//		//		FrontDetailTime:   p.DetailTime,
//		//		MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
//		//		//MsgLink:           p.MsgLink,
//		//		//MsgMarkDown:       p.MsgMarkDown,
//		//	}
//		//	return err, task
//		//} else { //我要做定时任务
//		//	tasker := func() {}
//		//	zap.L().Info("进入定时任务模式")
//		//	tasker = func() {
//		//		err := d.SendMessage(p)
//		//		if err != nil {
//		//			return
//		//		} else {
//		//			zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", user.Username, robot.RobotName))
//		//		}
//		//	}
//		//	TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
//		//	tid = strconv.Itoa(int(TaskID))
//		//	if err != nil {
//		//		err = mysql.ErrorSpecInvalid
//		//		return err, dingding.Task{}
//		//	}
//		//	//把定时任务添加到数据库中
//		//	task = dingding.Task{
//		//		Version:           p.Version,
//		//		TaskID:            tid,               //cron给的taskid
//		//		TaskName:          p.TaskName,        //任务名称
//		//		UserID:            user.UserId,       //所属用户id
//		//		UserName:          user.Username,     //所属用户姓名
//		//		RobotId:           robot.RobotId,     //机器人id
//		//		RobotName:         robot.RobotName,   //机器人name
//		//		RobotSecret:       robot.Secret,      //机器人Secret
//		//		DetailTimeForUser: detailTimeForUser, //给用户看的
//		//		Spec:              spec,              //cron后端定时规则
//		//		FrontRepateTime:   p.RepeateTime,     // 前端给的原始数据
//		//		FrontDetailTime:   p.DetailTime,
//		//		MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
//		//		//MsgLink:           p.MsgLink,
//		//		//MsgMarkDown:       p.MsgMarkDown,
//		//	}
//		//	err = mysql.InsertTask(task)
//		//	if err != nil {
//		//		zap.L().Info(fmt.Sprintf("定时任务插入数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", user.Username, robot.RobotName, p.DetailTime, zap.Error(err)))
//		//		return err, dingding.Task{}
//		//	}
//		//	zap.L().Info(fmt.Sprintf("定时任务插入数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", user.Username, robot.RobotName, p.DetailTime))
//		//}
//	}
//
//	global.GLOAB_CORN.Start()
//	//我想，我在此处我们应该把定时任务crontab给传递到上下文中，好让我在其他地方的路由中去拿到,但是不同路由的上下文是不同的，所以我们无法在另外的路由中拿到
//	//c.Set(mysql.CtxCornTab, utils.Gcontab)
//	//c.Next() // 后续的处理函数可以用过c.Get("username")来获取当前请求的用户信息
//	return err, task
//
//}

//关闭后重新开启定时任务，之前更新的deleted_at字段，现在需要更新task_id字段
func ReStart(c *gin.Context, p *dingding.ParamCronSend, oldTask dingding.Task) (err error, task dingding.Task) {
	spec, detailTimeForUser, err := HandleSpec(p)
	tid := "0"
	user_id, err := global.GetCurrentUserId(c)
	if err != nil {
		user_id = ""
	}
	user, err := (&dingding.DingUser{UserId: user_id}).GetUserByUserId()
	if err != nil {
		user = dingding.DingUser{}
	}
	//先来判断一下此用户是否有这个小机器人
	robot, err := (&dingding.DingRobot{RobotId: p.RobotId}).GetRobotByRobotId()
	if err != nil {
		zap.L().Error("通过机器人的robot_id获取机器人失败", zap.Error(err))
		return
	}
	//到了这里就说明这个用户有这个小机器人
	d := dingding.DingRobot{
		RobotId: robot.RobotId,
		Secret:  robot.Secret,
	}
	//crontab := cron.New(cron.WithSeconds()) //精确到秒
	//spec := "* 30 22 * * ?" //cron表达式，每五秒一次
	if p.MsgText.Msgtype == "text" {

	}
	if len(p.MsgText.At.AtMobiles) > 0 { //至少有@人，就去查对应的电话号码
		teles := []dingding.DingUser{}
		for _, AtMobile := range p.MsgText.At.AtMobiles {
			//拿到每个人的姓名，然后去数据库里面找这个人的电话号码
			tele := dingding.DingUser{}
			err := global.GLOAB_DB.Model(&robot).Where("personname = ? ", AtMobile.Name).Association("Teles").Find(&tele)
			if err != nil {
				zap.L().Error(fmt.Sprintf("未查询到该姓名对应的电话号码，用户:%s,机器人:%s,查询姓名:%s ", user.Name, robot.Name, AtMobile.Name), zap.Error(err))
				tele = dingding.DingUser{}
			}
			teles = append(teles, tele)
		}
		//teles中存储的是完整的人和电话号码的信息
		finalTele := make([]common.AtMobile, len(teles))
		for index, tele := range teles {
			finalTele[index].AtMobile = tele.Mobile
		}
		//把tele1中的所有字符串，赋值给finalTele
		p.MsgText.At.AtMobiles = finalTele
	}
	if (p.RepeateTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
		zap.L().Info("进入即时发送消息模式")
		err := d.SendMessage(p)
		if err != nil {
			return err, dingding.Task{}
		} else {
			zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", user.Name, robot.Name))
		}
		//定时任务
		task = dingding.Task{
			TaskID:            tid,
			TaskName:          p.TaskName,
			UserId:            user.UserId,
			RobotId:           robot.RobotId,
			RobotName:         robot.Name,
			RobotSecret:       robot.Secret,
			DetailTimeForUser: detailTimeForUser, //给用户看的
			Spec:              spec,              //cron后端定时规则
			FrontRepateTime:   p.RepeateTime,     // 前端给的原始数据
			FrontDetailTime:   p.DetailTime,
			MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
			//MsgLink:           p.MsgLink,
			//MsgMarkDown:       p.MsgMarkDown,
		}
		return err, task
	} else { //我要做定时任务

		tasker := func() {}
		zap.L().Info("进入定时任务模式")
		tasker = func() {
			err := d.SendMessage(p)
			if err != nil {
				return
			} else {
				zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", user.Name, robot.Name))
			}
		}

		TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
		tid = strconv.Itoa(int(TaskID))
		if err != nil {
			err = mysql.ErrorSpecInvalid
			return err, dingding.Task{}
		}
		//把定时任务添加到数据库中
		task := dingding.Task{
			Model: gorm.Model{
				ID: oldTask.ID,
			},
			TaskID: tid, //我们只用更新task_id这一个字段就可以了
		}
		username, _ := c.Get(global.CtxUserNameKey)
		err = mysql.UpdateTask(task)
		if err != nil {
			zap.L().Error(fmt.Sprintf("定时任务关闭后再次打开时更新数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", username, robot.Name, p.DetailTime, zap.Error(err)))
			return err, dingding.Task{}
		}
		zap.L().Info(fmt.Sprintf("定时任务关闭后再次打开时数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", username, robot.Name, p.DetailTime))

		//我想，我在此处我们应该把定时任务crontab给传递到上下文中，好让我在其他地方的路由中去拿到,但是不同路由的上下文是不同的，所以我们无法在另外的路由中拿到
		//c.Set(mysql.CtxCornTab, utils.Gcontab)
		//c.Next() // 后续的处理函数可以用过c.Get("username")来获取当前请求的用户信息
		return err, task
	}
}

//通过机器人的id，拿到特定机器人的所有定时任务
func GetTasks(c *gin.Context, p *params.ParamGetTasks) (err error, tasks []dingding.Task) {

	//user_id, err := global.GetCurrentUserId(c)
	//if err != nil {
	//	return err, []dingding.Task{}
	//}
	//_, err = (&dingding.DingUser{Model: gorm.Model{ID: uid}}).GetUserByIDOrUserID()
	//if err != nil {
	//	return err, []dingding.Task{}
	//}
	//
	////如何拿到定时任务呢？只通过机器人姓名是不行的，需要通过机器人的robot_id，因为robot_id是唯一的
	//robot, err := mysql.GetRobotByRobotId(p.RobotId)
	//
	//var tasks []dingding.Task
	//err = global.GLOAB_DB.Model(&robot).Unscoped().Association("Tasks").Find(&tasks) //通过机器人的id拿到机器人，拿到机器人后，我们就可以拿到所有的任务
	//if err != nil {
	//	zap.L().Error("通过机器人robot_id拿到该机器人的所有定时任务失败", zap.Error(err))
	//}

	return err, tasks
}
