package contract_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/domain"
	upd "github.com/homealias/homealias/backend/internal/http/update"
)

func TestUpdateRequiresTrustedIP(t *testing.T) {
	h := &upd.DuckHandler{
		Svc:     &domain.UpdateService{},
		TrustCF: true,
		Limiter: auth.NewRateLimiter(100, time.Minute),
	}
	req := httptest.NewRequest(http.MethodGet, "/update?token=x", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["success"] != false {
		t.Fatalf("body=%v", body)
	}
}
