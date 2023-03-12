package v2

//func GetAttendances(p *params.ParamGetAttendances) (Atts []params.ParamGetAttendances, err error) {
//	//var Atts []model.Attendance
//	err = global.GLOAB_DB.Model(&model.Attendance{}).Where("chat_bot_user_id = ?", p.ChatBotUserId).Find(&Atts).Error
//	if err != nil {
//		zap.L().Error("根据chat_bot_user_id查询数据库失败", zap.Error(err))
//		return
//	}
//	return
//}
