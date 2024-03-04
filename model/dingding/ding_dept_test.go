package dingding

import (
	"ding/model/common/localTime"
	"gorm.io/gorm"
	"reflect"
	"testing"
	"time"
)

func TestDingDept_GetAttendanceData(t *testing.T) {
	type fields struct {
		CreatedAt         time.Time
		UpdatedAt         time.Time
		UserList          []DingUser
		DeptId            int
		Deleted           gorm.DeletedAt
		Name              string
		ParentId          int
		DingToken         DingToken
		IsSendFirstPerson int
		RobotToken        string
		IsRobotAttendance bool
		IsJianShuOrBlog   int
		IsLeetCode        int
		ResponsibleUsers  []DingUser
		Children          []DingDept
		NumberAttendUser  int
	}
	type args struct {
		userids     []string
		curTime     *localTime.MySelfTime
		OnDutyTime  []string
		OffDutyTime []string
		isInSchool  bool
	}
	tests := []struct {
		name                    string
		fields                  fields
		args                    args
		wantResult              map[string][]DingAttendance
		wantAttendanceList      []DingAttendance
		wantNotRecordUserIdList []string
		wantErr                 bool
	}{
		// TODO: Add test cases.
		{
			name: "大海涛",
			fields: fields{
				DeptId:    546623914,
				DingToken: DingToken{Token: "81fd2415360d3bc68de54a8e0c5c43e4"},
			},
			args: args{

				//userids: GetUserIdListByUserList(deptAttendanceUser[DeptId])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DingDept{
				CreatedAt:         tt.fields.CreatedAt,
				UpdatedAt:         tt.fields.UpdatedAt,
				UserList:          tt.fields.UserList,
				DeptId:            tt.fields.DeptId,
				Deleted:           tt.fields.Deleted,
				Name:              tt.fields.Name,
				ParentId:          tt.fields.ParentId,
				DingToken:         tt.fields.DingToken,
				IsSendFirstPerson: tt.fields.IsSendFirstPerson,
				RobotToken:        tt.fields.RobotToken,
				IsRobotAttendance: tt.fields.IsRobotAttendance,
				IsJianShuOrBlog:   tt.fields.IsJianShuOrBlog,
				IsLeetCode:        tt.fields.IsLeetCode,
				ResponsibleUsers:  tt.fields.ResponsibleUsers,
				Children:          tt.fields.Children,
				NumberAttendUser:  tt.fields.NumberAttendUser,
			}
			gotResult, gotAttendanceList, gotNotRecordUserIdList, err := d.GetAttendanceData(tt.args.userids, tt.args.curTime, tt.args.OnDutyTime, tt.args.OffDutyTime, tt.args.isInSchool)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAttendanceData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("GetAttendanceData() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
			if !reflect.DeepEqual(gotAttendanceList, tt.wantAttendanceList) {
				t.Errorf("GetAttendanceData() gotAttendanceList = %v, want %v", gotAttendanceList, tt.wantAttendanceList)
			}
			if !reflect.DeepEqual(gotNotRecordUserIdList, tt.wantNotRecordUserIdList) {
				t.Errorf("GetAttendanceData() gotNotRecordUserIdList = %v, want %v", gotNotRecordUserIdList, tt.wantNotRecordUserIdList)
			}
		})
	}
}
