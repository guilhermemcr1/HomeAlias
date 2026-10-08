package auth

import (
	"context"
	"errors"
)

type ctxKey int

const userCtxKey ctxKey = 1

type Actor struct {
	ID    string
	Role  string // admin | user
	Email string
}

func WithActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, userCtxKey, a)
}

func ActorFrom(ctx context.Context) (Actor, bool) {
	a, ok := ctx.Value(userCtxKey).(Actor)
	return a, ok
}

var ErrNotFound = errors.New("not found")

// RequireOwner returns ErrNotFound for non-admin when ownerID != actor.ID
func RequireOwner(actor Actor, ownerID string) error {
	if actor.Role == "admin" {
		return nil
	}
	if actor.ID != ownerID {
		return ErrNotFound
	}
	return nil
}

func AdminOnly(actor Actor) error {
	if actor.Role != "admin" {
		return ErrNotFound
	}
	return nil
}
