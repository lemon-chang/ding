package mysql

import (
	"ding/global"
	"ding/model/dingding"
	"ding/settings"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func Init(cfg *settings.MySQLConfig) (err error) {
	DSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
	db, err := gorm.Open(mysql.New(mysql.Config{
		//DSN: "root:123456@tcp(121.43.119.224:3306)/gorm_class?charset=utf8mb4&parseTime=True&loc=Local",
		DSN: DSN, // 1. 连接信息
	}), &gorm.Config{ // 2. 选项
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true, //不用物理外键，使用逻辑外键
	})
	if err != nil {
		zap.L().Debug("数据库链接失败")
		return err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10) //
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	global.GLOAB_DB = db
	if err != nil {
		fmt.Println(err)
	}
	//err = reboot()
	//if err != nil {
	//	zap.L().Error("重启后项目恢复存在失败情况，程序没有终止，请排查", zap.Error(err))
	//}
	//global.GLOAB_CORN.Start() //在项目最初运行的时候，就要打开定时器，不然无法恢复成功
	//zap.L().Info("重启项目后恢复定时任务成功")
	return nil
}
func GxpInit(cfg *settings.MySQLConfig) (err error) {
	DSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/gxp?charset=utf8&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)
	db, err := gorm.Open(mysql.New(mysql.Config{
		//DSN: "root:123456@tcp(121.43.119.224:3306)/gorm_class?charset=utf8mb4&parseTime=True&loc=Local",
		DSN: DSN, // 1. 连接信息
	}), &gorm.Config{ // 2. 选项
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true, //不用物理外键，使用逻辑外键
	})
	if err != nil {
		zap.L().Debug("数据库链接失败")
		return err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10) //
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	global.GLOAB_DB1 = db
	if err != nil {
		fmt.Println(err)
	}

	global.GLOAB_DB1.AutoMigrate(&dingding.TongXinUser{}, &dingding.Record{})
	return nil
}

//func RegisterTables(db *gorm.DB) (err error) {
//	err = db.AutoMigrate(
//		model.User{},
//		model.Robot{},
//		model.Tele{},
//		model.Task{},
//		model.Attendance{},
//		model.MsgText{},
//		model.MsgLink{},
//		model.MsgMarkDown{},
//		model.MarkDown{},
//		model.At{},
//		model.Text{},
//		model.AtMobile{},
//		dingding2.DingDept{},
//		dingding2.DingAttendGroup{},
//		system.Config{},
//	)
//	if err != nil {
//		zap.L().Error("register table failed", zap.Error(err))
//		os.Exit(0)
//	}
//	//err = db.SetupJoinTable(&examination.ExamPaper{}, "ExamQuestions", &examination.PaperQuestions{})
//	//if err != nil {
//	//	global.GVA_LOG.Error("自定义多对多关系表出错 failed", zap.Error(err))
//	//	os.Exit(0)
//	//}
//	zap.L().Info("register table success")
//	return err
//}

func SwitchOffRobotAttendance() {

}
