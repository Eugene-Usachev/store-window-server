package e2e_tests

import "sync"

var testLock sync.RWMutex

func AcquireTestLockShared() {
	testLock.RLock()
}

func ReleaseTestLockShared() {
	testLock.RUnlock()
}

func AcquireTestLockExclusive() {
	testLock.Lock()
}

func ReleaseTestLockExclusive() {
	testLock.Unlock()
}
