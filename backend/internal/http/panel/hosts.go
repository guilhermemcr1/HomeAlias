package panel

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
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

type HostsAPI struct {
	DB             *sqlx.DB
	Provider       dns.Provider
	EncKey         []byte
	Audit          *audit.Writer
	WarningAfter   int
	OfflineAfter   int
}

func (h *HostsAPI) List(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	var rows []domain.Host
	var err error
	if actor.Role == "admin" {
		err = h.DB.SelectContext(r.Context(), &rows, `SELECT * FROM hosts ORDER BY fqdn`)
	} else {
		err = h.DB.SelectContext(r.Context(), &rows, `SELECT * FROM hosts WHERE owner_id=? ORDER BY fqdn`, actor.ID)
	}
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	now := time.Now().UTC()
	out := make([]map[string]any, 0, len(rows))
	for _, host := range rows {
		wSec, oSec := domain.HostLimits(host, h.WarningAfter, h.OfflineAfter)
		status := domain.DeriveStatus(domain.LatestSeen(host), wSec, oSec, now)
		m := map[string]any{
			"id": host.ID, "owner_id": host.OwnerID, "connection_id": host.ConnectionID,
			"fqdn": host.FQDN, "zone_name": host.ZoneName, "enable_a": host.EnableA, "enable_aaaa": host.EnableAAAA,
			"proxied": host.Proxied, "ttl": host.TTL, "last_ipv4": host.LastIPv4, "last_ipv6": host.LastIPv6,
			"last_seen_v4": host.LastSeenV4, "last_seen_v6": host.LastSeenV6, "status": status,
			"warning_after_sec": host.WarningAfterSec, "offline_after_sec": host.OfflineAfterSec,
		}
		out = append(out, m)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *HostsAPI) Create(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	limitBody(w, r)
	var body struct {
		ConnectionID  string `json:"connection_id"`
		ZoneID        string `json:"zone_id"`
		ZoneName      string `json:"zone_name"`
		Name          string `json:"name"`
		EnableA       bool   `json:"enable_a"`
		EnableAAAA    bool   `json:"enable_aaaa"`
		Proxied       bool   `json:"proxied"`
		TTL           int    `json:"ttl"`
		AdoptExisting bool   `json:"adopt_existing"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var verr error
	if verr = validate.ID("Conexão", body.ConnectionID); reject(w, verr) {
		return
	}
	if body.ZoneID == "" || len(body.ZoneID) > 64 {
		http.Error(w, "Zona inválida.", http.StatusBadRequest)
		return
	}
	if body.ZoneName, verr = validate.ZoneName(body.ZoneName); reject(w, verr) {
		return
	}
	if body.Name, verr = validate.HostName(body.Name); reject(w, verr) {
		return
	}
	if body.TTL, verr = validate.TTL(body.TTL); reject(w, verr) {
		return
	}
	if !body.EnableA && !body.EnableAAAA {
		http.Error(w, "Ative ao menos um tipo de registro (A ou AAAA).", http.StatusBadRequest)
		return
	}
	var conn domain.Connection
	err := h.DB.GetContext(r.Context(), &conn, `SELECT * FROM connections WHERE id=?`, body.ConnectionID)
	if err != nil || auth.RequireOwner(actor, conn.OwnerID) != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	fqdn := domain.FQDN(body.ZoneName, body.Name)
	if reject(w, validate.FQDNLen(fqdn)) {
		return
	}
	var existing string
	err = h.DB.GetContext(r.Context(), &existing, `SELECT id FROM hosts WHERE fqdn=?`, fqdn)
	if err == nil {
		http.Error(w, "fqdn taken", http.StatusConflict)
		return
	}
	plain, err := auth.Decrypt(h.EncKey, conn.APITokenCipher)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	pctx := cf.WithToken(r.Context(), plain)
	found, err := h.existingRecords(pctx, body.ZoneID, fqdn, body.EnableA, body.EnableAAAA)
	if err != nil {
		http.Error(w, "Não foi possível consultar a Cloudflare para verificar registros existentes.", http.StatusBadGateway)
		return
	}
	if len(found) > 0 && !body.AdoptExisting {
		writeJSON(w, http.StatusConflict, map[string]any{"code": "record_exists", "fqdn": fqdn, "records": found})
		return
	}
	id := uuid.NewString()
	_, err = h.DB.ExecContext(r.Context(), `
		INSERT INTO hosts (id, owner_id, connection_id, zone_id, zone_name, name, fqdn, enable_a, enable_aaaa, proxied, ttl, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, actor.ID, body.ConnectionID, body.ZoneID, body.ZoneName, body.Name, fqdn, body.EnableA, body.EnableAAAA, body.Proxied, body.TTL, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			http.Error(w, "fqdn taken", http.StatusConflict)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_ = h.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "host_create", ResourceType: "host", ResourceID: id, ResourceOwnerID: &actor.ID, IP: r.RemoteAddr, Summary: "host " + fqdn})
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "fqdn": fqdn})
}

