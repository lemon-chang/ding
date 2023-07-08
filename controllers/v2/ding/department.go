package ding

import (
	"ding/controllers"
	"ding/global"
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
		response.FailWithMessage("参数错误", c)
		return
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
		response.FailWithMessage("参数错误", c)
	}
	//var t dingding2.DingToken
	//token, err := t.GetAccessToken()
	var d dingding2.DingDept
	DepartmentList, _, err := d.GetDeptByListFromMysql(&p)
	for i, dept := range DepartmentList {
		//获取到部门负责人
		var userids []string
		global.GLOAB_DB.Table("user_dept").Where("is_responsible = ? AND ding_dept_dept_id = ?", true, dept.DeptId).Select("ding_user_user_id").Find(&userids)
		if err != nil {
			zap.L().Error("查询部门下的负责人id失败", zap.Error(err))
			response.FailWithMessage("查询部门下的负责人id失败", c)
			return
		}
		err := global.GLOAB_DB.Model(&dingding2.DingUser{}).Where("user_id IN ?", userids).Find(&DepartmentList[i].ResponsibleUsers).Error
		//fmt.Println(users)
		//DepartmentList[i].ResponsibleUsers = users
		if err != nil {
			zap.L().Error("查询部门下的负责人信息失败", zap.Error(err))
			response.FailWithMessage("查询部门下的负责人信息失败", c)
			return
		}
	}
	//成功后返回部门信息
	if err != nil {
		response.FailWithMessage("获取子部门信息失败！", c)
		return
	}
	response.OkWithDetailed(DepartmentList, "获取部门信息成功", c)
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
	//判断要操作的部门是否存在
	var count int64
	err := global.GLOAB_DB.Table("ding_depts").Where("dept_id", p.DeptID).Count(&count).Error
	if count == 0 {
		response.FailWithMessage("部门不存在", c)
		return
	}
	err = global.GLOAB_DB.Table("ding_depts").Where("dept_id", p.DeptID).Update("is_robot_attendance", p.IsRobotAttendance).Error
	//使用这个会报错
	//d := dingding2.DingDept{}
	//err = d.UpdateDept(&p)
	if err != nil {
		response.FailWithMessage("更新部门信息失败！", c)
		return
	}
	response.OkWithMessage("更新部门信息成功！", c)
}

//更新部门是否在校信息
func UpdateSchool(c *gin.Context) {
	var s ding.ParameIsInSchool
	if err := c.ShouldBindJSON(&s); err != nil {
		zap.L().Error("UpdateSchool invaild param", zap.Error(err))
		response.ResponseError(c, response.CodeInvalidParam)
		return
	}
	if s.GroupId == 0 {
		response.FailWithMessage("部门名称或者部门id不能为空", c)
		return
	}
	d := dingding2.DingAttendGroup{}
	err := d.UpdateSchool(&s)
	if err != nil {
		zap.L().Error("更新数据库问题", zap.Error(err))
		response.ResponseError(c, response.CodeInvalidParam)
		return
	}
	response.ResponseSuccess(c, response.CodeSuccess)
}

type ParamSetDeptManager struct {
	UserId         string `json:"user_id"`
	DeptId         int    `json:"dept_id"`
	is_responsible bool   `json:"is_responsible"`
}

//设置或修改部门负责人
func SetDeptManager(c *gin.Context) {
	//给我一个用户id和该用户所在的部门id，存到user_dept表中，在每次查考勤的时候就会抄送给部门负责人一份
	var p *ParamSetDeptManager
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("参数错误", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	//更新数据库中的字段
	err := global.GLOAB_DB.Table("user_dept").Where("ding_user_user_id = ? AND ding_dept_dept_id = ?", p.UserId, p.DeptId).Update("is_responsible", p.is_responsible).Error
	if err != nil {
		zap.L().Error("更新管理员字段失败", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.ResponseSuccess(c, "更新成功")
}
