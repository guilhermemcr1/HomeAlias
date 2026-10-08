package unit_test

import (
	"testing"
	"time"

	"github.com/homealias/homealias/backend/internal/domain"
)

func TestDeriveStatusDefaults(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	seen := now.Add(-10 * time.Minute)
	if got := domain.DeriveStatus(&seen, 900, 3600, now); got != domain.StatusOnline {
		t.Fatalf("want online got %s", got)
	}
	seen = now.Add(-30 * time.Minute)
	if got := domain.DeriveStatus(&seen, 900, 3600, now); got != domain.StatusWarning {
		t.Fatalf("want warning got %s", got)
	}
	seen = now.Add(-2 * time.Hour)
	if got := domain.DeriveStatus(&seen, 900, 3600, now); got != domain.StatusOffline {
		t.Fatalf("want offline got %s", got)
	}
	if got := domain.DeriveStatus(nil, 900, 3600, now); got != domain.StatusNeverSeen {
		t.Fatalf("want never_seen got %s", got)
	}
}

func TestHostLimitsOverride(t *testing.T) {
	w, o := 900, 3600
	host := domain.Host{}
	hw, ho := domain.HostLimits(host, w, o)
	if hw != 900 || ho != 3600 {
		t.Fatalf("defaults")
	}
	ow, oo := 25*3600, 48*3600
	host.WarningAfterSec = &ow
	host.OfflineAfterSec = &oo
	hw, ho = domain.HostLimits(host, w, o)
	if hw != ow || ho != oo {
		t.Fatalf("override failed")
	}
}
