package unit_test

import (
	"context"
	"net"
	"testing"

	"github.com/homealias/homealias/backend/internal/dns"
)

type mockProvider struct {
	upserts int
}

func (m *mockProvider) ValidateCredentials(ctx context.Context, token string) ([]dns.Zone, error) {
	return []dns.Zone{{ID: "z1", Name: "exemplo.com"}}, nil
}
func (m *mockProvider) UpsertRecord(ctx context.Context, in dns.UpsertInput) error {
	m.upserts++
	return nil
}
func (m *mockProvider) DeleteRecord(ctx context.Context, in dns.DeleteInput) error { return nil }
func (m *mockProvider) GetRecord(ctx context.Context, in dns.GetInput) (*dns.Record, error) {
	return nil, nil
}

func TestMockProviderCountsUpserts(t *testing.T) {
	m := &mockProvider{}
	_ = m.UpsertRecord(context.Background(), dns.UpsertInput{Content: "1.2.3.4"})
	_ = m.UpsertRecord(context.Background(), dns.UpsertInput{Content: "1.2.3.4"})
	if m.upserts != 2 {
		t.Fatalf("upserts=%d", m.upserts)
	}
	_ = net.ParseIP("198.51.100.7")
}