type existingRecord struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

// existingRecords lista os registros A/AAAA que já existem na zona para o nome informado.
func (h *HostsAPI) existingRecords(ctx context.Context, zoneID, fqdn string, a, aaaa bool) ([]existingRecord, error) {
	out := []existingRecord{}
	for _, t := range []struct {
		typ string
		on  bool
	}{{"A", a}, {"AAAA", aaaa}} {
		if !t.on {
			continue
		}
		rec, err := h.Provider.GetRecord(ctx, dns.GetInput{ZoneID: zoneID, Type: t.typ, Name: fqdn})
		if err != nil {
			return nil, err
		}
		if rec != nil {
			out = append(out, existingRecord{Type: t.typ, Content: rec.Content, TTL: rec.TTL, Proxied: rec.Proxied})
		}
	}
	return out, nil
}

// Check diz, antes de criar, se o nome já existe no app ou na Cloudflare (A/AAAA), para o painel perguntar o que fazer.
func (h *HostsAPI) Check(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	limitBody(w, r)
	var body struct {
		ConnectionID string `json:"connection_id"`
		ZoneID       string `json:"zone_id"`
		ZoneName     string `json:"zone_name"`
		Name         string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var verr error
	if body.ZoneID == "" || len(body.ZoneID) > 64 {
		http.Error(w, "Zona inválida.", http.StatusBadRequest)
		return
	}
	if body.ZoneName, verr = validate.ZoneName(body.ZoneName); reject(w, verr) {
		return
	}
	if body.Name, verr = validate.HostName(body.Name); reject(w, verr) {
		return
	}
	var conn domain.Connection
	if err := h.DB.GetContext(r.Context(), &conn, `SELECT * FROM connections WHERE id=?`, body.ConnectionID); err != nil || auth.RequireOwner(actor, conn.OwnerID) != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	fqdn := domain.FQDN(body.ZoneName, body.Name)
	var id string
	inApp := h.DB.GetContext(r.Context(), &id, `SELECT id FROM hosts WHERE fqdn=?`, fqdn) == nil
	plain, err := auth.Decrypt(h.EncKey, conn.APITokenCipher)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	found, err := h.existingRecords(cf.WithToken(r.Context(), plain), body.ZoneID, fqdn, true, true)
	if err != nil {
		http.Error(w, "Não foi possível consultar a Cloudflare.", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"fqdn": fqdn, "in_app": inApp, "records": found})
}

func (h *HostsAPI) Patch(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	id := chi.URLParam(r, "id")
	var host domain.Host
	err := h.DB.GetContext(r.Context(), &host, `SELECT * FROM hosts WHERE id=?`, id)
	if err != nil || auth.RequireOwner(actor, host.OwnerID) != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var body struct {
		EnableA         *bool `json:"enable_a"`
		EnableAAAA      *bool `json:"enable_aaaa"`
		Proxied         *bool `json:"proxied"`
		TTL             *int  `json:"ttl"`
		WarningAfterSec *int  `json:"warning_after_sec"`
		OfflineAfterSec *int  `json:"offline_after_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.EnableA != nil {
		host.EnableA = *body.EnableA
	}
	if body.EnableAAAA != nil {
		host.EnableAAAA = *body.EnableAAAA
	}
	if body.Proxied != nil {
		host.Proxied = *body.Proxied
	}
	if body.TTL != nil {
		ttl, verr := validate.TTL(*body.TTL)
		if reject(w, verr) {
			return
		}
		host.TTL = ttl
	}
	host.WarningAfterSec = body.WarningAfterSec
	host.OfflineAfterSec = body.OfflineAfterSec
	_, err = h.DB.ExecContext(r.Context(), `
		UPDATE hosts SET enable_a=?, enable_aaaa=?, proxied=?, ttl=?, warning_after_sec=?, offline_after_sec=? WHERE id=?`,
		host.EnableA, host.EnableAAAA, host.Proxied, host.TTL, host.WarningAfterSec, host.OfflineAfterSec, id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if body.Proxied != nil || body.TTL != nil {
		_ = h.applyDNS(r, host)
	}
	_ = h.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "host_update", ResourceType: "host", ResourceID: id, ResourceOwnerID: &host.OwnerID, IP: r.RemoteAddr, Summary: "host updated"})
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

func (h *HostsAPI) applyDNS(r *http.Request, host domain.Host) error {
	var conn domain.Connection
	if err := h.DB.GetContext(r.Context(), &conn, `SELECT * FROM connections WHERE id=?`, host.ConnectionID); err != nil {
		return err
	}
	plain, err := auth.Decrypt(h.EncKey, conn.APITokenCipher)
	if err != nil {
		return err
	}
	pctx := cf.WithToken(r.Context(), plain)
	if host.EnableA && host.LastIPv4 != nil {
		_ = h.Provider.UpsertRecord(pctx, dns.UpsertInput{ZoneID: host.ZoneID, Type: "A", Name: host.FQDN, Content: *host.LastIPv4, TTL: host.TTL, Proxied: host.Proxied})
	}
	if host.EnableAAAA && host.LastIPv6 != nil {
		_ = h.Provider.UpsertRecord(pctx, dns.UpsertInput{ZoneID: host.ZoneID, Type: "AAAA", Name: host.FQDN, Content: *host.LastIPv6, TTL: host.TTL, Proxied: host.Proxied})
	}
	return nil
}

func (h *HostsAPI) Delete(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	id := chi.URLParam(r, "id")
	deleteDNS := r.URL.Query().Get("delete_dns") == "true"
	var host domain.Host
	err := h.DB.GetContext(r.Context(), &host, `SELECT * FROM hosts WHERE id=?`, id)
	if err != nil || auth.RequireOwner(actor, host.OwnerID) != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if deleteDNS {
		var conn domain.Connection
		_ = h.DB.GetContext(r.Context(), &conn, `SELECT * FROM connections WHERE id=?`, host.ConnectionID)
		if plain, err := auth.Decrypt(h.EncKey, conn.APITokenCipher); err == nil {
			pctx := cf.WithToken(r.Context(), plain)
			_ = h.Provider.DeleteRecord(pctx, dns.DeleteInput{ZoneID: host.ZoneID, Type: "A", Name: host.FQDN})
			_ = h.Provider.DeleteRecord(pctx, dns.DeleteInput{ZoneID: host.ZoneID, Type: "AAAA", Name: host.FQDN})
		}
	}
	_, _ = h.DB.ExecContext(r.Context(), `DELETE FROM token_hosts WHERE host_id=?`, id)
	_, _ = h.DB.ExecContext(r.Context(), `DELETE FROM hosts WHERE id=?`, id)
	_ = h.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "host_delete", ResourceType: "host", ResourceID: id, ResourceOwnerID: &host.OwnerID, IP: r.RemoteAddr, Summary: "host deleted"})
	w.WriteHeader(http.StatusNoContent)
}

func (h *HostsAPI) Sync(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	id := chi.URLParam(r, "id")
	var host domain.Host
	err := h.DB.GetContext(r.Context(), &host, `SELECT * FROM hosts WHERE id=?`, id)
	if err != nil || auth.RequireOwner(actor, host.OwnerID) != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := h.applyDNS(r, host); err != nil {
		http.Error(w, "sync failed", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
