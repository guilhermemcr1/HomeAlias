package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPassword(encoded, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}
	memory, time, threads, err := parseArgonParams(parts[3])
	if err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// Limites defensivos para parâmetros lidos do banco (um hash adulterado não pode
// forçar alocação ou tempo arbitrários).
const (
	maxArgonMemory  = 256 * 1024
	maxArgonTime    = 10
	maxArgonThreads = 16
)

func parseArgonParams(s string) (memory, time uint32, threads uint8, err error) {
	if _, err = fmt.Sscanf(s, "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return 0, 0, 0, err
	}
	if memory == 0 || time == 0 || threads == 0 || memory > maxArgonMemory || time > maxArgonTime || threads > maxArgonThreads {
		return 0, 0, 0, fmt.Errorf("argon2 params out of range")
	}
	return memory, time, threads, nil
}

// NeedsRehash informa se o hash foi gerado com parâmetros mais fracos que os atuais.
func NeedsRehash(encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false
	}
	m, t, p, err := parseArgonParams(parts[3])
	if err != nil {
		return false
	}
	return m < argonMemory || t < argonTime || p < argonThreads
}
