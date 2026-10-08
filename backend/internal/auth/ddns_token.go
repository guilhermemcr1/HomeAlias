package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func GenerateDDNSToken() (raw string, hash string, prefix string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	prefix = fmt.Sprintf("ha_%s", raw[:8])
	return raw, hash, prefix, nil
}

func HashDDNSToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
