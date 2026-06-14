package controller

import (
	"testing"
)

func TestParseBoolEnv(t *testing.T) { //nolint:paralleltest
	tests := []struct {
		name    string
		value   string
		set     bool
		def     bool
		want    bool
		wantErr bool
	}{
		{name: "unset returns default false", set: false, def: false, want: false},
		{name: "unset returns default true", set: false, def: true, want: true},
		{name: "empty returns default", set: true, value: "", def: true, want: true},
		{name: "true overrides default", set: true, value: "true", def: false, want: true},
		{name: "false overrides default", set: true, value: "false", def: true, want: false},
		{name: "invalid value returns error", set: true, value: "yes please", def: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const envName = "TFNOTIFY_TEST_BOOL_ENV"
			if tt.set {
				t.Setenv(envName, tt.value)
			}

			got, err := parseBoolEnv(envName, tt.def)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseBoolEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
