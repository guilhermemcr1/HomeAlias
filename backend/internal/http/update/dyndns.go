package update

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/domain"
)

type DynHandler struct {
	Svc     *domain.UpdateService
	TrustCF bool
	Limiter *auth.RateLimiter
}

func (h *DynHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	user, pass, ok := basicAuth(r)
	_ = user
	if !ok || pass == "" {
		writeText(w, "badauth")
		return
	}
	hostname := r.Form.Get("hostname")
	ip, okIP := ClientIP(r, h.TrustCF)
	if !okIP {
		writeText(w, "badauth")
		return
	}
	if okAllow, _ := h.Limiter.Allow("upd:" + auth.HashDDNSToken(pass)); !okAllow {
		writeText(w, "abuse")
		return
	}
	res := h.Svc.Handle(r.Context(), domain.UpdateRequest{
		TokenRaw:   pass,
		Hostname:   hostname,
		ClientIP:   ip,
		ClientType: "dyndns",
	})
	if res.DynText != "" {
		writeText(w, res.DynText)
		return
	}
	writeText(w, "911")
}

func writeText(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func basicAuth(r *http.Request) (username, password string, ok bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Basic ") {
		return "", "", false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(h, "Basic "))
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
