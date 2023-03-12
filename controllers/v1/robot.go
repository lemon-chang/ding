package v1

import (
	"ding/controllers"
	"ding/logic"
	"ding/model/params"
	"ding/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

//Send 机器人发送消息

func GetTasks(c *gin.Context) {
	var p params.ParamGetTasks
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("GetTasks invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}

	}
	//开始获取任务
	err, tasks := logic.GetTasks(c, &p)
	if err != nil {
		zap.L().Error("获取某机器人的所有任务失败，失败原因：", zap.Error(err))
		response.ResponseError(c, response.CodeInvalidParam)
		return
	}
	response.ResponseSuccess(c, tasks)

}
