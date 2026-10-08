package panel

import (
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/homealias/homealias/backend/internal/audit"
	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/store"
	"github.com/jmoiron/sqlx"
)

type AuthAPI struct {
	DB             *sqlx.DB
	Sessions       *store.DBSessionStore
	Idle           time.Duration
	LoginRL        *auth.RateLimiter
	AccountRL      *auth.RateLimiter
	Audit          *audit.Writer
	CookieSecure   bool
	TrustCFHeaders bool
}

// HTTPS keeps Secure cookies even when direct HTTP is enabled for the LAN.
func (a *AuthAPI) secureCookies(r *http.Request) bool {
	return a.CookieSecure || r.TLS != nil || (a.TrustCFHeaders && strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"))
}

func (a *AuthAPI) Login(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	ip := clientHost(r.RemoteAddr)
	email := strings.ToLower(strings.TrimSpace(body.Email))
	// Dois limites: por IP (varredura de várias contas) e por conta (ataque distribuído).
	if ok, retry := a.LoginRL.Allow("login-ip:" + ip); !ok {
		tooMany(w, retry)
		return
	}
	if a.AccountRL != nil {
		if ok, retry := a.AccountRL.Allow("login-acct:" + email); !ok {
			tooMany(w, retry)
			return
		}
	}
	u, err := store.GetUserByEmail(r.Context(), a.DB, body.Email)
	fail := func() {
		_ = a.Audit.Write(r.Context(), audit.Entry{Action: "login_failed", ResourceType: "user", IP: r.RemoteAddr, Summary: "login failed"})
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	}
	if err != nil || u.Status != "active" || !u.PasswordHash.Valid {
		// Gasta o mesmo custo de hash para não revelar, pelo tempo, se o e-mail existe.
		_, _ = auth.VerifyPassword(dummyHash(), body.Password)
		fail()
		return
	}
	ok, err := auth.VerifyPassword(u.PasswordHash.String, body.Password)
	if err != nil || !ok {
		fail()
		return
	}
	if auth.NeedsRehash(u.PasswordHash.String) {
		if nh, herr := auth.HashPassword(body.Password); herr == nil {
			_, _ = a.DB.ExecContext(r.Context(), `UPDATE users SET password_hash=? WHERE id=?`, nh, u.ID)
		}
	}
	sess, err := a.Sessions.Create(u.ID, r.RemoteAddr, a.Idle)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	auth.SetSessionCookie(w, sess.ID, a.Idle, a.secureCookies(r))
	csrf, _ := auth.NewCSRFToken()
	auth.SetCSRFCookie(w, csrf, a.secureCookies(r))
	_, _ = a.DB.ExecContext(r.Context(), `UPDATE users SET last_login_at=? WHERE id=?`, time.Now().UTC(), u.ID)
	actorID := u.ID
	_ = a.Audit.Write(r.Context(), audit.Entry{ActorID: &actorID, ActorRole: u.Role, Action: "login", ResourceType: "user", ResourceID: u.ID, IP: r.RemoteAddr, Summary: "login ok"})
	w.WriteHeader(http.StatusNoContent)
}

func (a *AuthAPI) Logout(w http.ResponseWriter, r *http.Request) {
	id, err := auth.ReadSessionCookie(r)
	if err == nil {
		_ = a.Sessions.Delete(id)
	}
	auth.ClearSessionCookie(w, a.secureCookies(r))
	w.WriteHeader(http.StatusNoContent)
}

func Middleware(sessions *store.DBSessionStore, db *sqlx.DB, idle time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sid, err := auth.ReadSessionCookie(r)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			sess, err := sessions.Get(sid)
			if err == sql.ErrNoRows || err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			u, err := store.GetUserByID(r.Context(), db, sess.UserID)
			if err != nil || u.Status != "active" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			_ = sessions.Touch(sid, idle)
			ctx := auth.WithActor(r.Context(), auth.Actor{ID: u.ID, Role: u.Role, Email: u.Email})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !auth.ValidateCSRF(r) {
			http.Error(w, "csrf", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

var (
	dummyOnce sync.Once
	dummyH    string
)

func dummyHash() string {
	dummyOnce.Do(func() { dummyH, _ = auth.HashPassword("homealias-dummy-password") })
	return dummyH
}

func clientHost(remote string) string {
	if h, _, err := net.SplitHostPort(remote); err == nil {
		return h
	}
	return remote
}

func tooMany(w http.ResponseWriter, retry time.Duration) {
	secs := int(retry.Seconds()) + 1
	w.Header().Set("Retry-After", strconv.Itoa(secs))
	http.Error(w, "Muitas tentativas. Aguarde e tente novamente.", http.StatusTooManyRequests)
}
