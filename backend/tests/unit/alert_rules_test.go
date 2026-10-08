package unit_test

import (
	"testing"
	"time"
)

func tokenExpiringShouldFire(expiresAt *time.Time, leadDays int, now time.Time) bool {
	if expiresAt == nil {
		return false
	}
	lead := now.Add(time.Duration(leadDays) * 24 * time.Hour)
	return !expiresAt.After(lead) && expiresAt.After(now)
}

func TestTokenExpiringDefault14Days(t *testing.T) {
	now := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	if tokenExpiringShouldFire(nil, 14, now) {
		t.Fatal("nil expiry must not fire")
	}
	in10 := now.Add(10 * 24 * time.Hour)
	if !tokenExpiringShouldFire(&in10, 14, now) {
		t.Fatal("10 days should fire with 14 lead")
	}
	in20 := now.Add(20 * 24 * time.Hour)
	if tokenExpiringShouldFire(&in20, 14, now) {
		t.Fatal("20 days should not fire")
	}
}
