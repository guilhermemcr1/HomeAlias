package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/homealias/homealias/backend/internal/config"
	"github.com/homealias/homealias/backend/internal/dns"
	httpapi "github.com/homealias/homealias/backend/internal/http"
	"github.com/homealias/homealias/backend/internal/store"
)

// fakeProvider imita a Cloudflare sem rede.
type fakeProvider struct{}

func (fakeProvider) ValidateCredentials(context.Context, string) ([]dns.Zone, error) {
	return []dns.Zone{{ID: "z1", Name: "exemplo.com"}}, nil
}
func (fakeProvider) UpsertRecord(context.Context, dns.UpsertInput) error { return nil }
func (fakeProvider) DeleteRecord(context.Context, dns.DeleteInput) error { return nil }
func (fakeProvider) GetRecord(context.Context, dns.GetInput) (*dns.Record, error) {
	return nil, nil
}

// session é um cliente HTTP que guarda cookies manualmente (os cookies são Secure e o httptest roda em HTTP).
type session struct {
	t       *testing.T
	srv     *httptest.Server
	cookies map[string]string
}

func newSession(t *testing.T, srv *httptest.Server) *session {
	return &session{t: t, srv: srv, cookies: map[string]string{}}
}

func (s *session) do(method, path string, body any, withCSRF bool) (int, []byte) {
	s.t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, s.srv.URL+path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range s.cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}
	if withCSRF && method != http.MethodGet {
		req.Header.Set("X-CSRF-Token", s.cookies["homealias_csrf"])
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	for _, c := range res.Cookies() {
		if c.MaxAge < 0 || c.Value == "" {
			delete(s.cookies, c.Name)
		} else {
			s.cookies[c.Name] = c.Value
		}
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(res.Body)
	return res.StatusCode, buf.Bytes()
}

func (s *session) call(method, path string, body any, want int) []byte {
	s.t.Helper()
	code, out := s.do(method, path, body, true)
	if code != want {
		s.t.Fatalf("%s %s = %d, esperado %d: %s", method, path, code, want, out)
	}
	return out
}

func testDB(t *testing.T) (string, func() *config.Config) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("defina TEST_DATABASE_DSN (banco dedicado, nome terminando em _test) para rodar os testes de API")
	}
	name := dsn[strings.LastIndex(dsn, "/")+1:]
	if i := strings.Index(name, "?"); i >= 0 {
		name = name[:i]
	}
	if !strings.HasSuffix(name, "_test") {
		t.Fatalf("recusado: o banco %q não termina em _test (o teste apaga todas as tabelas)", name)
	}
	return dsn, func() *config.Config {
		return &config.Config{
			EncryptionKey:   bytes.Repeat([]byte{7}, 32),
			SessionKey:      bytes.Repeat([]byte{9}, 32),
			TrustCFHeaders:  true,
			WarningAfterSec: 900,
			OfflineAfterSec: 3600,
			SessionIdle:     time.Hour,
		}
	}
}

