package panel

import (
	"net/http"

	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type HistoryAPI struct {
	DB *sqlx.DB
}

func (h *HistoryAPI) List(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	q := r.URL.Query()
	hostID := q.Get("host_id")
	result := q.Get("result")
	var rows []domain.UpdateEvent
	var err error
	if actor.Role == "admin" {
		query := `SELECT * FROM update_events WHERE 1=1`
		args := []any{}
		if hostID != "" {
			query += ` AND host_id=?`
			args = append(args, hostID)
		}
		if result != "" {
			query += ` AND result=?`
			args = append(args, result)
		}
		query += ` ORDER BY created_at DESC LIMIT 200`
		err = h.DB.SelectContext(r.Context(), &rows, query, args...)
	} else {
		query := `SELECT * FROM update_events WHERE owner_id=?`
		args := []any{actor.ID}
		if hostID != "" {
			query += ` AND host_id=?`
			args = append(args, hostID)
		}
		if result != "" {
			query += ` AND result=?`
			args = append(args, result)
		}
		query += ` ORDER BY created_at DESC LIMIT 200`
		err = h.DB.SelectContext(r.Context(), &rows, query, args...)
	}
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
