package panel

import (
	"encoding/json"
	"net/http"

	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/jmoiron/sqlx"
)

type SettingsAPI struct {
	DB *sqlx.DB
}

func (s *SettingsAPI) GetStatusDefaults(w http.ResponseWriter, r *http.Request) {
	var warning, offline string
	_ = s.DB.GetContext(r.Context(), &warning, `SELECT setting_value FROM instance_settings WHERE setting_key='status_warning_after_sec'`)
	_ = s.DB.GetContext(r.Context(), &offline, `SELECT setting_value FROM instance_settings WHERE setting_key='status_offline_after_sec'`)
	writeJSON(w, http.StatusOK, map[string]string{
		"warning_after_sec": warning,
		"offline_after_sec": offline,
	})
}

func (s *SettingsAPI) PutStatusDefaults(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	if err := auth.AdminOnly(actor); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var body struct {
		WarningAfterSec int `json:"warning_after_sec"`
		OfflineAfterSec int `json:"offline_after_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	_, _ = s.DB.ExecContext(r.Context(), `UPDATE instance_settings SET setting_value=? WHERE setting_key='status_warning_after_sec'`, body.WarningAfterSec)
	_, _ = s.DB.ExecContext(r.Context(), `UPDATE instance_settings SET setting_value=? WHERE setting_key='status_offline_after_sec'`, body.OfflineAfterSec)
	writeJSON(w, http.StatusOK, body)
}
