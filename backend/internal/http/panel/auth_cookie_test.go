package panel

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/homealias/homealias/backend/internal/auth"
)

func TestCookieTransportPolicy(t *testing.T) {
	for _, tc := range []struct {
		name                string
		enforce, trust, tls bool
		forwarded           string
		secure              bool
	}{
		{name: "HTTP LAN", secure: false},
		{name: "HTTPS only", enforce: true, secure: true},
		{name: "direct HTTPS", tls: true, secure: true},
		{name: "Cloudflare HTTPS", trust: true, forwarded: "https", secure: true},
		{name: "untrusted proxy", forwarded: "https", secure: false},
		{name: "Cloudflare HTTP", trust: true, forwarded: "http", secure: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := &AuthAPI{CookieSecure: tc.enforce, TrustCFHeaders: tc.trust}
			req := httptest.NewRequest("GET", "http://10.0.0.181:8080", nil)
			if tc.tls {
				req.TLS = &tls.ConnectionState{}
			}
			req.Header.Set("X-Forwarded-Proto", tc.forwarded)
			secure := api.secureCookies(req)
			if secure != tc.secure {
				t.Fatalf("Secure=%v, want %v", secure, tc.secure)
			}
			w := httptest.NewRecorder()
			auth.SetSessionCookie(w, "session", time.Hour, secure)
			auth.SetCSRFCookie(w, "csrf", secure)
			auth.ClearSessionCookie(w, secure)
			cookies := w.Result().Cookies()
			if len(cookies) != 3 {
				t.Fatalf("cookies=%d", len(cookies))
			}
			for _, c := range cookies {
				if c.Secure != tc.secure || c.Path != "/" {
					t.Fatalf("unexpected cookie policy: %#v", c)
				}
				if c.Name == auth.SessionCookieName && !c.HttpOnly {
					t.Fatal("session must remain HttpOnly")
				}
			}
			if cookies[2].MaxAge != -1 {
				t.Fatal("logout must expire the cookie")
			}
			req.AddCookie(cookies[1])
			req.Header.Set(auth.CSRFHeaderName, "csrf")
			if !auth.ValidateCSRF(req) {
				t.Fatal("valid CSRF rejected")
			}
			req.Header.Set(auth.CSRFHeaderName, "wrong")
			if auth.ValidateCSRF(req) {
				t.Fatal("invalid CSRF accepted")
			}
		})
	}
}
