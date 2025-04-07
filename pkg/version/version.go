package version

import (
	"fmt"
	"runtime"
)

// Version information
var (
	// Major version component
	Major = 0
	// Minor version component
	Minor = 1
	// Patch version component
	Patch = 0
	// Pre-release version component
	PreRelease = ""
	// Build metadata
	BuildMetadata = ""
	// Version in string format - set dynamically at build time
	Version = "0.1.0"
	// GitCommit is the git commit that was compiled - set dynamically at build time
	GitCommit = ""
	// BuildDate is the date of the build - set dynamically at build time
	BuildDate = ""
	// GoVersion is the version of go used to compile
	GoVersion = runtime.Version()
	// Platform is the operating system and architecture combination
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	// Name of the application
	AppName = "git-llm-review"
	// Description of the application
	Description = "A tool that uses LLMs to review code changes in Git repositories"
)

// GetVersionInfo returns a formatted version string with additional build information
func GetVersionInfo() string {
	var versionString string

	// Start with the basic version
	versionString = fmt.Sprintf("%s version %s", AppName, Version)

	// Add Git commit if available
	if GitCommit != "" {
		versionString += fmt.Sprintf("\nGit commit: %s", GitCommit)
	}

	// Add build date if available
	if BuildDate != "" {
		versionString += fmt.Sprintf("\nBuild date: %s", BuildDate)
	}

	// Add Go version
	versionString += fmt.Sprintf("\nGo version: %s", GoVersion)

	// Add platform information
	versionString += fmt.Sprintf("\nPlatform: %s", Platform)

	return versionString
}
