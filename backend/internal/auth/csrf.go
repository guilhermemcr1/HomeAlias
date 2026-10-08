package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const CSRFCookieName = "homealias_csrf"
const CSRFHeaderName = "X-CSRF-Token"

func NewCSRFToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SetCSRFCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ValidateCSRF(r *http.Request) bool {
	c, err := r.Cookie(CSRFCookieName)
	if err != nil || c.Value == "" {
		return false
	}
	header := r.Header.Get(CSRFHeaderName)
	return header != "" && subtleEqual(header, c.Value)
}

func subtleEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
