package panel

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-sql-driver/mysql"
	"github.com/homealias/homealias/backend/internal/audit"
	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/jmoiron/sqlx"
)

// A small database/sql adapter exercises the real handler without a live server.
type connectionDeleteDB struct {
	owner                 string
	queryErr, deleteErr   error
	deleted               int64
	deleteArgs, auditArgs []driver.NamedValue
}

func (d *connectionDeleteDB) Connect(context.Context) (driver.Conn, error) { return d, nil }
func (d *connectionDeleteDB) Driver() driver.Driver                        { return d }
func (d *connectionDeleteDB) Open(string) (driver.Conn, error)             { return d, nil }
func (d *connectionDeleteDB) Close() error                                 { return nil }
func (d *connectionDeleteDB) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected transaction")
}
func (d *connectionDeleteDB) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (d *connectionDeleteDB) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if query != `SELECT id, owner_id FROM connections WHERE id=?` || len(args) != 1 || args[0].Value != "connection" {
		return nil, errors.New("unexpected lookup")
	}
	return &connectionDeleteRows{owner: d.owner}, d.queryErr
}
func (d *connectionDeleteDB) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if query == `DELETE FROM connections WHERE id=? AND owner_id=?` {
		d.deleteArgs = args
		return driver.RowsAffected(d.deleted), d.deleteErr
	}
	if strings.Contains(query, "INSERT INTO audit_log") {
		d.auditArgs = args
		return driver.RowsAffected(1), nil
	}
	return nil, errors.New("unexpected write")
}

type connectionDeleteRows struct {
	owner string
	done  bool
}

func (*connectionDeleteRows) Columns() []string { return []string{"id", "owner_id"} }
func (*connectionDeleteRows) Close() error      { return nil }
func (r *connectionDeleteRows) Next(values []driver.Value) error {
	if r.done || r.owner == "" {
		return io.EOF
	}
	r.done = true
	values[0], values[1] = "connection", r.owner
	return nil
}

func TestConnectionDelete(t *testing.T) {
	for _, tc := range []struct {
		name                string
		actor               auth.Actor
		owner               string
		queryErr, deleteErr error
		deleted             int64
		want                int
		write               bool
	}{
		{name: "owner", actor: auth.Actor{ID: "owner", Role: "user"}, owner: "owner", deleted: 1, want: 204, write: true},
		{name: "admin", actor: auth.Actor{ID: "admin", Role: "admin"}, owner: "owner", deleted: 1, want: 204, write: true},
		{name: "other user", actor: auth.Actor{ID: "other", Role: "user"}, owner: "owner", want: 404},
		{name: "anonymous", owner: "owner", want: 401},
		{name: "missing", actor: auth.Actor{ID: "owner", Role: "user"}, want: 404},
		{name: "lookup failure", actor: auth.Actor{ID: "owner", Role: "user"}, queryErr: errors.New("database unavailable"), want: 500},
		{name: "linked host or concurrent creation", actor: auth.Actor{ID: "owner", Role: "user"}, owner: "owner", deleteErr: &mysql.MySQLError{Number: 1451}, want: 409, write: true},
		{name: "write failure", actor: auth.Actor{ID: "owner", Role: "user"}, owner: "owner", deleteErr: errors.New("database unavailable"), want: 500, write: true},
		{name: "concurrent removal", actor: auth.Actor{ID: "owner", Role: "user"}, owner: "owner", want: 404, write: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &connectionDeleteDB{owner: tc.owner, queryErr: tc.queryErr, deleteErr: tc.deleteErr, deleted: tc.deleted}
			db := sqlx.NewDb(sql.OpenDB(fake), "mysql")
			defer db.Close()
			api := ConnectionsAPI{DB: db, Audit: &audit.Writer{DB: db}}
			router := chi.NewRouter()
			router.Delete("/api/connections/{id}", api.Delete)
			req := httptest.NewRequest(http.MethodDelete, "/api/connections/connection", nil)
			if tc.actor.ID != "" {
				req = req.WithContext(auth.WithActor(req.Context(), tc.actor))
			}
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tc.want {
				t.Fatalf("status = %d, want %d: %s", res.Code, tc.want, res.Body.String())
			}
			if (len(fake.deleteArgs) > 0) != tc.write {
				t.Fatal("incorrect deletion permission")
			}
			if tc.write && (len(fake.deleteArgs) != 2 || fake.deleteArgs[0].Value != "connection" || fake.deleteArgs[1].Value != tc.owner) {
				t.Fatal("delete must be scoped to the resource and owner")
			}
			if tc.want == 204 {
				if len(fake.auditArgs) != 9 || fake.auditArgs[3].Value != "connection_delete" || fake.auditArgs[6].Value != tc.owner {
					t.Fatal("missing scoped audit entry")
				}
			} else if len(fake.auditArgs) != 0 {
				t.Fatal("failed deletion was recorded as successful")
			}
			if tc.want == 409 && !strings.Contains(res.Body.String(), "hosts vinculados") {
				t.Fatal("conflict must explain how to remove linked hosts")
			}
		})
	}
}
