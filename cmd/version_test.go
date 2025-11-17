package main

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	v := GetVersion()

	// Check that struct is populated
	if v.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if v.GOOS == "" {
		t.Error("GOOS should not be empty")
	}

	if v.GOARCH == "" {
		t.Error("GOARCH should not be empty")
	}

	// Check that Go runtime values match
	if v.GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %v, want %v", v.GoVersion, runtime.Version())
	}

	if v.GOOS != runtime.GOOS {
		t.Errorf("GOOS = %v, want %v", v.GOOS, runtime.GOOS)
	}

	if v.GOARCH != runtime.GOARCH {
		t.Errorf("GOARCH = %v, want %v", v.GOARCH, runtime.GOARCH)
	}
}

func TestVersionInfo_String(t *testing.T) {
	v := VersionInfo{
		Version:   "1.0.0",
		Commit:    "abc1234",
		BuildTime: "2024-01-15T10:00:00Z",
		GoVersion: "go1.21.0",
		GOOS:      "darwin",
		GOARCH:    "arm64",
	}

	str := v.String()

	// Check that all fields are present in output
	tests := []struct {
		name  string
		field string
	}{
		{"Version", "1.0.0"},
		{"Commit", "abc1234"},
		{"BuildTime", "2024-01-15T10:00:00Z"},
		{"GoVersion", "go1.21.0"},
		{"OS/Arch", "darwin/arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(str, tt.field) {
				t.Errorf("VersionInfo.String() should contain %q, got:\n%s", tt.field, str)
			}
		})
	}
}

func TestVersionInfo_ShortString(t *testing.T) {
	tests := []struct {
		name    string
		version VersionInfo
		want    string
	}{
		{
			name: "Standard version",
			version: VersionInfo{
				Version: "1.0.0",
				Commit:  "abc1234",
			},
			want: "v1.0.0 (abc1234)",
		},
		{
			name: "Dev version",
			version: VersionInfo{
				Version: "dev",
				Commit:  "1234567",
			},
			want: "vdev (1234567)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.ShortString()
			if got != tt.want {
				t.Errorf("VersionInfo.ShortString() = %v, want %v", got, tt.want)
			}
		})
	}
}
