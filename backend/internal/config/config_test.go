// Package config contains configuration structures and loading logic for the WB backend service.
// It supports reading configuration from YAML files and environment variables using cleanenv.
package config

import (
	"testing"
)

func TestPostgresql_DSN(t *testing.T) {
	postgresql := Postgresql{
		User:     "user",
		Password: "password",
		DBname:   "database",
		SSLmode:  "enable",
		Host:     "localhost",
		Port:     5432,
	}
	tests := []struct {
		name string
		p    Postgresql
		want string
	}{
		{
			name: "basic",
			p:    postgresql,
			want: "user=user password=password dbname=database sslmode=enable host=localhost port=5432",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.DSN(); got != tt.want {
				t.Errorf("Postgresql.DSN() = %v, want %v", got, tt.want)
			}
		})
	}
}
