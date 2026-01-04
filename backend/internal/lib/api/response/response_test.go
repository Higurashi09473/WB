// Package response provides standardized JSON response structures and helpers
// for HTTP handlers in the WB backend API.
package response

import (
	"reflect"
	"testing"
)

func TestOK(t *testing.T) {
	tests := []struct {
		name string
		want Response
	}{
		{
			name: "basic",
			want: Response{Status: statusOK},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OK(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want Response
	}{
		{
			name: "with message",
			args: args{msg: "some error"},
			want: Response{Status: statusError, Error: "some error"},
		},
		{
			name: "empty message",
			args: args{msg: ""},
			want: Response{Status: statusError, Error: ""},
		},
		{
			name: "nil message",
			args: args{},
			want: Response{Status: statusError, Error: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Error(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}