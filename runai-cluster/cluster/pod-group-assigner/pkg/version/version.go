package version

import (
	"fmt"
	"runtime"
)

// This version information is injected via LDFLAGS in the make build
var (
	buildDate      = "" // output from `date -u +'%Y-%m-%dT%H:%M:%SZ'`
	gitCommit      = "" // output from `git rev-parse HEAD`
	gitShortCommit = "" // output from `git rev-parse HEAD`
	gitTag         = "" // output from `git describe --exact-match --tags HEAD` (if clean tree state)
)

type VersionInfo struct {
	Version   string
	BuildDate string
	GitCommit string
	GitTag    string
	GoVersion string
	Compiler  string
	Platform  string
}

func (version VersionInfo) String() string {
	return version.Version
}

func Version() VersionInfo {
	versionStr := fmt.Sprintf("%s-DEVELOPMENT", gitShortCommit)
	if //goland:noinspection GoBoolExpressions --> its populated by ldFlags
	gitTag != "" {
		versionStr = gitTag
	}

	return VersionInfo{
		Version:   versionStr,
		BuildDate: buildDate,
		GitCommit: gitCommit,
		GitTag:    gitTag,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
