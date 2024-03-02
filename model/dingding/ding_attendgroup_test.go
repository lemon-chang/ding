package dingding

import (
	"ding/model/params"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"reflect"
	"testing"
	"time"
)

func TestTimeTransFrom(t *testing.T) {
	//TimeTransFrom()
}

func Test_timestampTran(t *testing.T) {
	type args struct {
		format string
		t      int64
	}
	tests := []struct {
		name  string
		args  args
		wantS string
	}{
		// TODO: Add test cases.
		{"case1", args{format: "2006:01:02 15:04:05", t: 1672976808000}, ""},
		{"case1", args{format: "15:04:05", t: 1672976808000}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS := timestampTran(tt.args.format, tt.args.t); gotS != tt.wantS {
				t.Errorf("timestampTran() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

func TestDingAttendGroup_GetGroupDeptNumber(t *testing.T) {
	type fields struct {
		GroupId       int
		GroupName     string
		MemberCount   int
		WorkDayList   []string
		ClassesList   []string
		SelectedClass []struct {
			Setting struct {
				PermitLateMinutes int `json:"permit_late_minutes"` //允许迟到时长
			} `gorm:"-" json:"setting"`
			Sections []struct {
				Times []struct {
					CheckTime string `json:"check_time"` //打卡时间
					CheckType string `json:"check_type"` //打卡类型
				} `gorm:"-" json:"times"`
			} `gorm:"-" json:"sections"`
		}
		DingToken         DingToken
		IsRobotAttendance bool
		RobotAttendTaskID int
		IsSendFirstPerson int
	}
	tests := []struct {
		name          string
		fields        fields
		wantDeptUsers map[string][]DingUser
		wantErr       bool
	}{
		// TODO: Add test cases.
		{
			name:   "测试",
			fields: fields{GroupId: 1082662924, DingToken: DingToken{Token: "92a18001ddbd3382aca786ead4bd6889"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &DingAttendGroup{
				GroupId:           tt.fields.GroupId,
				GroupName:         tt.fields.GroupName,
				MemberCount:       tt.fields.MemberCount,
				WorkDayList:       tt.fields.WorkDayList,
				ClassesList:       tt.fields.ClassesList,
				SelectedClass:     tt.fields.SelectedClass,
				DingToken:         tt.fields.DingToken,
				IsRobotAttendance: tt.fields.IsRobotAttendance,
				RobotAttendTaskID: tt.fields.RobotAttendTaskID,
				IsSendFirstPerson: tt.fields.IsSendFirstPerson,
			}
			gotDeptUsers, err := a.GetGroupDeptNumber()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGroupDeptNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDeptUsers, tt.wantDeptUsers) {
				t.Errorf("GetGroupDeptNumber() gotDeptUsers = %v, want %v", gotDeptUsers, tt.wantDeptUsers)
			}
		})
	}
}

func TestDingAttendGroup_AlertAttendByRobot(t *testing.T) {
	type fields struct {
		CreatedAt     time.Time
		UpdatedAt     time.Time
		DeletedAt     gorm.DeletedAt
		GroupId       int
		GroupName     string
		MemberCount   int
		WorkDayList   []string
		ClassesList   []string
		SelectedClass []struct {
			Setting struct {
				PermitLateMinutes int `json:"permit_late_minutes"` //允许迟到时长
			} `gorm:"-" json:"setting"`
			Sections []struct {
				Times []struct {
					CheckTime string `json:"check_time"` //打卡时间
					CheckType string `json:"check_type"` //打卡类型
				} `gorm:"-" json:"times"`
			} `gorm:"-" json:"sections"`
		}
		DingToken              DingToken
		IsRobotAttendance      bool
		RobotAttendTaskID      int
		RobotAttendAlterTaskID int
		IsSendFirstPerson      int
		IsInSchool             bool
		IsReady                int
		ReadyTime              int
		NextTime               string
		IsSecondClass          int
		RestTimes              []RestTime
	}
	type args struct {
		p *params.ParamAllDepartAttendByRobot
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantTaskID cron.EntryID
		wantErr    bool
	}{
		// TODO: Add test cases.
		{
			name:   "测试",
			fields: fields{GroupId: 1082662924},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &DingAttendGroup{
				CreatedAt:         tt.fields.CreatedAt,
				UpdatedAt:         tt.fields.UpdatedAt,
				DeletedAt:         tt.fields.DeletedAt,
				GroupId:           tt.fields.GroupId,
				GroupName:         tt.fields.GroupName,
				MemberCount:       tt.fields.MemberCount,
				WorkDayList:       tt.fields.WorkDayList,
				ClassesList:       tt.fields.ClassesList,
				SelectedClass:     tt.fields.SelectedClass,
				DingToken:         tt.fields.DingToken,
				IsRobotAttendance: tt.fields.IsRobotAttendance,
				RobotAttendTaskID: tt.fields.RobotAttendTaskID,
				RobotAlterTaskID:  tt.fields.RobotAttendAlterTaskID,
				IsSendFirstPerson: tt.fields.IsSendFirstPerson,
				IsInSchool:        tt.fields.IsInSchool,
				IsAlert:           tt.fields.IsReady,
				ReadyTime:         tt.fields.ReadyTime,
				NextTime:          tt.fields.NextTime,
				IsSecondClass:     tt.fields.IsSecondClass,
				RestTimes:         tt.fields.RestTimes,
			}
			gotTaskID, err := a.AlertAttendByRobot(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("AlertAttendByRobot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTaskID != tt.wantTaskID {
				t.Errorf("AlertAttendByRobot() gotTaskID = %v, want %v", gotTaskID, tt.wantTaskID)
			}
		})
	}
}
