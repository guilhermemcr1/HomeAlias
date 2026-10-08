package panel

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/homealias/homealias/backend/internal/audit"
	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/dns"
	cf "github.com/homealias/homealias/backend/internal/dns/cloudflare"
	"github.com/homealias/homealias/backend/internal/domain"
	"github.com/homealias/homealias/backend/internal/validate"
	"github.com/jmoiron/sqlx"
)

type ConnectionsAPI struct {
	DB       *sqlx.DB
	Provider dns.Provider
	EncKey   []byte
	Audit    *audit.Writer
}

func (c *ConnectionsAPI) List(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	var rows []domain.Connection
	var err error
	if actor.Role == "admin" {
		err = c.DB.SelectContext(r.Context(), &rows, `SELECT id, owner_id, name, provider, api_token_ciphertext, api_token_suffix, account_hint, zones_json, status, last_checked_at, last_error FROM connections`)
	} else {
		err = c.DB.SelectContext(r.Context(), &rows, `SELECT id, owner_id, name, provider, api_token_ciphertext, api_token_suffix, account_hint, zones_json, status, last_checked_at, last_error FROM connections WHERE owner_id=?`, actor.ID)
	}
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (c *ConnectionsAPI) Create(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	limitBody(w, r)
	var body struct {
		Name        string `json:"name"`
		APIToken    string `json:"api_token"`
		AccountHint string `json:"account_hint"`
		Test        *bool  `json:"test"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var verr error
	if body.Name, verr = validate.Name("Nome", body.Name); reject(w, verr) {
		return
	}
	if body.APIToken, verr = validate.APIToken(body.APIToken); reject(w, verr) {
		return
	}
	if body.AccountHint != "" {
		if body.AccountHint, verr = validate.Name("Dica da conta", body.AccountHint); reject(w, verr) {
			return
		}
	}
	doTest := body.Test == nil || *body.Test
	var zones []dns.Zone
	status := "untested"
	var lastErr *string
	now := time.Now().UTC()
	if doTest {
		z, err := c.Provider.ValidateCredentials(r.Context(), body.APIToken)
		if err != nil {
			http.Error(w, cloudflareMessage(err), http.StatusBadRequest)
			return
		}
		zones = z
		status = "valid"
	}
	cipher, err := auth.Encrypt(c.EncKey, body.APIToken)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	id := uuid.NewString()
	zonesJSON, _ := json.Marshal(zones)
	var hint *string
	if body.AccountHint != "" {
		hint = &body.AccountHint
	}
	zs := string(zonesJSON)
	_, err = c.DB.ExecContext(r.Context(), `
		INSERT INTO connections (id, owner_id, name, provider, api_token_ciphertext, api_token_suffix, account_hint, zones_json, status, last_checked_at, last_error, created_at, updated_at)
		VALUES (?, ?, ?, 'cloudflare', ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, actor.ID, body.Name, cipher, auth.MaskSuffix(body.APIToken, 4), hint, zs, status, now, lastErr, now, now)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_ = c.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "connection_create", ResourceType: "connection", ResourceID: id, ResourceOwnerID: &actor.ID, IP: r.RemoteAddr, Summary: "connection created"})
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "status": status, "api_token_suffix": auth.MaskSuffix(body.APIToken, 4), "zones": zones})
}

func (c *ConnectionsAPI) Test(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	id := chi.URLParam(r, "id")
	var conn domain.Connection
	err := c.DB.GetContext(r.Context(), &conn, `SELECT * FROM connections WHERE id=?`, id)
	if err != nil || auth.RequireOwner(actor, conn.OwnerID) != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	plain, err := auth.Decrypt(c.EncKey, conn.APITokenCipher)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	zones, err := c.Provider.ValidateCredentials(cf.WithToken(r.Context(), plain), plain)
	now := time.Now().UTC()
	if err != nil {
		msg := cloudflareMessage(err)
		_, _ = c.DB.ExecContext(r.Context(), `UPDATE connections SET status='invalid', last_error=?, last_checked_at=? WHERE id=?`, msg, now, id)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	zj, _ := json.Marshal(zones)
	_, _ = c.DB.ExecContext(r.Context(), `UPDATE connections SET status='valid', zones_json=?, last_error=NULL, last_checked_at=? WHERE id=?`, string(zj), now, id)
	writeJSON(w, http.StatusOK, map[string]any{"status": "valid", "zones": zones})
}
