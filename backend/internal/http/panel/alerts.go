package panel

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/homealias/homealias/backend/internal/alert"
	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/validate"
	"github.com/jmoiron/sqlx"
)

type AlertsAPI struct {
	DB       *sqlx.DB
	Telegram *alert.Telegram
	SMTP     *alert.SMTP
}

func (a *AlertsAPI) ListChannels(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	type row struct {
		ID            string  `db:"id" json:"id"`
		OwnerID       *string `db:"owner_id" json:"owner_id"`
		Scope         string  `db:"scope" json:"scope"`
		Type          string  `db:"type" json:"type"`
		Name          string  `db:"name" json:"name"`
		Destination   string  `db:"destination" json:"destination"`
		LastSendStatus string `db:"last_send_status" json:"last_send_status"`
	}
	var rows []row
	var err error
	if actor.Role == "admin" {
		err = a.DB.SelectContext(r.Context(), &rows, `SELECT id, owner_id, scope, type, name, destination, last_send_status FROM alert_channels`)
	} else {
		err = a.DB.SelectContext(r.Context(), &rows, `SELECT id, owner_id, scope, type, name, destination, last_send_status FROM alert_channels WHERE owner_id=? OR scope='instance'`, actor.ID)
	}
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (a *AlertsAPI) CreateChannel(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	limitBody(w, r)
	var body struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		Destination string `json:"destination"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var verr error
	if body.Name, verr = validate.Name("Nome", body.Name); reject(w, verr) {
		return
	}
	if body.Destination, verr = validate.ChannelDestination(body.Type, body.Destination); reject(w, verr) {
		return
	}
	if body.Scope == "" {
		body.Scope = "user"
	}
	if body.Scope != "user" && body.Scope != "instance" {
		http.Error(w, "Escopo inválido.", http.StatusBadRequest)
		return
	}
	if body.Scope == "instance" && actor.Role != "admin" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	id := uuid.NewString()
	var owner *string
	if body.Scope == "user" {
		owner = &actor.ID
	}
	_, err := a.DB.ExecContext(r.Context(), `
		INSERT INTO alert_channels (id, owner_id, scope, type, name, destination, last_send_status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 'never', ?)`, id, owner, body.Scope, body.Type, body.Name, body.Destination, time.Now().UTC())
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (a *AlertsAPI) TestChannel(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	id := chi.URLParam(r, "id")
	var ch struct {
		OwnerID     *string `db:"owner_id"`
		Type        string  `db:"type"`
		Destination string  `db:"destination"`
		Scope       string  `db:"scope"`
	}
	err := a.DB.GetContext(r.Context(), &ch, `SELECT owner_id, type, destination, scope FROM alert_channels WHERE id=?`, id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if ch.Scope == "user" && ch.OwnerID != nil && auth.RequireOwner(actor, *ch.OwnerID) != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var sendErr error
	msg := "HomeAlias teste de alerta"
	switch ch.Type {
	case "telegram":
		sendErr = a.Telegram.Send(ch.Destination, msg)
	case "email":
		sendErr = a.SMTP.Send(ch.Destination, "Teste do canal de e-mail", "Este é um teste de envio do HomeAlias.\n\nSe esta mensagem chegou até você, o envio para este endereço funcionou.\n\nVocê pode voltar à página Alertas para consultar o resultado do teste.")
	default:
		http.Error(w, "Escolha e-mail ou Telegram para enviar a mensagem.", http.StatusBadRequest)
		return
	}
	status := "ok"
	var errStr *string
	if sendErr != nil {
		status = "error"
		s := sendErr.Error()
		errStr = &s
	}
	_, _ = a.DB.ExecContext(r.Context(), `UPDATE alert_channels SET last_send_status=?, last_send_error=?, last_send_at=? WHERE id=?`, status, errStr, time.Now().UTC(), id)
	writeJSON(w, http.StatusOK, map[string]any{"status": status, "error": errStr})
}

func (a *AlertsAPI) CreateRule(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	var body struct {
		Trigger        string   `json:"trigger"`
		ParamJSON      any      `json:"param"`
		ScopeType      string   `json:"scope_type"`
		ScopeID        *string  `json:"scope_id"`
		ChannelIDs     []string `json:"channel_ids"`
		MinIntervalSec int      `json:"min_interval_sec"`
		ForUserID      *string  `json:"for_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.MinIntervalSec == 0 {
		body.MinIntervalSec = 300
	}
	createdByAdmin := actor.Role == "admin" && body.ForUserID != nil
	ownerID := actor.ID
	if createdByAdmin {
		ownerID = *body.ForUserID
	}
	if body.ScopeType == "instance" && actor.Role != "admin" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	pj, _ := json.Marshal(body.ParamJSON)
	id := uuid.NewString()
	_, err := a.DB.ExecContext(r.Context(), `
		INSERT INTO alert_rules (id, owner_id, created_by_admin, trigger_type, param_json, scope_type, scope_id, min_interval_sec, state, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'normal', 1, ?)`,
		id, ownerID, createdByAdmin, body.Trigger, string(pj), body.ScopeType, body.ScopeID, body.MinIntervalSec, time.Now().UTC())
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	for _, cid := range body.ChannelIDs {
		_, _ = a.DB.ExecContext(r.Context(), `INSERT INTO rule_channels (rule_id, channel_id) VALUES (?, ?)`, id, cid)
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "created_by_admin": createdByAdmin})
}

func (a *AlertsAPI) ListRules(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	type row struct {
		ID              string  `db:"id" json:"id"`
		OwnerID         *string `db:"owner_id" json:"owner_id"`
		CreatedByAdmin  bool    `db:"created_by_admin" json:"created_by_admin"`
		TriggerType     string  `db:"trigger_type" json:"trigger"`
		ScopeType       string  `db:"scope_type" json:"scope_type"`
		State           string  `db:"state" json:"state"`
		Enabled         bool    `db:"enabled" json:"enabled"`
	}
	var rows []row
	var err error
	if actor.Role == "admin" {
		err = a.DB.SelectContext(r.Context(), &rows, `SELECT id, owner_id, created_by_admin, trigger_type, scope_type, state, enabled FROM alert_rules`)
	} else {
		err = a.DB.SelectContext(r.Context(), &rows, `SELECT id, owner_id, created_by_admin, trigger_type, scope_type, state, enabled FROM alert_rules WHERE owner_id=? OR scope_type='instance'`, actor.ID)
	}
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
