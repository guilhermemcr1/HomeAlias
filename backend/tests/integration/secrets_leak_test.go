package integration_test

import (
	"strings"
	"testing"

	"github.com/homealias/homealias/backend/internal/auth"
)

func TestMaskSuffixNeverFullSecret(t *testing.T) {
	secret := "super-secret-cloudflare-token-value"
	masked := auth.MaskSuffix(secret, 4)
	if strings.Contains(masked, secret) {
		t.Fatal("masked contains full secret")
	}
	if !strings.HasPrefix(masked, "…") {
		t.Fatalf("mask=%s", masked)
	}
}
