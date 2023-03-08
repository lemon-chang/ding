package v2

import (
	"ding/controllers"
	"ding/model/dingding"
	"ding/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func OutGoing(c *gin.Context) {
	var p dingding.ParamReveiver
	err := c.ShouldBindJSON(&p)
	err = c.ShouldBindHeader(&p)
	if err != nil {
		zap.L().Error("OutGoing invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}
	}

	err = dingding.SendSessionWebHook(&p)
	if err != nil {
		zap.L().Error("钉钉机器人回调出错", zap.Error(err))
		response.ResponseErrorWithMsg(c, response.CodeServerBusy, "钉钉机器人回调出错")
		return
	}
	response.ResponseSuccess(c, "回调成功")

}
