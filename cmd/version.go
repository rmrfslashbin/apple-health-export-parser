package main

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	// Version is the version number
	Version = "dev"
	// Commit is the git commit hash
	Commit = "none"
	// BuildTime is the build timestamp
	BuildTime = "unknown"
)

// VersionInfo holds version information
type VersionInfo struct {
	Version   string
	Commit    string
	BuildTime string
	GoVersion string
	GOOS      string
	GOARCH    string
}

// GetVersion returns the full version info
func GetVersion() VersionInfo {
	return VersionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
	}
}

// String returns a formatted version string
func (v VersionInfo) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Version:    %s\n", v.Version))
	sb.WriteString(fmt.Sprintf("Commit:     %s\n", v.Commit))
	sb.WriteString(fmt.Sprintf("Built:      %s\n", v.BuildTime))
	sb.WriteString(fmt.Sprintf("Go version: %s\n", v.GoVersion))
	sb.WriteString(fmt.Sprintf("OS/Arch:    %s/%s", v.GOOS, v.GOARCH))

	return sb.String()
}

// ShortString returns a concise version string
func (v VersionInfo) ShortString() string {
	return fmt.Sprintf("v%s (%s)", v.Version, v.Commit[:7])
}
