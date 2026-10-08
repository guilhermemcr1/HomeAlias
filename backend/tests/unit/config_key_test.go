package unit_test

import (
	"encoding/hex"
	"testing"

	"github.com/homealias/homealias/backend/internal/config"
)

func TestParseKey32RawAndHex(t *testing.T) {
	raw := "0123456789abcdef0123456789abcdef"
	b, err := config.ParseKey32(raw)
	if err != nil || len(b) != 32 {
		t.Fatalf("raw: %v len=%d", err, len(b))
	}
	hx := hex.EncodeToString([]byte(raw))
	b2, err := config.ParseKey32(hx)
	if err != nil || len(b2) != 32 {
		t.Fatalf("hex: %v len=%d", err, len(b2))
	}
}
