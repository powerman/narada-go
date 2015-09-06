package narada

import (
	"syscall"
	"testing"
	"time"
)

func TestSharedLock(t *testing.T) {
	lock1, err := SharedLock(time.Second)
	if err != nil {
		t.Errorf("lock1 = SharedLock(), err = %v", err)
	}
	lock2, err := SharedLock(0)
	if err != nil {
		t.Errorf("lock2 = SharedLock(), err = %v", err)
	}
	if err = lock1.UnLock(); err != nil {
		t.Errorf("lock1.UnLock(), err = %v", err)
	}
	if err = lock2.UnLock(); err != nil {
		t.Errorf("lock2.UnLock(), err = %v", err)
	}
	if err = lock2.UnLock(); err != syscall.EBADF {
		t.Errorf("lock2.UnLock(), err = %v", err)
	}
}
