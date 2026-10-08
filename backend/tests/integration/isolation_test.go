package integration_test

import (
	"testing"

	"github.com/homealias/homealias/backend/internal/auth"
)

func TestRequireOwnerHidesCrossUser(t *testing.T) {
	actor := auth.Actor{ID: "a", Role: "user"}
	if err := auth.RequireOwner(actor, "b"); err != auth.ErrNotFound {
		t.Fatalf("want ErrNotFound")
	}
	admin := auth.Actor{ID: "a", Role: "admin"}
	if err := auth.RequireOwner(admin, "b"); err != nil {
		t.Fatalf("admin may access")
	}
}
