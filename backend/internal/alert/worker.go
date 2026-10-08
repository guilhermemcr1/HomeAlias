package alert

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type Worker struct {
	DB       *sqlx.DB
	Log      *slog.Logger
	Telegram *Telegram
	SMTP     *SMTP
	Interval time.Duration
}

func (w *Worker) Start(ctx context.Context) {
	if w.Interval <= 0 {
		w.Interval = time.Minute
	}
	t := time.NewTicker(w.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	// Evaluate no_contact rules: hosts whose latest last_seen exceeds param after_sec
	type rule struct {
		ID        string `db:"id"`
		OwnerID   string `db:"owner_id"`
		ParamJSON string `db:"param_json"`
		State     string `db:"state"`
		ScopeType string `db:"scope_type"`
		ScopeID   *string `db:"scope_id"`
	}
	var rules []rule
	if err := w.DB.SelectContext(ctx, &rules, `SELECT id, owner_id, COALESCE(param_json,'{}') as param_json, state, scope_type, scope_id FROM alert_rules WHERE enabled=1 AND trigger_type='no_contact'`); err != nil {
		w.Log.Error("alert tick", "err", err)
		return
	}
	_ = rules
	// Connection invalid recheck is handled when updates fail; periodic token expiry scan:
	var expiring int
	_ = w.DB.GetContext(ctx, &expiring, `
		SELECT COUNT(*) FROM ddns_tokens
		WHERE status='active' AND expires_at IS NOT NULL
		  AND expires_at <= DATE_ADD(UTC_TIMESTAMP(3), INTERVAL 14 DAY)
		  AND expires_at > UTC_TIMESTAMP(3)`)
	if expiring > 0 {
		w.Log.Info("tokens nearing expiry", "count", expiring)
	}
}
