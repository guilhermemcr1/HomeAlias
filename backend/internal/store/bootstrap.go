package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/homealias/homealias/backend/internal/validate"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/jmoiron/sqlx"
)

func BootstrapAdmin(ctx context.Context, db *sqlx.DB, email, password string) error {
	var count int
	if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users WHERE role = 'admin'`); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if err := validate.Password(password); err != nil {
		return fmt.Errorf("HOMEALIAS_ADMIN_PASSWORD fraca: %w (use 12+ caracteres, não comum)", err)
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	id := uuid.NewString()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, name, password_hash, role, status, created_at)
		VALUES (?, ?, ?, ?, 'admin', 'active', ?)`,
		id, strings.ToLower(email), "Admin", hash, time.Now().UTC())
	return err
}

type UserRow struct {
	ID           string         `db:"id"`
	Email        string         `db:"email"`
	Name         string         `db:"name"`
	PasswordHash sql.NullString `db:"password_hash"`
	Role         string         `db:"role"`
	Status       string         `db:"status"`
}

func GetUserByEmail(ctx context.Context, db *sqlx.DB, email string) (*UserRow, error) {
	var u UserRow
	err := db.GetContext(ctx, &u, `SELECT id, email, name, password_hash, role, status FROM users WHERE email = ?`, strings.ToLower(email))
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByID(ctx context.Context, db *sqlx.DB, id string) (*UserRow, error) {
	var u UserRow
	err := db.GetContext(ctx, &u, `SELECT id, email, name, password_hash, role, status FROM users WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
