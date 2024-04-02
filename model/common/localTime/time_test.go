package localTime

import (
	"testing"
	"time"
)

func TestMySelfTime_GetCurTime(t1 *testing.T) {
	type fields struct {
		TimeStamp   int64
		Format      string
		Time        time.Time
		Duration    int
		ClassNumber int
		Week        int
		StartWeek   int
		Semester    string
	}
	type args struct {
		commutingTime map[string][]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "测试"},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &MySelfTime{
				TimeStamp:   tt.fields.TimeStamp,
				Format:      tt.fields.Format,
				Time:        tt.fields.Time,
				Duration:    tt.fields.Duration,
				ClassNumber: tt.fields.ClassNumber,
				Week:        tt.fields.Week,
				StartWeek:   tt.fields.StartWeek,
				Semester:    tt.fields.Semester,
			}
			if err := t.GetCurTime(tt.args.commutingTime); (err != nil) != tt.wantErr {
				t1.Errorf("GetCurTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
