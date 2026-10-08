package httpapi

import (
	"github.com/homealias/homealias/backend/internal/auth"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// clientIP substitui o chimw.RealIP: só confia em CF-Connecting-IP quando a
// instância está atrás do Cloudflare. X-Forwarded-For/X-Real-IP enviados por
// qualquer cliente permitiriam burlar o rate limit e forjar o IP da auditoria.
func clientIP(trustCF bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustCF {
				if ip := net.ParseIP(strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))); ip != nil {
					r.RemoteAddr = net.JoinHostPort(ip.String(), "0")
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// accessLog registra método, caminho, status e duração. Não loga a query string:
// o endpoint /update aceita ?token=..., que não pode ir para os logs.
func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}
		log.Printf("[%s] %s %s from %s - %d %dB in %s", chimw.GetReqID(r.Context()), r.Method, r.URL.Path, r.RemoteAddr, status, ww.BytesWritten(), time.Since(start).Round(time.Microsecond))
	})
}

const csp = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; " +
	"font-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", csp)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Strict-Transport-Security", "max-age=31536000")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func remoteHost(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

// rateLimitIP limita requisições por IP (o IP real, via clientIP) em um grupo de rotas.
func rateLimitIP(rl *auth.RateLimiter, scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ok, retry := rl.Allow(scope + ":" + remoteHost(r.RemoteAddr)); !ok {
				w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())+1))
				http.Error(w, "Muitas requisições. Aguarde e tente novamente.", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
