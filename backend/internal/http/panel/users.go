package panel

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/homealias/homealias/backend/internal/audit"
	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/store"
	"github.com/homealias/homealias/backend/internal/validate"
	"github.com/jmoiron/sqlx"
)

type UsersAPI struct {
	DB     *sqlx.DB
	Audit  *audit.Writer
	Sess   *store.DBSessionStore
}

func (u *UsersAPI) List(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	if err := auth.AdminOnly(actor); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var rows []store.UserRow
	if err := u.DB.SelectContext(r.Context(), &rows, `SELECT id, email, name, password_hash, role, status FROM users ORDER BY email`); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{"id": row.ID, "email": row.Email, "name": row.Name, "role": row.Role, "status": row.Status})
	}
	writeJSON(w, http.StatusOK, out)
}

func (u *UsersAPI) Create(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	limitBody(w, r)
	if err := auth.AdminOnly(actor); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var body struct {
		Email           string `json:"email"`
		Name            string `json:"name"`
		InitialPassword string `json:"initial_password"`
		SendInvite      bool   `json:"send_invite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	email, verr := validate.Email(body.Email)
	if reject(w, verr) {
		return
	}
	if body.Name, verr = validate.Name("Nome", body.Name); reject(w, verr) {
		return
	}
	if !body.SendInvite && body.InitialPassword != "" {
		if reject(w, validate.Password(body.InitialPassword)) {
			return
		}
	}
	id := uuid.NewString()
	if body.SendInvite || body.InitialPassword == "" {
		raw, hash, _, err := auth.GenerateDDNSToken()
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		_, err = u.DB.ExecContext(r.Context(), `
			INSERT INTO users (id, email, name, password_hash, role, status, created_at)
			VALUES (?, ?, ?, NULL, 'user', 'invite_pending', ?)`, id, email, body.Name, time.Now().UTC())
		if err != nil {
			http.Error(w, "Este e-mail já está em uso.", http.StatusConflict)
			return
		}
		_, _ = u.DB.ExecContext(r.Context(), `
			INSERT INTO invites (id, email, token_hash, expires_at, created_by, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`, uuid.NewString(), email, hash, time.Now().UTC().Add(72*time.Hour), actor.ID, time.Now().UTC())
		_ = u.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "user_invite", ResourceType: "user", ResourceID: id, IP: r.RemoteAddr, Summary: "invite created"})
		writeJSON(w, http.StatusCreated, map[string]any{"id": id, "email": email, "status": "invite_pending", "invite_token": raw})
		return
	}
	hash, err := auth.HashPassword(body.InitialPassword)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_, err = u.DB.ExecContext(r.Context(), `
		INSERT INTO users (id, email, name, password_hash, role, status, created_at)
		VALUES (?, ?, ?, ?, 'user', 'active', ?)`, id, email, body.Name, hash, time.Now().UTC())
	if err != nil {
		http.Error(w, "Este e-mail já está em uso.", http.StatusConflict)
		return
	}
	_ = u.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "user_create", ResourceType: "user", ResourceID: id, IP: r.RemoteAddr, Summary: "user created with initial password"})
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "email": email, "status": "active"})
}

func (u *UsersAPI) Disable(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	if err := auth.AdminOnly(actor); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	id := chi.URLParam(r, "id")
	_, err := u.DB.ExecContext(r.Context(), `UPDATE users SET status='disabled' WHERE id=? AND role='user'`, id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_ = u.Sess.DeleteUserSessions(r.Context(), id)
	_ = u.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "user_disable", ResourceType: "user", ResourceID: id, IP: r.RemoteAddr, Summary: "user disabled"})
	w.WriteHeader(http.StatusNoContent)
}

func (u *UsersAPI) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var verr error
	if len(body.Token) == 0 || len(body.Token) > 256 {
		http.Error(w, "Código do convite inválido.", http.StatusBadRequest)
		return
	}
	if body.Name != "" {
		if body.Name, verr = validate.Name("Nome", body.Name); reject(w, verr) {
			return
		}
	}
	if reject(w, validate.Password(body.Password)) {
		return
	}
	hash := auth.HashDDNSToken(strings.TrimSpace(body.Token))
	var invite struct {
		ID    string    `db:"id"`
		Email string    `db:"email"`
		Exp   time.Time `db:"expires_at"`
	}
	err := u.DB.GetContext(r.Context(), &invite, `SELECT id, email, expires_at FROM invites WHERE token_hash=? AND consumed_at IS NULL`, hash)
	if err != nil || time.Now().UTC().After(invite.Exp) {
		http.Error(w, "Este convite expirou ou já foi usado. Peça um novo link ao administrador.", http.StatusBadRequest)
		return
	}
	ph, err := auth.HashPassword(body.Password)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	name := body.Name
	if name == "" {
		name = invite.Email
	}
	_, err = u.DB.ExecContext(r.Context(), `UPDATE users SET password_hash=?, name=?, status='active' WHERE email=? AND status='invite_pending'`, ph, name, invite.Email)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_, _ = u.DB.ExecContext(r.Context(), `UPDATE invites SET consumed_at=? WHERE id=?`, time.Now().UTC(), invite.ID)
	w.WriteHeader(http.StatusNoContent)
}

// Update altera nome, e-mail e perfil (role) de outro usuário. O admin edita a própria conta em /conta.
func (u *UsersAPI) Update(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	if err := auth.AdminOnly(actor); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	limitBody(w, r)
	id := chi.URLParam(r, "id")
	var body struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	target, err := store.GetUserByID(r.Context(), u.DB, id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
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
	if body.Role != "admin" && body.Role != "user" {
		http.Error(w, "Perfil inválido.", http.StatusBadRequest)
		return
	}
	if id == actor.ID && body.Role != target.Role {
		http.Error(w, "Você não pode alterar o próprio perfil de acesso.", http.StatusBadRequest)
		return
	}
	if _, err := u.DB.ExecContext(r.Context(), `UPDATE users SET name=?, email=?, role=? WHERE id=?`, name, email, body.Role, id); err != nil {
		if isDuplicate(err) {
			http.Error(w, "Este e-mail já está em uso.", http.StatusConflict)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if email != target.Email || body.Role != target.Role {
		_ = u.Sess.DeleteUserSessions(r.Context(), id)
	}
	_ = u.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "user_update", ResourceType: "user", ResourceID: id, IP: r.RemoteAddr, Summary: "user updated"})
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "email": email, "name": name, "role": body.Role})
}

// ResetPassword define uma nova senha para outro usuário e encerra as sessões dele.
func (u *UsersAPI) ResetPassword(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	if err := auth.AdminOnly(actor); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	limitBody(w, r)
	id := chi.URLParam(r, "id")
	var body struct {
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if id == actor.ID {
		http.Error(w, "Para trocar a sua senha use Minha conta.", http.StatusBadRequest)
		return
	}
	if reject(w, validate.Password(body.NewPassword)) {
		return
	}
	hash, err := auth.HashPassword(body.NewPassword)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	res, err := u.DB.ExecContext(r.Context(), `UPDATE users SET password_hash=?, status=IF(status='invite_pending','active',status) WHERE id=?`, hash, id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	_ = u.Sess.DeleteUserSessions(r.Context(), id)
	_ = u.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "user_password_reset", ResourceType: "user", ResourceID: id, IP: r.RemoteAddr, Summary: "password reset by admin"})
	w.WriteHeader(http.StatusNoContent)
}

// Enable reativa um usuário desativado.
func (u *UsersAPI) Enable(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.ActorFrom(r.Context())
	if err := auth.AdminOnly(actor); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	id := chi.URLParam(r, "id")
	_, err := u.DB.ExecContext(r.Context(), `UPDATE users SET status='active' WHERE id=? AND status='disabled' AND password_hash IS NOT NULL`, id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_ = u.Audit.Write(r.Context(), audit.Entry{ActorID: &actor.ID, ActorRole: actor.Role, Action: "user_enable", ResourceType: "user", ResourceID: id, IP: r.RemoteAddr, Summary: "user enabled"})
	w.WriteHeader(http.StatusNoContent)
}
