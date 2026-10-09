package panel

import (
	"context"
	"errors"
	"testing"

	"github.com/homealias/homealias/backend/internal/dns"
	"github.com/homealias/homealias/backend/internal/domain"
)

type hostDNSFake struct {
	dns.Provider
	existing map[string]*dns.Record
	inputs   []dns.UpsertInput
	failA    bool
}

func (f *hostDNSFake) GetRecord(_ context.Context, in dns.GetInput) (*dns.Record, error) {
	return f.existing[in.Type], nil
}
func (f *hostDNSFake) UpsertRecord(_ context.Context, in dns.UpsertInput) error {
	f.inputs = append(f.inputs, in)
	if f.failA && in.Type == "A" {
		return errors.New("upstream unavailable")
	}
	return nil
}

func TestSyncDNSProxyAndTTL(t *testing.T) {
	for _, proxied := range []bool{false, true} {
		v4, v6 := "198.51.100.7", "2001:db8::7"
		fake := &hostDNSFake{}
		host := domain.Host{FQDN: "test.example.com", EnableA: true, EnableAAAA: true, LastIPv4: &v4, LastIPv6: &v6, Proxied: proxied, TTL: 600}
		if err := syncDNSRecords(context.Background(), fake, host); err != nil {
			t.Fatal(err)
		}
		if len(fake.inputs) != 2 {
			t.Fatal("both record families must be synchronized")
		}
		wantTTL := 600
		if proxied {
			wantTTL = 1
		}
		for _, in := range fake.inputs {
			if in.Proxied != proxied || in.TTL != wantTTL {
				t.Fatalf("incorrect DNS options: %+v", in)
			}
		}
	}
}

func TestSyncDNSPreservesAdoptedContent(t *testing.T) {
	fake := &hostDNSFake{existing: map[string]*dns.Record{"A": {Content: "198.51.100.8"}}}
	err := syncDNSRecords(context.Background(), fake, domain.Host{FQDN: "test.example.com", EnableA: true, Proxied: true, TTL: 1})
	if err != nil || len(fake.inputs) != 1 || fake.inputs[0].Content != "198.51.100.8" {
		t.Fatalf("adopted record: %v, %+v", err, fake.inputs)
	}
}

func TestSyncDNSReportsPartialFailureAndAttemptsBoth(t *testing.T) {
	v4, v6 := "198.51.100.7", "2001:db8::7"
	fake := &hostDNSFake{failA: true}
	err := syncDNSRecords(context.Background(), fake, domain.Host{EnableA: true, EnableAAAA: true, LastIPv4: &v4, LastIPv6: &v6, TTL: 1})
	if err == nil || len(fake.inputs) != 2 {
		t.Fatalf("failure must be reported after attempting both families: %v", err)
	}
}

func TestSyncDNSRequiresFirstUpdateWhenNoRecordExists(t *testing.T) {
	err := syncDNSRecords(context.Background(), &hostDNSFake{}, domain.Host{EnableA: true, TTL: 1})
	if !errors.Is(err, errNoDNSRecord) {
		t.Fatalf("expected first-update guidance, got %v", err)
	}
}
