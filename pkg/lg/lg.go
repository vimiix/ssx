package lg

import (
	"log"
	"sync"
)

var (
	verbose bool
	lock    = sync.RWMutex{}
)

func SetVerbose(v bool) {
	lock.Lock()
	defer lock.Unlock()
	verbose = v
}

func Logf(format string, v ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if verbose {
		log.Printf(format, v...)
	}
}

func Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}
