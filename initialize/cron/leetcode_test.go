package cron

import "testing"

func Test_getLeetCodeNumRaw(t *testing.T) {
	type args struct {
		leetCodeAddress string
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
		wantErr   bool
	}{
		// TODO: Add test cases.
		{args: args{leetCodeAddress: "xing-he-8f"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCount, err := getLeetCodeNumRaw(tt.args.leetCodeAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLeetCodeNumRaw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCount != tt.wantCount {
				t.Errorf("getLeetCodeNumRaw() gotCount = %v, want %v", gotCount, tt.wantCount)
			}
		})
	}
}
