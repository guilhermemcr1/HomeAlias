package panel

import (
	"net/http"
	"time"

	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/jmoiron/sqlx"
)

type AuditAPI struct {
	DB *sqlx.DB
}

func (a *AuditAPI) List(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	if err := auth.AdminOnly(actor); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	type row struct {
		ID           int64     `db:"id" json:"id"`
		CreatedAt    time.Time `db:"created_at" json:"created_at"`
		ActorID      *string   `db:"actor_id" json:"actor_id"`
		ActorRole    string    `db:"actor_role" json:"actor_role"`
		Action       string    `db:"action" json:"action"`
		ResourceType string    `db:"resource_type" json:"resource_type"`
		ResourceID   string    `db:"resource_id" json:"resource_id"`
		IP           string    `db:"ip" json:"ip"`
		Summary      string    `db:"summary" json:"summary"`
	}
	var rows []row
	q := `SELECT id, created_at, actor_id, actor_role, action, resource_type, resource_id, ip, summary FROM audit_log WHERE 1=1`
	args := []any{}
	if actorID := r.URL.Query().Get("actor_id"); actorID != "" {
		q += ` AND actor_id=?`
		args = append(args, actorID)
	}
	if action := r.URL.Query().Get("action"); action != "" {
		q += ` AND action=?`
		args = append(args, action)
	}
	q += ` ORDER BY created_at DESC LIMIT 200`
	if err := a.DB.SelectContext(r.Context(), &rows, q, args...); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
