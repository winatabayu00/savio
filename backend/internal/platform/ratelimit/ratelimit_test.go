package ratelimit

import (
	"testing"
	"time"
)

func TestAllowWindow(t *testing.T) {
	l := New(3, time.Minute)
	now := time.Now()
	if !l.Allow("k", now) || !l.Allow("k", now) || !l.Allow("k", now) {
		t.Fatal("expected first three to pass")
	}
	if l.Allow("k", now) {
		t.Fatal("expected fourth to be rejected")
	}
	// new window resets
	if !l.Allow("k", now.Add(2*time.Minute)) {
		t.Fatal("expected reset window to allow")
	}
	// separate key unaffected
	if !l.Allow("other", now) {
		t.Fatal("expected other key to be allowed")
	}
}
