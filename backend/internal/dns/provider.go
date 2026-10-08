package dns

import "context"

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Record struct {
	ID      string
	Type    string // A | AAAA
	Name    string
	Content string
	TTL     int
	Proxied bool
}

type UpsertInput struct {
	ZoneID  string
	Type    string
	Name    string // FQDN
	Content string
	TTL     int
	Proxied bool
}

type DeleteInput struct {
	ZoneID string
	Type   string
	Name   string
}

type GetInput struct {
	ZoneID string
	Type   string
	Name   string
}

type Provider interface {
	ValidateCredentials(ctx context.Context, token string) ([]Zone, error)
	UpsertRecord(ctx context.Context, in UpsertInput) error
	DeleteRecord(ctx context.Context, in DeleteInput) error
	GetRecord(ctx context.Context, in GetInput) (*Record, error)
}
