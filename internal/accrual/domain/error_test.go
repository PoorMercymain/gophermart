package domain

import (
	"reflect"
	"testing"
	"time"
)

func Test_getTimeInterval(t *testing.T) {
	type args struct {
		seconds float64
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "Test 1 second",
			args: args{
				seconds: 1,
			},
			want: time.Duration(1) * time.Second,
		},
		{
			name: "Test 10 seconds",
			args: args{
				seconds: 10,
			},
			want: time.Duration(10) * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTimeInterval(tt.args.seconds); !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("getTimeInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
