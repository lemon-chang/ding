package model

//type Attendance struct {
//	gorm.Model
//	Content           string `json:"content"`            //机器人接收到的消息
//	ChatbotUserId     string `json:"chatbot_user_id"`    //加密的机器人id，我们只能这样存储了，钉钉回调无法具体到某个机器人
//	SenderNick        string `json:"sender_nick"`        //@机器人的那个人
//	ConversationTitle string `json:"conversation_title"` //群聊名称
//	SenderStaffId string `json:"sender_staff_id"` //userId
//	ChatBotUserId string `json:"chat_bot_user_id"` //考勤属于机器人，但是不能 机器人表中的ID以及RobotId关联起来，需要和ChatBotUserId关联起来，我们需要重写外键和重写引用
//}
