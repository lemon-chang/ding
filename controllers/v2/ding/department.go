package ding

import (
	"ding/controllers"
	dingding2 "ding/model/dingding"
	"ding/model/params"
	"ding/model/params/ding"
	"ding/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

//递归获取部门列表（官方接口）
func ImportDeptData(c *gin.Context) {
	//var p params.ParamGetDepartmentList
	//if err := c.ShouldBindJSON(&p); err != nil {
	//	zap.L().Error("GetDepartmentList invaild param", zap.Error(err))
	//	errs, ok := err.(validator.ValidationErrors)
	//	if !ok {
	//		response.ResponseError(c, response.CodeInvalidParam)
	//		return
	//	} else {
	//		response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
	//		return
	//	}
	//}
	var d dingding2.DingDept
	t := dingding2.DingToken{}
	token, err := t.GetAccessToken()

	d.DingToken.Token = token

	departmentList, err := d.ImportDeptData()
	if err != nil {
		response.ResponseErrorWithMsg(c, 0, gin.H{
			"message": err.Error(),
		})
		return
	}
	response.ResponseSuccess(c, gin.H{
		"message": "导入部门数据成功",
		"data":    departmentList,
	})
}

//获取考勤组列表 （官方接口）
func GetAttendancesGroups(c *gin.Context) {
	var p params.ParamGetAttendanceGroups
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("BatchInsertGroupMembers invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}
	}
	var d dingding2.DingAttendGroup
	d.DingToken.Token = p.Token
	AttendancesGroups, err := d.GetAttendancesGroups(p.Offset, p.Size)
	if err != nil {
		response.FailWithMessage("获取考勤组失败", c)
		return
	}
	response.OkWithDetailed(AttendancesGroups, "获取考勤组成功", c)
}

//获取子部门通过id （官方接口）
func GetSubDepartmentListByID(c *gin.Context) {
	var p params.ParamGetDepartmentListByID
	if err := c.ShouldBindQuery(&p); err != nil {
		zap.L().Error("GetDepartmentListByID invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}
	}
	var d dingding2.DingDept
	d.DingToken.Token = p.Token
	d.DeptId = p.ID
	subDepartments, err := d.GetDepartmentListByID()
	if err != nil {
		response.FailWithMessage("获取子部门信息失败！", c)
		return
	}
	response.OkWithDetailed(subDepartments, "获取子部门信息成功", c)
}

//获取子部门通过id （mysql）
func GetSubDepartmentListByID2(c *gin.Context) {
	var p params.ParamGetDepartmentListByID2
	if err := c.ShouldBindQuery(&p); err != nil {
		zap.L().Error("GetDepartmentListByID invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}
	}
	//var t dingding2.DingToken
	//token, err := t.GetAccessToken()
	var d dingding2.DingDept
	//d.DingToken.Token = token
	//d.DeptId = p.ID
	d.DeptId = p.ID
	subDepartments, err := d.GetDepartmentListByID2()
	if err != nil {
		response.FailWithMessage("获取子部门信息失败！", c)
		return
	}
	response.OkWithDetailed(subDepartments, "获取子部门信息成功", c)
}

func GetDeptListFromMysql(c *gin.Context) {
	var p params.ParamGetDeptListFromMysql
	if err := c.ShouldBindQuery(&p); err != nil {
		zap.L().Error("GetDepartmentListByID invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}
	}
	//var t dingding2.DingToken
	//token, err := t.GetAccessToken()
	var d dingding2.DingDept
	DepartmentList, total, err := d.GetDeptByListFromMysql(&p)
	if err != nil {
		response.FailWithMessage("获取子部门信息失败！", c)
		return
	}
	response.OkWithDetailed(gin.H{"list": DepartmentList, "total": total}, "获取部门信息成功", c)
}

//更新部门信息
func UpdateDept(c *gin.Context) {
	var p ding.ParamUpdateDeptToCron
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("UpdateDept invaild param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ResponseError(c, response.CodeInvalidParam)
			return
		} else {
			response.ResponseErrorWithMsg(c, response.CodeInvalidParam, controllers.RemoveTopStruct(errs.Translate(controllers.Trans)))
			return
		}
	}
	if p.DeptID == 0 {
		response.FailWithMessage("部门名称或者部门id不能为空", c)
		return
	}
	d := dingding2.DingDept{}
	err := d.UpdateDept(&p)
	if err != nil {
		response.FailWithMessage("更新部门信息失败！", c)
		return
	}
	response.OkWithMessage("更新部门信息成功！", c)
}
