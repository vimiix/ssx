package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	// Major is the major version of ptk
	Major = "0"
	// Minor is the minor version of ptk
	Minor = "0"
	// Patch is the patch version of ptk
	Patch = "0"
	// Revision is the current git commit hash
	Revision = "Unknown"
	// BuildDate is the built date time
	BuildDate = "Unknown"
)

func Semver() string {
	return fmt.Sprintf("%s.%s.%s", Major, Minor, Patch)
}

func Detail() string {
	return strings.TrimSpace(fmt.Sprintf(`
Version:    %s
Revision:   %s
Buildtime:  %s
OS/Arch:    %s/%s
`, Semver(), Revision, BuildDate, runtime.GOOS, runtime.GOARCH))
}
