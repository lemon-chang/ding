package mysql

import (
	"crypto/md5"
	"encoding/hex"
)

const secret = "liwenzhou.com"

// CheckUserExist 注册时检查用户是否存在
//func CheckUserExist(p *params.ParamSignUp) (err error) {
//
//	//result := GLOAB_DB.First(&p)//此处我们不能这样写，如果这样写的话，我们就要去param_sign_up里面查找数据
//	var user *dingding.DingUser
//	result := global.GLOAB_DB.Where("username = ?", p.Username).First(&user)
//	// 检查 ErrRecordNotFound 错误
//	if errors.Is(result.Error, gorm.ErrRecordNotFound) { //如果之前没有注册过
//		return
//	} else { //之前注册过了，或者有其他的错误
//		return ErrorUserExist
//	}
//
//}

// InsertUser 向数据库中插入一个新的用户记录
//func InsertUser(user *dingding.DingUser) (err error) {
//	//对密码进行加密
//	user.Password = encryptPassword(user.Password)
//	//执行sql数据入库
//	//sqlStr := `insert into user(user_id,username,password) values(?,?,?)`
//	//_, err = db.Exec(sqlStr, user.UserID, user.Username, user.Password)
//	//return err
//	result := global.GLOAB_DB.Create(&user)
//	if result.Error != nil {
//		return result.Error
//	} else {
//		fmt.Println(result.RowsAffected)
//		return
//	}
//
//}

//encryptPassword 解析用户密码
func encryptPassword(oPassword string) string {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum([]byte(oPassword)))
}

//encryptPassword  用户登录
//func Login(user *dingding.DingUser) (err error) {
//	opassword := user.Password //此处是用户输入的密码，不一定是对的
//	result := global.GLOAB_DB.Where(&dingding.DingUser{Name: user.Name}).First(user)
//	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
//		return ErrorUserNotExist
//	}
//	if result.Error != nil {
//		return result.Error
//	}
//	//如果到了这里还没有结束的话，那就说明该用户至少是存在的，于是我们解析一下密码
//
//	password := encryptPassword(opassword)
//	//拿到解析后的密码，我们看看是否正确
//	if password != user.Password {
//		return ErrorInvalidPassword
//	}
//	//如果能到这里的话，那就登录成功了
//	return nil
//
//}

//GetUserByID 通过id得到当前用户
