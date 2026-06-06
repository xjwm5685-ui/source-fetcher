package main

import (
	"fmt"
	"runtime"
)

// Version information (set by build flags)
var (
	// Version is the semantic version of source-fetcher
	Version = "1.0.1"

	// BuildTime is when the binary was built
	BuildTime = "unknown"

	// GitCommit is the git commit hash
	GitCommit = "unknown"
)

// VersionInfo contains all version-related information
type VersionInfo struct {
	Version   string
	BuildTime string
	GitCommit string
	GoVersion string
	Platform  string
	Arch      string
}

// GetVersionInfo returns structured version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	info := GetVersionInfo()
	return fmt.Sprintf("source-fetcher %s", info.Version)
}

// GetFullVersionString returns a detailed version string
func GetFullVersionString() string {
	info := GetVersionInfo()
	return fmt.Sprintf(`source-fetcher version %s
  Build Time: %s
  Git Commit: %s
  Go Version: %s
  Platform:   %s/%s`,
		info.Version,
		info.BuildTime,
		info.GitCommit,
		info.GoVersion,
		info.Platform,
		info.Arch,
	)
}

// PrintVersion prints the full version information
func PrintVersion() {
	fmt.Println(GetFullVersionString())
}