func TestPanelAPIEndToEnd(t *testing.T) {
	dsn, mkCfg := testDB(t)

	m, err := migrate.New("file://../../migrations", "mysql://"+dsn+"&multiStatements=true")
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := m.Drop(); err != nil {
		t.Fatalf("drop: %v", err)
	}
	m.Close()
	m, _ = migrate.New("file://../../migrations", "mysql://"+dsn+"&multiStatements=true")
	if err := m.Up(); err != nil {
		t.Fatalf("up: %v", err)
	}
	m.Close()

	db, err := store.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := store.BootstrapAdmin(context.Background(), db, "admin@test.local", "AdminPass123!"); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(httpapi.NewRouter(mkCfg(), db, fakeProvider{}))
	defer srv.Close()

	admin := newSession(t, srv)

	t.Run("autenticação", func(t *testing.T) {
		admin.t = t
		if code, _ := admin.do(http.MethodGet, "/api/hosts", nil, false); code != http.StatusUnauthorized {
			t.Fatalf("sem sessão deveria ser 401, veio %d", code)
		}
		if code, _ := admin.do(http.MethodPost, "/api/auth/login", map[string]string{"email": "admin@test.local", "password": "errada"}, false); code != http.StatusUnauthorized {
			t.Fatalf("senha errada deveria ser 401, veio %d", code)
		}
		admin.call(http.MethodPost, "/api/auth/login", map[string]string{"email": "admin@test.local", "password": "AdminPass123!"}, http.StatusNoContent)
		var me struct{ Email, Role string }
		if err := json.Unmarshal(admin.call(http.MethodGet, "/api/auth/me", nil, 200), &me); err != nil || me.Role != "admin" {
			t.Fatalf("me inválido: %+v %v", me, err)
		}
		if code, _ := admin.do(http.MethodPost, "/api/connections", map[string]any{}, false); code != http.StatusForbidden {
			t.Fatalf("POST sem CSRF deveria ser 403, veio %d", code)
		}
	})

	t.Run("listagens respondem 200 com banco vazio", func(t *testing.T) {
		admin.t = t
		for _, p := range []string{"/api/users", "/api/connections", "/api/hosts", "/api/history", "/api/alert-channels", "/api/alert-rules", "/api/audit", "/api/settings/status-defaults"} {
			admin.call(http.MethodGet, p, nil, 200)
		}
	})

	var hostID, connectionID string
	t.Run("fluxo conexão, host, token e atualização", func(t *testing.T) {
		admin.t = t
		admin.call(http.MethodPost, "/api/connections", map[string]any{"name": "Conta", "api_token": "cf-token-1234", "test": true}, http.StatusCreated)
		var conns []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Suffix string `json:"api_token_suffix"`
		}
		_ = json.Unmarshal(admin.call(http.MethodGet, "/api/connections", nil, 200), &conns)
		if len(conns) != 1 || conns[0].Status != "valid" || !strings.HasSuffix(conns[0].Suffix, "1234") {
			t.Fatalf("conexão inesperada: %+v", conns)
		}
		connectionID = conns[0].ID

		admin.call(http.MethodPost, "/api/hosts", map[string]any{
			"connection_id": conns[0].ID, "zone_id": "z1", "zone_name": "exemplo.com", "name": "casa", "enable_a": true,
		}, http.StatusCreated)
		var hosts []struct {
			ID     string `json:"id"`
			FQDN   string `json:"fqdn"`
			Status string `json:"status"`
			IPv4   string `json:"last_ipv4"`
		}
		_ = json.Unmarshal(admin.call(http.MethodGet, "/api/hosts", nil, 200), &hosts)
		if len(hosts) != 1 || hosts[0].FQDN != "casa.exemplo.com" || hosts[0].Status != "never_seen" {
			t.Fatalf("host inesperado: %+v", hosts)
		}
		hostID = hosts[0].ID
		if code, _ := admin.do(http.MethodDelete, "/api/connections/"+connectionID, nil, false); code != http.StatusForbidden {
			t.Fatalf("DELETE sem CSRF deveria ser 403, veio %d", code)
		}
		out := admin.call(http.MethodDelete, "/api/connections/"+connectionID, nil, http.StatusConflict)
		if !strings.Contains(string(out), "hosts vinculados") {
			t.Fatal("conexão em uso deve orientar a remoção dos hosts")
		}

		var tok struct {
			Token        string            `json:"token"`
			Instructions map[string]string `json:"instructions"`
		}
		_ = json.Unmarshal(admin.call(http.MethodPost, "/api/tokens", map[string]any{"name": "roteador", "host_ids": []string{hostID}}, http.StatusCreated), &tok)
		if tok.Token == "" || len(tok.Instructions) == 0 {
			t.Fatalf("token vazio: %+v", tok)
		}

		// O cliente chama /update sem sessão; o IP vem da conexão.
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/update?token="+tok.Token+"&hostname=casa.exemplo.com", nil)
		req.Header.Set("CF-Connecting-IP", "203.0.113.42")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 200 {
			t.Fatalf("/update = %d", res.StatusCode)
		}
		_ = json.Unmarshal(admin.call(http.MethodGet, "/api/hosts", nil, 200), &hosts)
		if hosts[0].Status != "online" || hosts[0].IPv4 != "203.0.113.42" {
			t.Fatalf("host deveria estar online com IP após update: %+v", hosts[0])
		}
		var events []map[string]any
		_ = json.Unmarshal(admin.call(http.MethodGet, "/api/history", nil, 200), &events)
		if len(events) == 0 {
			t.Fatal("histórico vazio após update")
		}
		admin.call(http.MethodGet, "/api/audit", nil, 200)
	})

	t.Run("usuário comum: convite, isolamento e desativação", func(t *testing.T) {
		admin.t = t
		var created struct {
			InviteToken string `json:"invite_token"`
		}
		_ = json.Unmarshal(admin.call(http.MethodPost, "/api/users", map[string]any{"email": "ana@test.local", "name": "Ana", "send_invite": true}, http.StatusCreated), &created)
		if created.InviteToken == "" {
			t.Fatal("convite sem token")
		}
		user := newSession(t, srv)
		if code, _ := user.do(http.MethodPost, "/api/auth/invite/accept", map[string]string{"token": "invalido", "password": "x", "name": "x"}, false); code != http.StatusBadRequest {
			t.Fatalf("convite inválido deveria ser 400, veio %d", code)
		}
		user.call(http.MethodPost, "/api/auth/invite/accept", map[string]string{"token": created.InviteToken, "password": "UserPass123!", "name": "Ana"}, http.StatusNoContent)
		user.call(http.MethodPost, "/api/auth/login", map[string]string{"email": "ana@test.local", "password": "UserPass123!"}, http.StatusNoContent)

		var hosts []any
		_ = json.Unmarshal(user.call(http.MethodGet, "/api/hosts", nil, 200), &hosts)
		if len(hosts) != 0 {
			t.Fatalf("usuário comum não deveria ver hosts do admin: %v", hosts)
		}
		user.call(http.MethodGet, "/api/connections", nil, 200)
		user.call(http.MethodDelete, "/api/connections/"+connectionID, nil, http.StatusNotFound)
		user.call(http.MethodGet, "/api/history", nil, 200)
		if code, _ := user.do(http.MethodGet, "/api/users", nil, false); code != http.StatusNotFound {
			t.Fatalf("/api/users para não-admin deveria ser 404, veio %d", code)
		}
		if code, _ := user.do(http.MethodPost, "/api/hosts/"+hostID+"/sync", nil, true); code != http.StatusNotFound {
			t.Fatalf("sync em host alheio deveria ser 404, veio %d", code)
		}

		var users []struct{ ID, Email string }
		_ = json.Unmarshal(admin.call(http.MethodGet, "/api/users", nil, 200), &users)
		for _, u := range users {
			if u.Email == "ana@test.local" {
				admin.call(http.MethodPost, "/api/users/"+u.ID+"/disable", nil, http.StatusNoContent)
			}
		}
		if code, _ := user.do(http.MethodGet, "/api/hosts", nil, false); code != http.StatusUnauthorized {
			t.Fatalf("usuário desativado deveria perder a sessão (401), veio %d", code)
		}
	})

	t.Run("remover host e sair", func(t *testing.T) {
		admin.t = t
		admin.call(http.MethodPost, "/api/hosts/"+hostID+"/sync", nil, 200)
		admin.call(http.MethodDelete, "/api/hosts/"+hostID+"?delete_dns=true", nil, http.StatusNoContent)
		admin.call(http.MethodDelete, "/api/connections/"+connectionID, nil, http.StatusNoContent)
		admin.call(http.MethodDelete, "/api/connections/"+connectionID, nil, http.StatusNotFound)
		var remaining []any
		_ = json.Unmarshal(admin.call(http.MethodGet, "/api/connections", nil, http.StatusOK), &remaining)
		if len(remaining) != 0 {
			t.Fatal("conexão permaneceu após exclusão")
		}
		admin.call(http.MethodPost, "/api/auth/logout", nil, http.StatusNoContent)
		if code, _ := admin.do(http.MethodGet, "/api/auth/me", nil, false); code != http.StatusUnauthorized {
			t.Fatalf("após logout /me deveria ser 401, veio %d", code)
		}
	})
}
