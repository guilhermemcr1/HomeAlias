package panel

import (
	"strings"
	"testing"

	"github.com/homealias/homealias/backend/internal/domain"
)

func TestClientInstructionsMatchHostFamilies(t *testing.T) {
	for _, tc := range []struct {
		name    string
		a, aaaa bool
	}{
		{"IPv4", true, false}, {"IPv6", false, true}, {"dual", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			instructions := clientInstructions("dummy-token", []domain.Host{{FQDN: "test.example.com", EnableA: tc.a, EnableAAAA: tc.aaaa}})
			for _, key := range []string{"curl", "docker"} {
				if strings.Contains(instructions[key], "curl -4 ") != tc.a || strings.Contains(instructions[key], "curl -6 ") != tc.aaaa {
					t.Fatalf("%s: incorrect families: %s", key, instructions[key])
				}
			}
		})
	}
}
