package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func RunRetention(ctx context.Context, db *sqlx.DB, days int) error {
	if days <= 0 {
		days = 90
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	if _, err := db.ExecContext(ctx, `DELETE FROM update_events WHERE created_at < ?`, cutoff); err != nil {
		return fmt.Errorf("events: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM audit_log WHERE created_at < ?`, cutoff); err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	now := time.Now().UTC()
	if _, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, now); err != nil {
		return fmt.Errorf("sessions: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM invites WHERE expires_at < ? AND consumed_at IS NULL`, now); err != nil {
		return fmt.Errorf("invites: %w", err)
	}
	return nil
}
