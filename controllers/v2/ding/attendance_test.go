package ding

import (
	v2 "ding/logic/v2"
	"ding/model/params"
	"ding/utils"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDeptFirstShowUpMorning(t *testing.T) {
	p := &params.ParamGetDeptFirstShowUpMorning{
		GroupID: 952645016,
		Token:   "427528104bfe34ca8cfcd29553274d01",
	}
	utils.InitCorn()
	v2.DeptFirstShowUpMorning(p)
}
func TestTimeTransFrom(t *testing.T) {
	v2.TimeTransFrom()
}
func TestDeptFirstShowUpMorning1(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		//DeptFirstShowUpMorning(c),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
