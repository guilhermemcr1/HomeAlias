package panel

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/homealias/homealias/backend/internal/dns"
	"github.com/homealias/homealias/backend/internal/domain"
)

func TestReplacementTokenMustCoverLinkedZones(t *testing.T) {
	for _, tc := range []struct {
		name      string
		available []dns.Zone
		required  []string
		want      bool
	}{
		{"unused connection", nil, nil, true},
		{"all linked zones", []dns.Zone{{ID: "one"}, {ID: "two"}}, []string{"one", "two"}, true},
		{"missing linked zone", []dns.Zone{{ID: "one"}}, []string{"one", "two"}, false},
		{"revoked access", nil, []string{"one"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if coversZones(tc.available, tc.required) != tc.want {
				t.Fatal("incorrect zone coverage")
			}
		})
	}
}

type renameDNSFake struct {
	hostDNSFake
	records map[string]*dns.Record
}

func (f *renameDNSFake) GetRecord(_ context.Context, in dns.GetInput) (*dns.Record, error) {
	return f.records[in.Name+"/"+in.Type], nil
}

func TestRenamePreservesExistingAddressAndDNSOptions(t *testing.T) {
	fake := &renameDNSFake{records: map[string]*dns.Record{"old.example.com/A": {Content: "198.51.100.7"}, "old.example.com/AAAA": {Content: "2001:db8::7"}}}
	old := domain.Host{ZoneID: "zone", FQDN: "old.example.com"}
	next := domain.Host{ZoneID: "zone", FQDN: "new.example.com", EnableA: true, EnableAAAA: true, Proxied: true, TTL: 1}
	if err := prepareRename(context.Background(), fake, old, &next); err != nil {
		t.Fatal(err)
	}
	if err := syncDNSRecords(context.Background(), fake, next); err != nil {
		t.Fatal(err)
	}
	if len(fake.inputs) != 2 {
		t.Fatal("expected two new records")
	}
	for _, in := range fake.inputs {
		if in.Name != next.FQDN || !in.Proxied || in.TTL != 1 {
			t.Fatalf("incorrect rename input: %+v", in)
		}
	}
	if fake.records[old.FQDN+"/A"].Content != "198.51.100.7" {
		t.Fatal("old record was changed")
	}
}

func TestRenameRejectsOccupiedNameBeforeAnyWrite(t *testing.T) {
	fake := &renameDNSFake{records: map[string]*dns.Record{"new.example.com/AAAA": {Content: "2001:db8::8"}}}
	next := domain.Host{FQDN: "new.example.com", EnableA: true}
	err := prepareRename(context.Background(), fake, domain.Host{FQDN: "old.example.com"}, &next)
	if !errors.Is(err, errDNSNameTaken) || len(fake.inputs) != 0 {
		t.Fatalf("collision must reject without writes: %v", err)
	}
}

func TestConnectionResponseDoesNotExposeEncryptedSecret(t *testing.T) {
	encoded, err := json.Marshal(domain.Connection{APITokenCipher: "sensitive-ciphertext", APITokenSuffix: "…abcd"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "sensitive-ciphertext") {
		t.Fatal("ciphertext exposed in response")
	}
}
