package app

import (
	"reflect"
	"testing"
)

func TestNewApp(t *testing.T) {
	type args struct {
		name     string
		host     string
		port     int
		metadata map[string]string
	}
	tests := []struct {
		name string
		args args
		want *App
	}{
		{
			name: "",
			args: args{
				name:     "test-app",
				host:     "127.0.0.1",
				port:     4399,
				metadata: map[string]string{"version": "0.0.1"},
			},
			want: &App{
				ID:       "test-app:127.0.0.1:4399",
				Name:     "test-app",
				Host:     "127.0.0.1",
				Port:     4399,
				Address:  "127.0.0.1:4399",
				Metadata: map[string]string{"version": "0.0.1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.name, tt.args.host, tt.args.port, tt.args.metadata)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
