package version

import (
	"fmt"
	"runtime"
	"strings"
)

// modify these values from build flag
var (
	Major     = "0"
	Minor     = "0"
	Patch     = "0"
	Revision  = "Unknown"
	BuildDate = "Unknown"
)

// Semver semantic version
func Semver() string {
	return fmt.Sprintf("%s.%s.%s", Major, Minor, Patch)
}

// Detail all infomations of ssx version
func Detail() string {
	return strings.TrimSpace(fmt.Sprintf(`
Version:    %s
Revision:   %s
Buildtime:  %s
OS/Arch:    %s/%s
`, Semver(), Revision, BuildDate, runtime.GOOS, runtime.GOARCH))
}
