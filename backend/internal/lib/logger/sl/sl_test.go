// Package sl provides small utility helpers for working with slog.Logger.
// Currently it contains a convenient way to log errors as attributes.
package sl

import (
	"errors"
	"log/slog"
	"reflect"
	"testing"
)

func TestErr(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want slog.Attr
	}{
		{
			name: "basic",
			args: args{err: errors.New("error")},
			want: slog.Attr{
				Key:   "error",
				Value: slog.StringValue("error"),
			},
		},
		{
			name: "nil",
			args: args{},
			want: slog.Attr{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Err(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Err() = %v, want %v", got, tt.want)
			}
		})
	}
}
