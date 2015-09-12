// Package bootstrap provides Narada bootstrap lock logic.
//
// Try to acquire bootstrap lock while init().
// Wait for lock for $NARADA_BOOTSTRAP_TIMEOUT seconds (float, 15.0 by default),
// and then either terminate current program (default) or do nothing
// (if $NARADA_BOOTSTRAP_GRACEFUL set to non-empty value).
package bootstrap

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/powerman/narada-go/narada"
)

var lock *narada.Lock

const defaultTimeout = 15 * time.Second

func init() {
	if err := initLock(); err != nil {
		log.Fatalf("can't get bootstrap lock: %v", err)
	}
}

func initLock() error {
	var timeout = defaultTimeout
	if val := os.Getenv("NARADA_BOOTSTRAP_TIMEOUT"); val != "" {
		t, err := strconv.ParseFloat(val, 64)
		if err != nil || t < 0 {
			log.Printf("bad NARADA_BOOTSTRAP_TIMEOUT=%q", val)
		} else {
			timeout = time.Duration(t * float64(time.Second))
		}
	}
	var isGraceful bool
	if val := os.Getenv("NARADA_BOOTSTRAP_GRACEFUL"); val != "" {
		isGraceful = true
	}

	if err := Lock(timeout); err != nil && !isGraceful {
		return err
	}
	return nil
}

// Lock try to acquire bootstrap lock.
func Lock(wait time.Duration) error {
	if lock != nil {
		return errors.New("bootstrap lock already acquired")
	}
	if l, err := narada.SharedLock(wait); err == nil {
		lock = &l
	} else {
		return err
	}
	return nil
}

// UnLock release bootstrap lock.
func Unlock() error {
	if lock == nil {
		return errors.New("bootstrap lock not acquired")
	}
	if err := lock.UnLock(); err != nil {
		return err
	}
	lock = nil
	return nil
}

// HasLock returns true if bootstrap lock currently acquired.
func HasLock() bool {
	return lock != nil
}
