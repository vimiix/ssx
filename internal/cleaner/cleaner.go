package cleaner

import (
	"sync"
)

var (
	mu  = sync.Mutex{}
	cbs []func()
)

func RegisterCallback(cb func()) {
	mu.Lock()
	defer mu.Unlock()
	cbs = append(cbs, cb)
}

func Clean() {
	mu.Lock()
	defer mu.Unlock()
	for _, cb := range cbs {
		cb()
	}
}
