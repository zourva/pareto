package concurrent

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// see ref to: https://github.com/tidwall/spinlock/blob/master/locker.go

// spinLock is a spin lock implementation.
//
// NOTE: A Locker must not be copied after first use.
type spinLock struct {
	_    sync.Mutex // for copy protection compiler warning
	lock uintptr
}

// Lock locks l.
// Based on compare-and-swap, 0 is defined as unlocked, while 1 is locked.
//
// If the lock is already in use, the calling goroutine
// blocks until the locker is available.
func (l *spinLock) Lock() {
	for !atomic.CompareAndSwapUintptr(&l.lock, 0, 1) {
		runtime.Gosched()
	}
}

// Unlock unlocks l.
func (l *spinLock) Unlock() {
	atomic.StoreUintptr(&l.lock, 0)
}

// NewSpinLock creates a spin lock.
//
// NOTE: It is NOT re-entrant.
func NewSpinLock() sync.Locker {
	var lock spinLock
	return &lock
}
