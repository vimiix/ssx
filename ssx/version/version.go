package version

import (
	"fmt"
	"runtime"
	"strings"
)

// modify these values from build flag
var (
	Version   = "develop"
	Revision  = "unknown"
	BuildDate = "unknown"
)

// Detail all infomations of ssx version
func Detail() string {
	return strings.TrimSpace(fmt.Sprintf(`
Version:    %s
Revision:   %s
Buildtime:  %s
OS/Arch:    %s/%s
`, Version, Revision, BuildDate, runtime.GOOS, runtime.GOARCH))
}
