package domain

import "time"

const (
	StatusOnline     = "online"
	StatusWarning    = "warning"
	StatusOffline    = "offline"
	StatusNeverSeen  = "never_seen"
)

func DeriveStatus(lastSeen *time.Time, warningAfter, offlineAfter int, now time.Time) string {
	if lastSeen == nil {
		return StatusNeverSeen
	}
	age := now.Sub(*lastSeen)
	if age <= time.Duration(warningAfter)*time.Second {
		return StatusOnline
	}
	if age <= time.Duration(offlineAfter)*time.Second {
		return StatusWarning
	}
	return StatusOffline
}

func HostLimits(h Host, globalWarning, globalOffline int) (warning, offline int) {
	warning = globalWarning
	offline = globalOffline
	if h.WarningAfterSec != nil {
		warning = *h.WarningAfterSec
	}
	if h.OfflineAfterSec != nil {
		offline = *h.OfflineAfterSec
	}
	return warning, offline
}

func LatestSeen(h Host) *time.Time {
	switch {
	case h.LastSeenV4 != nil && h.LastSeenV6 != nil:
		if h.LastSeenV4.After(*h.LastSeenV6) {
			return h.LastSeenV4
		}
		return h.LastSeenV6
	case h.LastSeenV4 != nil:
		return h.LastSeenV4
	default:
		return h.LastSeenV6
	}
}
