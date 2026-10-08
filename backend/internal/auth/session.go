package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

const SessionCookieName = "homealias_session"

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	LastSeen  time.Time
	IP        string
}

type SessionStore interface {
	Create(userID, ip string, idle time.Duration) (*Session, error)
	Get(id string) (*Session, error)
	Touch(id string, idle time.Duration) error
	Delete(id string) error
}

func NewSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SetSessionCookie(w http.ResponseWriter, id string, idle time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(idle.Seconds()),
	})
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func ReadSessionCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}
