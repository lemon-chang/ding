package ding

//获取考勤记录
//func GetAttendances(c *gin.Context) {
//	var p *params.ParamGetAttendances
//	if err := c.ShouldBindJSON(&p); err != nil {
//		zap.L().Error("BatchInsertGroupMembers invaild param", zap.Error(err))
//		errs, ok := err.(validator.ValidationErrors)
//		if !ok {
//			response.ResponseError(c, response.CodeInvalidParam)
//			return
//		} else {
//			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
//			return
//		}
//	}
//	atts, err := v2.GetAttendances(p)
//	if err != nil {
//		zap.L().Error("logic GetAttendances查询失败", zap.Error(err))
//		response.ResponseError(c, response.CodeInvalidParam)
//	}
//	response.ResponseSuccess(c, atts)
//
//}
