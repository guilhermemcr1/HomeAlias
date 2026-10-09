package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/homealias/homealias/backend/internal/alert"
	"github.com/homealias/homealias/backend/internal/audit"
	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/clientfiles"
	"github.com/homealias/homealias/backend/internal/config"
	"github.com/homealias/homealias/backend/internal/dns"
	"github.com/homealias/homealias/backend/internal/domain"
	"github.com/homealias/homealias/backend/internal/http/panel"
	upd "github.com/homealias/homealias/backend/internal/http/update"
	"github.com/homealias/homealias/backend/internal/store"
	"github.com/jmoiron/sqlx"
)

func NewRouter(cfg *config.Config, db *sqlx.DB, provider dns.Provider) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(clientIP(cfg.TrustCFHeaders))
	r.Use(securityHeaders)
	r.Use(chimw.Recoverer)
	r.Use(accessLog)

	sessions := &store.DBSessionStore{DB: db}
	auditor := &audit.Writer{DB: db}
	loginRL := auth.NewRateLimiter(10, time.Minute)       // tentativas de login por IP
	loginAcctRL := auth.NewRateLimiter(8, 15*time.Minute) // tentativas de login por conta
	inviteRL := auth.NewRateLimiter(10, time.Minute)      // aceite de convite por IP
	apiRL := auth.NewRateLimiter(240, time.Minute)        // API do painel por IP
	updIPRL := auth.NewRateLimiter(120, time.Minute)      // /update por IP
	updRL := auth.NewRateLimiter(60, time.Minute)         // /update por token

	updateSvc := &domain.UpdateService{DB: db, Provider: provider, EncKey: cfg.EncryptionKey}
	r.Get("/client/{file}", clientfiles.Handler(func(r *http.Request) string { return chi.URLParam(r, "file") }))
	r.With(rateLimitIP(updIPRL, "upd-ip")).Method(http.MethodGet, "/update", &upd.DuckHandler{Svc: updateSvc, TrustCF: cfg.TrustCFHeaders, Limiter: updRL})
	r.With(rateLimitIP(updIPRL, "upd-ip")).Method(http.MethodPost, "/update", &upd.DuckHandler{Svc: updateSvc, TrustCF: cfg.TrustCFHeaders, Limiter: updRL})
	r.With(rateLimitIP(updIPRL, "upd-ip")).Method(http.MethodGet, "/nic/update", &upd.DynHandler{Svc: updateSvc, TrustCF: cfg.TrustCFHeaders, Limiter: updRL})
	r.With(rateLimitIP(updIPRL, "upd-ip")).Method(http.MethodPost, "/nic/update", &upd.DynHandler{Svc: updateSvc, TrustCF: cfg.TrustCFHeaders, Limiter: updRL})

	authAPI := &panel.AuthAPI{DB: db, Sessions: sessions, Idle: cfg.SessionIdle, LoginRL: loginRL, AccountRL: loginAcctRL, Audit: auditor, CookieSecure: cfg.CookieSecure, TrustCFHeaders: cfg.TrustCFHeaders}
	accountAPI := &panel.AccountAPI{DB: db, Sessions: sessions, Audit: auditor, RL: auth.NewRateLimiter(5, time.Minute)}
	usersAPI := &panel.UsersAPI{DB: db, Audit: auditor, Sess: sessions}
	connAPI := &panel.ConnectionsAPI{DB: db, Provider: provider, EncKey: cfg.EncryptionKey, Audit: auditor}
	hostsAPI := &panel.HostsAPI{DB: db, Provider: provider, EncKey: cfg.EncryptionKey, Audit: auditor, WarningAfter: cfg.WarningAfterSec, OfflineAfter: cfg.OfflineAfterSec}
	tokensAPI := &panel.TokensAPI{DB: db, Audit: auditor}
	histAPI := &panel.HistoryAPI{DB: db}
	alertsAPI := &panel.AlertsAPI{DB: db, Telegram: alert.NewTelegram(cfg.TelegramBotToken), SMTP: &alert.SMTP{Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Pass: cfg.SMTPPass, From: cfg.SMTPFrom}}
	auditAPI := &panel.AuditAPI{DB: db}
	settingsAPI := &panel.SettingsAPI{DB: db}

	r.Route("/api", func(api chi.Router) {
		api.Use(rateLimitIP(apiRL, "api"))
		api.Post("/auth/login", authAPI.Login)
		api.With(rateLimitIP(inviteRL, "invite")).Post("/auth/invite/accept", usersAPI.AcceptInvite)

		api.Group(func(priv chi.Router) {
			priv.Use(panel.Middleware(sessions, db, cfg.SessionIdle))
			priv.Use(panel.CSRFMiddleware)
			priv.Post("/auth/logout", authAPI.Logout)
			priv.Get("/auth/me", accountAPI.Me)
			priv.Patch("/auth/me", accountAPI.UpdateProfile)
			priv.Post("/auth/password", accountAPI.ChangePassword)
			priv.Get("/users", usersAPI.List)
			priv.Post("/users", usersAPI.Create)
			priv.Patch("/users/{id}", usersAPI.Update)
			priv.Post("/users/{id}/password", usersAPI.ResetPassword)
			priv.Post("/users/{id}/enable", usersAPI.Enable)
			priv.Post("/users/{id}/disable", usersAPI.Disable)
			priv.Get("/connections", connAPI.List)
			priv.Post("/connections", connAPI.Create)
			priv.Patch("/connections/{id}", connAPI.Patch)
			priv.Delete("/connections/{id}", connAPI.Delete)
			priv.Post("/connections/{id}/test", connAPI.Test)
			priv.Get("/hosts", hostsAPI.List)
			priv.Post("/hosts", hostsAPI.Create)
			priv.Post("/hosts/check", hostsAPI.Check)
			priv.Patch("/hosts/{id}", hostsAPI.Patch)
			priv.Delete("/hosts/{id}", hostsAPI.Delete)
			priv.Post("/hosts/{id}/sync", hostsAPI.Sync)
			priv.Post("/tokens", tokensAPI.Create)
			priv.Post("/tokens/{id}/revoke", tokensAPI.Revoke)
			priv.Get("/history", histAPI.List)
			priv.Get("/alert-channels", alertsAPI.ListChannels)
			priv.Post("/alert-channels", alertsAPI.CreateChannel)
			priv.Post("/alert-channels/{id}/test", alertsAPI.TestChannel)
			priv.Get("/alert-rules", alertsAPI.ListRules)
			priv.Post("/alert-rules", alertsAPI.CreateRule)
			priv.Get("/audit", auditAPI.List)
			priv.Get("/settings/status-defaults", settingsAPI.GetStatusDefaults)
			priv.Put("/settings/status-defaults", settingsAPI.PutStatusDefaults)
		})
	})

	if cfg.StaticDir != "" {
		r.Handle("/*", spaHandler(cfg.StaticDir))
	}
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return r
}

// spaHandler serve os arquivos do build; qualquer caminho que não seja um arquivo
// existente (ex.: /login, /hosts) devolve index.html para o Vue Router resolver.
// Caminhos de API desconhecidos continuam 404.
func spaHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	index := filepath.Join(dir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if info, err := os.Stat(filepath.Join(dir, filepath.Clean("/"+p))); err == nil && !info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}
		// Arquivos de build e rotas de API inexistentes são 404 de verdade, não a página do app.
		if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/assets/") || strings.HasPrefix(p, "/update") || strings.HasPrefix(p, "/nic") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, index)
	})
}
