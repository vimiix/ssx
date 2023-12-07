package lg

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

var (
	verbose = atomic.Bool{}
	logger  = log.New(os.Stderr, "", 0)
)

func SetVerbose(v bool) {
	verbose.Store(v)
}

func defaultPrint(lvl, message string) {
	ts := time.Now().Format(time.RFC3339)
	logger.Print(
		strings.Join([]string{ts, lvl, message}, " "),
	)
}

// printFunc 便于单元测试打桩
var printFunc = defaultPrint

func Debug(format string, v ...any) {
	if verbose.Load() {
		printFunc("DEBUG", fmt.Sprintf(format, v...))
	}
}

func Info(format string, v ...any) {
	printFunc("INFO", fmt.Sprintf(format, v...))
}

func Warn(format string, v ...any) {
	printFunc("WARN", color.YellowString(format, v...))
}

func Error(format string, v ...any) {
	printFunc("ERROR", color.RedString(format, v...))
}
