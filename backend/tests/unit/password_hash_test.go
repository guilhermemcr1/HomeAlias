package unit

import (
	"testing"
	"time"

	"github.com/homealias/homealias/backend/internal/auth"
)

func TestPasswordHashRoundTripAndRehash(t *testing.T) {
	h, err := auth.HashPassword("cavalo-bateria-grampo-9")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := auth.VerifyPassword(h, "cavalo-bateria-grampo-9"); err != nil || !ok {
		t.Fatalf("senha correta recusada: %v", err)
	}
	if ok, _ := auth.VerifyPassword(h, "outra"); ok {
		t.Fatal("senha errada aceita")
	}
	if auth.NeedsRehash(h) {
		t.Fatal("hash atual não deve pedir rehash")
	}
	// hash legado (t=1) deve pedir rehash, e parâmetros absurdos devem ser recusados.
	legacy := "$argon2id$v=19$m=65536,t=1,p=4$c2FsdHNhbHRzYWx0c2FsdA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	if !auth.NeedsRehash(legacy) {
		t.Fatal("hash legado deveria pedir rehash")
	}
	huge := "$argon2id$v=19$m=4000000,t=1,p=4$c2FsdHNhbHRzYWx0c2FsdA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	if _, err := auth.VerifyPassword(huge, "x"); err == nil {
		t.Fatal("parâmetros fora do limite deveriam ser recusados")
	}
}

func TestRateLimiterBlocksAndResets(t *testing.T) {
	rl := auth.NewRateLimiter(3, 50*time.Millisecond)
	for i := 0; i < 3; i++ {
		if ok, _ := rl.Allow("k"); !ok {
			t.Fatalf("tentativa %d deveria passar", i)
		}
	}
	if ok, retry := rl.Allow("k"); ok || retry <= 0 {
		t.Fatal("4ª tentativa deveria ser bloqueada com retry > 0")
	}
	if ok, _ := rl.Allow("outro"); !ok {
		t.Fatal("chaves diferentes não devem interferir")
	}
	time.Sleep(60 * time.Millisecond)
	if ok, _ := rl.Allow("k"); !ok {
		t.Fatal("janela deveria ter reiniciado")
	}
}
