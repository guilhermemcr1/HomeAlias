package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/jmoiron/sqlx"
)

type DBSessionStore struct {
	DB *sqlx.DB
}

func (s *DBSessionStore) Create(userID, ip string, idle time.Duration) (*auth.Session, error) {
	id, err := auth.NewSessionID()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	exp := now.Add(idle)
	_, err = s.DB.Exec(`INSERT INTO sessions (id, user_id, expires_at, created_at, last_seen_at, ip) VALUES (?,?,?,?,?,?)`,
		id, userID, exp, now, now, ip)
	if err != nil {
		return nil, err
	}
	return &auth.Session{ID: id, UserID: userID, ExpiresAt: exp, LastSeen: now, IP: ip}, nil
}

func (s *DBSessionStore) Get(id string) (*auth.Session, error) {
	var row struct {
		ID        string    `db:"id"`
		UserID    string    `db:"user_id"`
		ExpiresAt time.Time `db:"expires_at"`
		LastSeen  time.Time `db:"last_seen_at"`
		IP        string    `db:"ip"`
	}
	err := s.DB.Get(&row, `SELECT id, user_id, expires_at, last_seen_at, ip FROM sessions WHERE id = ?`, id)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	if time.Now().UTC().After(row.ExpiresAt) {
		_ = s.Delete(id)
		return nil, sql.ErrNoRows
	}
	return &auth.Session{ID: row.ID, UserID: row.UserID, ExpiresAt: row.ExpiresAt, LastSeen: row.LastSeen, IP: row.IP}, nil
}

func (s *DBSessionStore) Touch(id string, idle time.Duration) error {
	now := time.Now().UTC()
	_, err := s.DB.Exec(`UPDATE sessions SET last_seen_at = ?, expires_at = ? WHERE id = ?`, now, now.Add(idle), id)
	return err
}

func (s *DBSessionStore) Delete(id string) error {
	_, err := s.DB.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (s *DBSessionStore) DeleteUserSessions(ctx context.Context, userID string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

// DeleteOtherSessions encerra todas as sessões do usuário, exceto a informada.
func (s *DBSessionStore) DeleteOtherSessions(ctx context.Context, userID, keepID string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ? AND id <> ?`, userID, keepID)
	return err
}
