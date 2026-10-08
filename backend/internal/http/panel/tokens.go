package panel

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/homealias/homealias/backend/internal/audit"
	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/domain"
	"github.com/homealias/homealias/backend/internal/validate"
	"github.com/jmoiron/sqlx"
)

type TokensAPI struct {
	DB    *sqlx.DB
	Audit *audit.Writer
}

func (t *TokensAPI) Create(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	var body struct {
		Name      string   `json:"name"`
		HostIDs   []string `json:"host_ids"`
		ExpiresAt *string  `json:"expires_at"`
	}
	limitBody(w, r)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.HostIDs) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var verr error
	if body.Name, verr = validate.Name("Nome do token", body.Name); reject(w, verr) {
		return
	}
	if len(body.HostIDs) > 50 {
		http.Error(w, "Um token pode cobrir no máximo 50 hosts.", http.StatusBadRequest)
		return
	}
	for _, hid := range body.HostIDs {
		if reject(w, validate.ID("Host", hid)) {
			return
		}
	}
	for _, hid := range body.HostIDs {
		var owner string
		err := t.DB.GetContext(r.Context(), &owner, `SELECT owner_id FROM hosts WHERE id=?`, hid)
		if err != nil || auth.RequireOwner(actor, owner) != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}
	raw, hash, prefix, err := auth.GenerateDDNSToken()
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	id := uuid.NewString()
	var exp *time.Time
	if body.ExpiresAt != nil && *body.ExpiresAt != "" {
		tm, err := time.Parse(time.RFC3339, *body.ExpiresAt)
		if err != nil || !tm.After(time.Now()) {
			http.Error(w, "A data de expiração deve ser futura (formato RFC 3339).", http.StatusBadRequest)
			return
		}
		exp = &tm
	}
	_, err = t.DB.ExecContext(r.Context(), `
		INSERT INTO ddns_tokens (id, owner_id, name, token_hash, token_prefix, expires_at, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 'active', ?)`, id, actor.ID, body.Name, hash, prefix, exp, time.Now().UTC())
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	for _, hid := range body.HostIDs {
		_, _ = t.DB.ExecContext(r.Context(), `INSERT INTO token_hosts (token_id, host_id) VALUES (?, ?)`, id, hid)
	}
	fqdns := make([]string, 0, len(body.HostIDs))
	for _, hid := range body.HostIDs {
		var fqdn string
		if err := t.DB.GetContext(r.Context(), &fqdn, `SELECT fqdn FROM hosts WHERE id=?`, hid); err == nil {
			fqdns = append(fqdns, fqdn)
		}
	}
	_ = t.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "token_create", ResourceType: "ddns_token", ResourceID: id, ResourceOwnerID: &actor.ID, IP: r.RemoteAddr, Summary: "token emitted"})
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": id, "token": raw, "token_prefix": prefix, "name": body.Name,
		"instructions": clientInstructions(raw, fqdns),
	})
}

func (t *TokensAPI) Revoke(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	id := chi.URLParam(r, "id")
	var tok domain.DDNSToken
	err := t.DB.GetContext(r.Context(), &tok, `SELECT * FROM ddns_tokens WHERE id=?`, id)
	if err != nil || auth.RequireOwner(actor, tok.OwnerID) != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	_, _ = t.DB.ExecContext(r.Context(), `UPDATE ddns_tokens SET status='revoked' WHERE id=?`, id)
	_ = t.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "token_revoke", ResourceType: "ddns_token", ResourceID: id, ResourceOwnerID: &tok.OwnerID, IP: r.RemoteAddr, Summary: "token revoked"})
	w.WriteHeader(http.StatusNoContent)
}

func clientInstructions(token string, fqdns []string) map[string]string {
	host := "casa.exemplo.com"
	if len(fqdns) > 0 {
		host = fqdns[0]
	}
	return map[string]string{
		"curl": "curl -fsS -H \"Authorization: Bearer " + token + "\" \"https://<HOMEALIAS_HOST>/update?hostname=" + host + "\"",
		"shell": "curl -fsS https://<HOMEALIAS_HOST>/client/update.sh -o homealias-update.sh && chmod +x homealias-update.sh && " +
			"HOMEALIAS_URL=https://<HOMEALIAS_HOST> HOMEALIAS_TOKEN=" + token + " HOMEALIAS_HOSTNAME=" + host + " ./homealias-update.sh",
		"docker": "docker run -d --name homealias-ddns --restart unless-stopped -e HA_TOKEN=" + token + " curlimages/curl:latest " +
			"sh -c 'while true; do curl -fsS -A homealias-docker/1.0 -H \"Authorization: Bearer $HA_TOKEN\" \"https://<HOMEALIAS_HOST>/update?hostname=" + host + "\"; echo; sleep 300; done'",
		"dyndns":  "Servidor: <HOMEALIAS_HOST> | Porta: 443 (HTTPS) | Caminho: /nic/update | Hostname: " + host + " | Usuario: homealias | Senha: " + token,
		"windows": "Baixe https://<HOMEALIAS_HOST>/client/update.ps1 e rode: powershell -ExecutionPolicy Bypass -File .\\homealias-update.ps1 -Url https://<HOMEALIAS_HOST> -Token " + token + " -Hostname " + host + " -Install",
	}
}
