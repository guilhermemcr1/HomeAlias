package panel

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/homealias/homealias/backend/internal/audit"
	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/store"
	"github.com/homealias/homealias/backend/internal/validate"
	"github.com/jmoiron/sqlx"
)

// isDuplicate identifica violação da chave única de e-mail no MariaDB.
func isDuplicate(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate")
}

// AccountAPI cuida do próprio perfil do usuário logado.
type AccountAPI struct {
	DB       *sqlx.DB
	Sessions *store.DBSessionStore
	Audit    *audit.Writer
	RL       *auth.RateLimiter
}

func (a *AccountAPI) Me(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	u, err := store.GetUserByID(r.Context(), a.DB, actor.ID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": u.ID, "email": u.Email, "name": u.Name, "role": u.Role})
}

// checkPassword confere a senha atual com limite de tentativas por usuário.
func (a *AccountAPI) checkPassword(w http.ResponseWriter, r *http.Request, u *store.UserRow, current string) bool {
	if ok, _ := a.RL.Allow("pw:" + u.ID); !ok {
		http.Error(w, "Muitas tentativas. Aguarde um minuto e tente de novo.", http.StatusTooManyRequests)
		return false
	}
	if !u.PasswordHash.Valid {
		http.Error(w, "A senha atual está incorreta.", http.StatusBadRequest)
		return false
	}
	ok, err := auth.VerifyPassword(u.PasswordHash.String, current)
	if err != nil || !ok {
		http.Error(w, "A senha atual está incorreta.", http.StatusBadRequest)
		return false
	}
	return true
}

// UpdateProfile altera nome e/ou e-mail. Trocar o e-mail exige a senha atual.
func (a *AccountAPI) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	limitBody(w, r)
	var body struct {
		Name            string `json:"name"`
		Email           string `json:"email"`
		CurrentPassword string `json:"current_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	u, err := store.GetUserByID(r.Context(), a.DB, actor.ID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name, verr := validate.Name("Nome", body.Name)
	if reject(w, verr) {
		return
	}
	email, verr := validate.Email(body.Email)
	if reject(w, verr) {
		return
	}
	if email != u.Email && !a.checkPassword(w, r, u, body.CurrentPassword) {
		return
	}
	if _, err := a.DB.ExecContext(r.Context(), `UPDATE users SET name=?, email=? WHERE id=?`, name, email, u.ID); err != nil {
		if isDuplicate(err) {
			http.Error(w, "Este e-mail já está em uso.", http.StatusConflict)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_ = a.Audit.Write(r.Context(), audit.Entry{ActorID: &u.ID, ActorRole: u.Role, Action: "profile_update", ResourceType: "user", ResourceID: u.ID, ResourceOwnerID: &u.ID, IP: r.RemoteAddr, Summary: "profile updated"})
	writeJSON(w, http.StatusOK, map[string]any{"id": u.ID, "email": email, "name": name, "role": u.Role})
}

// ChangePassword troca a senha e encerra as outras sessões do usuário.
func (a *AccountAPI) ChangePassword(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	limitBody(w, r)
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	u, err := store.GetUserByID(r.Context(), a.DB, actor.ID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if reject(w, validate.Password(body.NewPassword)) {
		return
	}
	if body.NewPassword == body.CurrentPassword {
		http.Error(w, "A nova senha deve ser diferente da atual.", http.StatusBadRequest)
		return
	}
	if !a.checkPassword(w, r, u, body.CurrentPassword) {
		return
	}
	hash, err := auth.HashPassword(body.NewPassword)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if _, err := a.DB.ExecContext(r.Context(), `UPDATE users SET password_hash=? WHERE id=?`, hash, u.ID); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if sid, err := auth.ReadSessionCookie(r); err == nil {
		_ = a.Sessions.DeleteOtherSessions(r.Context(), u.ID, sid)
	}
	_ = a.Audit.Write(r.Context(), audit.Entry{ActorID: &u.ID, ActorRole: u.Role, Action: "password_change", ResourceType: "user", ResourceID: u.ID, ResourceOwnerID: &u.ID, IP: r.RemoteAddr, Summary: "password changed"})
	w.WriteHeader(http.StatusNoContent)
}
