package update

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/domain"
)

type DuckHandler struct {
	Svc      *domain.UpdateService
	TrustCF  bool
	Limiter  *auth.RateLimiter
}

func (h *DuckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	token := r.Form.Get("token")
	if token == "" {
		authz := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			token = strings.TrimSpace(authz[7:])
		}
	}
	hostname := r.Form.Get("hostname")
	ip, ok := ClientIP(r, h.TrustCF)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, domain.UpdateResult{Success: false, Code: "unauthorized", Message: "trusted client ip required"})
		return
	}
	if HasClientIPParams(r) {
		// ignore ip=/myip= — never use them
	}
	if okAllow, retry := h.Limiter.Allow("upd:" + auth.HashDDNSToken(token)); !okAllow {
		w.Header().Set("Retry-After", strings.TrimSuffix(retry.String(), "s"))
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"success": false, "code": "rate_limited", "message": "rate limited",
			"retry_after_sec": int(retry.Seconds()),
		})
		return
	}
	res := h.Svc.Handle(r.Context(), domain.UpdateRequest{
		TokenRaw:   token,
		Hostname:   hostname,
		ClientIP:   ip,
		ClientType: DetectClientType(r, r.URL.Path),
	})
	status := http.StatusOK
	if !res.Success {
		switch res.Code {
		case "unauthorized":
			status = http.StatusUnauthorized
		case "rate_limited":
			status = http.StatusTooManyRequests
		case "temporary":
			status = http.StatusServiceUnavailable
		case "not_found", "bad_request":
			status = http.StatusBadRequest
		default:
			status = http.StatusBadRequest
		}
	}
	writeJSON(w, status, res)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
