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

func Debugf(format string, v ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if verbose {
		log.Printf("[debug] "+format, v...)
	}
}

func Infof(format string, v ...any) {
	log.Printf("[info] "+format, v...)
}

func Errorf(format string, v ...any) {
	log.Printf("[error] "+format, v...)
}
