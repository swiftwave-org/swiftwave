package rest

import (
	"testing"
	"time"
)

func TestThrottleAllowsAttemptsBelowThreshold(t *testing.T) {
	throttle := newLoginThrottle()
	for range loginFailureThreshold - 1 {
		throttle.fail("user:alice", loginFailureThreshold, loginBaseLockout, loginMaxLockout)
	}
	if wait := throttle.lockedFor("user:alice"); wait > 0 {
		t.Fatalf("should not be locked before the threshold, got %s", wait)
	}
}

func TestThrottleLocksAtThreshold(t *testing.T) {
	throttle := newLoginThrottle()
	for range loginFailureThreshold {
		throttle.fail("user:alice", loginFailureThreshold, loginBaseLockout, loginMaxLockout)
	}
	wait := throttle.lockedFor("user:alice")
	if wait <= 0 {
		t.Fatal("should be locked once the threshold is reached")
	}
	if wait > loginBaseLockout {
		t.Fatalf("first lockout should be the base duration, got %s", wait)
	}
}

func TestThrottleBacksOffExponentiallyUpToCap(t *testing.T) {
	throttle := newLoginThrottle()
	for range loginFailureThreshold {
		throttle.fail("user:alice", loginFailureThreshold, loginBaseLockout, loginMaxLockout)
	}
	previous := throttle.lockedFor("user:alice")
	for range 3 {
		throttle.fail("user:alice", loginFailureThreshold, loginBaseLockout, loginMaxLockout)
		current := throttle.lockedFor("user:alice")
		if current <= previous {
			t.Fatalf("lockout should grow with each failure, %s did not exceed %s", current, previous)
		}
		previous = current
	}
	for range 20 {
		throttle.fail("user:alice", loginFailureThreshold, loginBaseLockout, loginMaxLockout)
	}
	if wait := throttle.lockedFor("user:alice"); wait > loginMaxLockout {
		t.Fatalf("lockout %s exceeded the cap of %s", wait, loginMaxLockout)
	}
}

// A six digit code has a million possibilities, the point of the tighter budget is that
// an attacker cannot get anywhere near that many tries inside one validity window.
func TestTotpBudgetIsTighterThanPassword(t *testing.T) {
	if totpFailureThreshold >= loginFailureThreshold {
		t.Fatal("totp should allow fewer attempts than a password")
	}
	throttle := newLoginThrottle()
	for range totpFailureThreshold {
		throttle.fail("totp:user:alice", totpFailureThreshold, totpBaseLockout, totpMaxLockout)
	}
	wait := throttle.lockedFor("totp:user:alice")
	if wait < 30*time.Second {
		t.Fatalf("totp lockout %s is too short to blunt a brute force", wait)
	}
}

func TestThrottleIsPerKey(t *testing.T) {
	throttle := newLoginThrottle()
	for range loginFailureThreshold {
		throttle.fail("user:alice", loginFailureThreshold, loginBaseLockout, loginMaxLockout)
	}
	if wait := throttle.lockedFor("user:bob"); wait > 0 {
		t.Fatal("locking one account must not lock another")
	}
}

func TestLockedForReportsLongestLock(t *testing.T) {
	throttle := newLoginThrottle()
	for range loginFailureThreshold {
		throttle.fail("ip:10.0.0.1", loginFailureThreshold, loginBaseLockout, loginMaxLockout)
	}
	for range totpFailureThreshold + 4 {
		throttle.fail("totp:ip:10.0.0.1", totpFailureThreshold, totpBaseLockout, totpMaxLockout)
	}
	combined := throttle.lockedFor("ip:10.0.0.1", "totp:ip:10.0.0.1")
	if combined < throttle.lockedFor("totp:ip:10.0.0.1") {
		t.Fatal("a locked key must not be masked by a less locked one")
	}
}

func TestSuccessClearsFailures(t *testing.T) {
	throttle := newLoginThrottle()
	for range loginFailureThreshold {
		throttle.fail("user:alice", loginFailureThreshold, loginBaseLockout, loginMaxLockout)
	}
	throttle.clear("user:alice")
	if wait := throttle.lockedFor("user:alice"); wait > 0 {
		t.Fatalf("a successful login should reset the counter, still locked for %s", wait)
	}
}

func TestThrottleIsSafeUnderConcurrency(t *testing.T) {
	throttle := newLoginThrottle()
	done := make(chan struct{})
	for i := range 50 {
		go func() {
			defer func() { done <- struct{}{} }()
			throttle.fail("user:alice", loginFailureThreshold, loginBaseLockout, loginMaxLockout)
			throttle.lockedFor("user:alice", "ip:10.0.0.1")
			if i%10 == 0 {
				throttle.clear("ip:10.0.0.1")
			}
		}()
	}
	for range 50 {
		<-done
	}
}
