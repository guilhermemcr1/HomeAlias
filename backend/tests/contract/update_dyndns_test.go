package contract_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/domain"
	upd "github.com/homealias/homealias/backend/internal/http/update"
)

func TestDynDNSBadauthWithoutBasic(t *testing.T) {
	h := &upd.DynHandler{
		Svc:     &domain.UpdateService{},
		TrustCF: true,
		Limiter: auth.NewRateLimiter(100, time.Minute),
	}
	req := httptest.NewRequest(http.MethodGet, "/nic/update?hostname=casa.exemplo.com", nil)
	req.Header.Set("CF-Connecting-IP", "198.51.100.7")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Body.String() != "badauth" {
		t.Fatalf("body=%q", rr.Body.String())
	}
}
