// Package narada provides integration with Narada framework.
package narada

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

const lockfile = ".lock"
const locknew = ".lock.new"
const tick = time.Millisecond * 100

var ErrLockTimeout = errors.New("failed to acquire lock: timed out")

// Lock is Narada lock.
type Lock struct {
	f *os.File
}

// SharedLock try to get shared lock which is required to modify any
// project data (files or database).
//
// If wait <= 0 will wait forever until lock will be granted.
func SharedLock(wait time.Duration) (l Lock, err error) {
	var waited time.Duration
	if os.Getenv("NARADA_SKIP_LOCK") != "" {
		return
	}
	if l.f, err = os.OpenFile(lockfile, os.O_RDONLY|os.O_CREATE, 0644); err != nil {
		return
	}
	for {
		_, err2 := os.Stat(locknew)
		if err2 != nil && os.IsNotExist(err2) {
			err = unix.Flock(int(l.f.Fd()), unix.LOCK_SH|unix.LOCK_NB)
			if err == nil || err != unix.EWOULDBLOCK {
				return
			}
			err = nil
		}
		if wait > 0 && waited >= wait {
			return l, ErrLockTimeout
		}
		time.Sleep(tick)
		waited += tick
	}
}

// UnLock free lock set by SharedLock() (or second lock set by ExclusiveLock()).
func (l Lock) UnLock() error {
	if os.Getenv("NARADA_SKIP_LOCK") != "" {
		return nil
	}
	if l.f == nil {
		return nil
	}
	if err := unix.Flock(int(l.f.Fd()), unix.LOCK_UN); err != nil {
		return err
	}
	if err := l.f.Close(); err != nil {
		return err
	}
	l.f = nil
	return nil
}
