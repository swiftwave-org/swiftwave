package rest

import (
	"sync"
	"time"
)

const (
	// password attempts, counted per username and per source ip
	loginFailureThreshold = 5
	loginBaseLockout      = 30 * time.Second
	loginMaxLockout       = 15 * time.Minute
	// a totp code is only six digits, so it gets a far smaller budget than a password
	totpFailureThreshold = 3
	totpBaseLockout      = 5 * time.Minute
	totpMaxLockout       = 1 * time.Hour
	// a key with no recent failure is forgotten
	loginRecordTTL = 1 * time.Hour
)

type loginThrottle struct {
	mutex     sync.Mutex
	records   map[string]*loginAttemptRecord
	lastSweep time.Time
}

type loginAttemptRecord struct {
	failures    int
	lockedUntil time.Time
	lastFailure time.Time
}

func newLoginThrottle() *loginThrottle {
	return &loginThrottle{records: make(map[string]*loginAttemptRecord)}
}

// lockedFor reports how long the caller must wait before another attempt is accepted,
// zero when none of the keys is locked.
func (t *loginThrottle) lockedFor(keys ...string) time.Duration {
	now := time.Now()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	var longest time.Duration
	for _, key := range keys {
		record, ok := t.records[key]
		if !ok {
			continue
		}
		if remaining := record.lockedUntil.Sub(now); remaining > longest {
			longest = remaining
		}
	}
	return longest
}

// fail counts a failed attempt against key and locks it once the threshold is crossed,
// backing off exponentially with every further failure.
func (t *loginThrottle) fail(key string, threshold int, base, max time.Duration) {
	now := time.Now()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.sweep(now)
	record, ok := t.records[key]
	if !ok {
		record = &loginAttemptRecord{}
		t.records[key] = record
	}
	record.failures++
	record.lastFailure = now
	if record.failures < threshold {
		return
	}
	lockout := base
	for range record.failures - threshold {
		lockout *= 2
		if lockout >= max {
			lockout = max
			break
		}
	}
	record.lockedUntil = now.Add(lockout)
}

// clear forgets the failures recorded against keys, called once an attempt succeeds.
func (t *loginThrottle) clear(keys ...string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for _, key := range keys {
		delete(t.records, key)
	}
}

// sweep drops records that are both unlocked and idle, callers must hold the mutex.
func (t *loginThrottle) sweep(now time.Time) {
	if now.Sub(t.lastSweep) < loginRecordTTL {
		return
	}
	t.lastSweep = now
	for key, record := range t.records {
		if now.After(record.lockedUntil) && now.Sub(record.lastFailure) > loginRecordTTL {
			delete(t.records, key)
		}
	}
}
