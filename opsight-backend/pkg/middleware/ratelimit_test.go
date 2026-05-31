package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.Allow("192.168.1.1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	rl.Allow("10.0.0.1")
	rl.Allow("10.0.0.1")

	if rl.Allow("10.0.0.1") {
		t.Error("third request should be blocked")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	if !rl.Allow("1.1.1.1") {
		t.Error("first IP should be allowed")
	}
	if !rl.Allow("2.2.2.2") {
		t.Error("second IP should be allowed")
	}
	if rl.Allow("1.1.1.1") {
		t.Error("first IP should now be blocked")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)

	if !rl.Allow("10.0.0.1") {
		t.Error("first request should be allowed")
	}
	if rl.Allow("10.0.0.1") {
		t.Error("second request should be blocked")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("10.0.0.1") {
		t.Error("request after window should be allowed")
	}
}
