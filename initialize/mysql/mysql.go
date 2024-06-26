package mysql

import (
	"ding/global"
	"ding/initialize/viper"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func Init(cfg *viper.MySQLConfig) (err error) {
	DSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出到标准输出）
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // 日志级别
			Colorful:      true,        // 禁用彩色打印
		},
	)
	db, err := gorm.Open(mysql.New(mysql.Config{
		//DSN: "root:123456@tcp(121.43.119.224:3306)/gorm_class?charset=utf8mb4&parseTime=True&loc=Local",
		DSN: DSN, // 1. 连接信息

	}), &gorm.Config{ // 2. 选项
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true, //不用物理外键，使用逻辑外键
		Logger:                                   newLogger,
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
	//自动建表
	//err = RegisterTables(global.GLOAB_DB)
	//if err != nil {
	//	return
	return nil
}
