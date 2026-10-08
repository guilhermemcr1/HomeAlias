package unit_test

import (
	"strings"
	"testing"

	"github.com/homealias/homealias/backend/internal/validate"
)

func TestEmail(t *testing.T) {
	ok, err := validate.Email("  Ana@Exemplo.COM ")
	if err != nil || ok != "ana@exemplo.com" {
		t.Fatalf("got %q, %v", ok, err)
	}
	for _, bad := range []string{"", "abc", "a@b", "Ana <a@b.com>", "a@b.com\nBcc: x@y.com", strings.Repeat("a", 250) + "@b.com"} {
		if _, err := validate.Email(bad); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestNameLimitsAndControlChars(t *testing.T) {
	if _, err := validate.Name("Nome", "   "); err == nil {
		t.Error("blank must fail")
	}
	if _, err := validate.Name("Nome", strings.Repeat("é", 81)); err == nil {
		t.Error("81 runes must fail")
	}
	if _, err := validate.Name("Nome", "a\x00b"); err == nil {
		t.Error("control chars must fail")
	}
	if got, err := validate.Name("Nome", " Conta 🏠 "); err != nil || got != "Conta 🏠" {
		t.Errorf("emoji name: %q %v", got, err)
	}
}

func TestPassword(t *testing.T) {
	if validate.Password("abcD3fgh1jk") == nil || validate.Password(strings.Repeat("ab1", 50)+"xyzw") == nil {
		t.Error("length bounds not enforced")
	}
	for _, bad := range []string{"aaaaaaaaaaaaaa", "            ", "Change-Me-Now", "123456789012", "password12345"} {
		if validate.Password(bad) == nil {
			t.Errorf("%q deveria ser recusada", bad)
		}
	}
	for _, good := range []string{"cavalo-bateria-grampo-9", "Xk29!vLq#m8Rt"} {
		if err := validate.Password(good); err != nil {
			t.Errorf("%q deveria passar: %v", good, err)
		}
	}
}

func TestAPIToken(t *testing.T) {
	good := strings.Repeat("a", 40)
	if _, err := validate.APIToken(" " + good + " "); err != nil {
		t.Errorf("trimmed token should pass: %v", err)
	}
	for _, bad := range []string{"", "curto", "abc def " + good, strings.Repeat("a", 300)} {
		if _, err := validate.APIToken(bad); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestHostName(t *testing.T) {
	for _, ok := range []string{"casa", "Casa", "a.b", "@", "nas-01"} {
		if _, err := validate.HostName(ok); err != nil {
			t.Errorf("%q should pass: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "-a", "a-", "a..b", "a b", "a_b", "café", strings.Repeat("a", 64), "a/b"} {
		if _, err := validate.HostName(bad); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestTTLAndZone(t *testing.T) {
	if v, err := validate.TTL(0); err != nil || v != 1 {
		t.Error("0 maps to auto")
	}
	if _, err := validate.TTL(30); err == nil {
		t.Error("30 must fail")
	}
	if _, err := validate.ZoneName("localhost"); err == nil {
		t.Error("single label zone must fail")
	}
	if z, err := validate.ZoneName("Exemplo.com"); err != nil || z != "exemplo.com" {
		t.Error("zone normalised")
	}
}

func TestChannelDestination(t *testing.T) {
	cases := []struct {
		kind, dest string
		ok         bool
	}{
		{"email", "a@b.com", true}, {"email", "x", false},
		{"telegram", "123456789", true}, {"telegram", "-1001234567890", true}, {"telegram", "@meucanal", true},
		{"telegram", "abc", false}, {"sms", "1", false},
	}
	for _, c := range cases {
		_, err := validate.ChannelDestination(c.kind, c.dest)
		if (err == nil) != c.ok {
			t.Errorf("%s %q: ok=%v err=%v", c.kind, c.dest, c.ok, err)
		}
	}
}

func TestAPITokenNormalizesPasteArtifacts(t *testing.T) {
	core := strings.Repeat("aB3_-", 8)
	for _, in := range []string{"Bearer " + core, `"` + core + `"`, " " + core + "\n"} {
		got, err := validate.APIToken(in)
		if err != nil || got != core {
			t.Errorf("%q -> %q, %v", in, got, err)
		}
	}
	if _, err := validate.APIToken(strings.Repeat("a", 39) + "!"); err == nil {
		t.Error("invalid chars must be rejected")
	}
}
