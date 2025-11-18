package main

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	// Test that GetVersion returns a valid VersionInfo
	v := GetVersion()

	if v.Version == "" {
		t.Error("Version should not be empty")
	}

	if v.Commit == "" {
		t.Error("Commit should not be empty")
	}

	if v.BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}

	if v.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
}

func TestVersionInfoString(t *testing.T) {
	tests := []struct {
		name      string
		version   VersionInfo
		checkFor  []string
	}{
		{
			name: "complete version info",
			version: VersionInfo{
				Version:   "v2025.11.17",
				Commit:    "abc123",
				BuildTime: "2025-11-17T10:00:00Z",
				GoVersion: "go1.25.4",
			},
			checkFor: []string{
				"Version:",
				"v2025.11.17",
				"Commit:",
				"abc123",
				"Built:",
				"2025-11-17T10:00:00Z",
				"Go version:",
				"go1.25.4",
			},
		},
		{
			name: "dev version",
			version: VersionInfo{
				Version:   "dev",
				Commit:    "none",
				BuildTime: "unknown",
				GoVersion: "go1.25.4",
			},
			checkFor: []string{
				"Version:",
				"dev",
				"Commit:",
				"none",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()

			for _, check := range tt.checkFor {
				if !strings.Contains(got, check) {
					t.Errorf("String() output missing expected text: %q\nGot: %s", check, got)
				}
			}
		})
	}
}

func TestVersionInfoShortString(t *testing.T) {
	tests := []struct {
		name    string
		version VersionInfo
		want    string
	}{
		{
			name: "release version",
			version: VersionInfo{
				Version:   "v2025.11.17",
				Commit:    "abc123",
				BuildTime: "2025-11-17T10:00:00Z",
				GoVersion: "go1.25.4",
			},
			want: "vv2025.11.17 (abc123)",
		},
		{
			name: "dev version",
			version: VersionInfo{
				Version:   "dev",
				Commit:    "none",
				BuildTime: "unknown",
				GoVersion: "go1.25.4",
			},
			want: "vdev (none)",
		},
		{
			name: "dirty version",
			version: VersionInfo{
				Version:   "v1.0.0-dirty",
				Commit:    "def456",
				BuildTime: "2025-11-17T10:00:00Z",
				GoVersion: "go1.25.4",
			},
			want: "vv1.0.0-dirty (def456)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.ShortString(); got != tt.want {
				t.Errorf("ShortString() = %v, want %v", got, tt.want)
			}
		})
	}
}
