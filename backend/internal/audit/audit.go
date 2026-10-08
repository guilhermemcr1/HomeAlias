package audit

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Writer struct {
	DB *sqlx.DB
}

type Entry struct {
	ActorID         *string
	ActorRole       string
	Action          string
	ResourceType    string
	ResourceID      string
	ResourceOwnerID *string
	IP              string
	Summary         string
}

func (w *Writer) Write(ctx context.Context, e Entry) error {
	_, err := w.DB.ExecContext(ctx, `
		INSERT INTO audit_log (created_at, actor_id, actor_role, action, resource_type, resource_id, resource_owner_id, ip, summary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		time.Now().UTC(), e.ActorID, e.ActorRole, e.Action, e.ResourceType, e.ResourceID, e.ResourceOwnerID, e.IP, e.Summary)
	return err
}
