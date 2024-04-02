package ding

import (
	"ding/global"
	response2 "ding/model/common/response"
	dingding2 "ding/model/dingding"
	"ding/model/params"
	"ding/model/params/ding"
	"ding/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

// ImportDeptData 递归获取部门列表（官方接口）
func ImportDeptData(c *gin.Context) {
	var d dingding2.DingDept
	t := dingding2.DingToken{}
	token, err := t.GetAccessToken()
	d.DingToken.Token = token
	departmentList, err := d.ImportDeptData()
	if err != nil {
		response.FailWithMessage("导入部门数据失败", c)
		return
	}
	response.OkWithDetailed(departmentList, "导入部门数据成功", c)
}

// GetSubDepartmentListByID 获取子部门通过id （官方接口）
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

// GetSubDepartmentListByID2 获取子部门通过id （mysql）
func GetSubDepartmentListByID2(c *gin.Context) {
	var p params.ParamGetDepartmentListByID2
	if err := c.ShouldBindQuery(&p); err != nil {
		zap.L().Error("GetSubDepartmentListByID2 invaild param", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	var d dingding2.DingDept
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
		return
	}
	var d dingding2.DingDept
	DepartmentList, total, err := d.GetDeptByListFromMysql(&p)
	//成功后返回部门信息
	if err != nil {
		response.FailWithMessage("获取部门列表失败", c)
		return
	}
	response.OkWithDetailed(response2.PageResult{
		List:     DepartmentList,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, "获取部门信息成功", c)
}

// UpdateDept 更新部门信息
func UpdateDept(c *gin.Context) {
	var p ding.ParamUpdateDept
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("UpdateDept invaild param", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}

	err := (&dingding2.DingDept{DeptId: p.DeptID}).UpdateDept(&p)
	if err != nil {
		response.FailWithMessage("更新部门信息失败！", c)
		return
	}
	response.OkWithMessage("更新部门信息成功！", c)
}

func GetUserByDeptId(c *gin.Context) {
	deptId := c.Query("dept_id")
	deptid, _ := strconv.Atoi(deptId)
	var p *dingding2.DingDept
	err := global.GLOAB_DB.Preload("UserList").Where("dept_id", deptid).First(&p).Error
	if err != nil {
		zap.L().Error("查询列表错误", zap.Error(err))
		response.FailWithMessage("查询列表错误", c)
		return
	}
	response.OkWithDetailed(p, "查询成功", c)
}

func GetDepartmentRecursively(c *gin.Context) {

	list, total, err := (&dingding2.DingDept{}).GetDepartmentRecursively()
	if err != nil {
		zap.L().Error("获取失败！", zap.Error(err))
		response.FailWithMessage("获取失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(struct {
		List  []dingding2.DingDept
		Total int64
	}{
		List:  list,
		Total: total,
	}, "获取成功", c)
}
