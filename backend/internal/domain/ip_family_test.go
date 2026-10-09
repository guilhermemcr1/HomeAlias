package domain

import (
	"net"
	"testing"
	"time"
)

func TestIPFamilyMatchesEnabledRecord(t *testing.T) {
	for _, tc := range []struct {
		name, ip     string
		a, aaaa      bool
		wantRequired string
	}{
		{"A accepts IPv4", "198.51.100.7", true, false, ""},
		{"A accepts mapped IPv4", "::ffff:198.51.100.7", true, false, ""},
		{"A rejects IPv6", "2001:db8::7", true, false, "ipv4"},
		{"AAAA accepts IPv6", "2001:db8::7", false, true, ""},
		{"AAAA rejects IPv4", "198.51.100.7", false, true, "ipv6"},
		{"dual accepts IPv4", "198.51.100.7", true, true, ""},
		{"dual accepts IPv6", "2001:db8::7", true, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := ipFamilyMismatch(Host{EnableA: tc.a, EnableAAAA: tc.aaaa}, net.ParseIP(tc.ip), time.Now())
			if tc.wantRequired == "" {
				if res != nil {
					t.Fatalf("unexpected rejection: %+v", res)
				}
				return
			}
			if res == nil || res.Success || len(res.RequiredIPFamilies) != 1 || res.RequiredIPFamilies[0] != tc.wantRequired {
				t.Fatalf("expected family %s, got %+v", tc.wantRequired, res)
			}
		})
	}
}
