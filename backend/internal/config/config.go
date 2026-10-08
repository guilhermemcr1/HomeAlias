package config

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr         string
	DatabaseDSN      string
	EncryptionKey    []byte
	SessionKey       []byte
	AdminEmail       string
	AdminPassword    string
	TrustCFHeaders   bool
	CookieSecure     bool
	WarningAfterSec  int
	OfflineAfterSec  int
	RetentionDays    int
	SessionIdle      time.Duration
	SMTPHost         string
	SMTPPort         int
	SMTPUser         string
	SMTPPass         string
	SMTPFrom         string
	TelegramBotToken string
	MigrationsPath   string
	StaticDir        string
}

func Load() (*Config, error) {
	encKey := os.Getenv("HOMEALIAS_ENCRYPTION_KEY")
	sessKey := os.Getenv("HOMEALIAS_SESSION_KEY")
	dsn := os.Getenv("DATABASE_DSN")
	adminEmail := os.Getenv("HOMEALIAS_ADMIN_EMAIL")
	adminPass := os.Getenv("HOMEALIAS_ADMIN_PASSWORD")

	missing := []string{}
	if encKey == "" {
		missing = append(missing, "HOMEALIAS_ENCRYPTION_KEY")
	}
	if sessKey == "" {
		missing = append(missing, "HOMEALIAS_SESSION_KEY")
	}
	if dsn == "" {
		missing = append(missing, "DATABASE_DSN")
	}
	if adminEmail == "" {
		missing = append(missing, "HOMEALIAS_ADMIN_EMAIL")
	}
	if adminPass == "" {
		missing = append(missing, "HOMEALIAS_ADMIN_PASSWORD")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}
	encBytes, err := ParseKey32(encKey)
	if err != nil {
		return nil, fmt.Errorf("HOMEALIAS_ENCRYPTION_KEY: %w", err)
	}
	if len(sessKey) < 32 {
		return nil, fmt.Errorf("HOMEALIAS_SESSION_KEY must be at least 32 bytes (got %d)", len(sessKey))
	}

	if encKey == "0123456789abcdef0123456789abcdef" || sessKey == "0123456789abcdef0123456789abcdefsess" {
		return nil, fmt.Errorf("HOMEALIAS_ENCRYPTION_KEY/HOMEALIAS_SESSION_KEY ainda têm o valor de exemplo; gere novos com: openssl rand -hex 32")
	}
	if len(sessKey) < 32 || uniqueBytes(sessKey) < 8 {
		return nil, fmt.Errorf("HOMEALIAS_SESSION_KEY fraca; gere uma com: openssl rand -hex 32")
	}
	dsn = withParseTime(dsn)

	cfg := &Config{
		HTTPAddr:         envOr("HTTP_ADDR", "127.0.0.1:8080"),
		DatabaseDSN:      dsn,
		EncryptionKey:    encBytes,
		SessionKey:       []byte(sessKey),
		AdminEmail:       adminEmail,
		AdminPassword:    adminPass,
		TrustCFHeaders:   envBool("TRUST_CLOUDFLARE_HEADERS", true),
		CookieSecure:     envBool("COOKIE_SECURE", true),
		WarningAfterSec:  envInt("STATUS_WARNING_AFTER_SEC", 900),
		OfflineAfterSec:  envInt("STATUS_OFFLINE_AFTER_SEC", 3600),
		RetentionDays:    envInt("RETENTION_DAYS", 90),
		SessionIdle:      time.Duration(envInt("SESSION_IDLE_MINUTES", 60)) * time.Minute,
		SMTPHost:         os.Getenv("SMTP_HOST"),
		SMTPPort:         envInt("SMTP_PORT", 587),
		SMTPUser:         os.Getenv("SMTP_USER"),
		SMTPPass:         os.Getenv("SMTP_PASS"),
		SMTPFrom:         os.Getenv("SMTP_FROM"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		MigrationsPath:   envOr("MIGRATIONS_PATH", "file://migrations"),
		StaticDir:        envOr("STATIC_DIR", ""),
	}
	return cfg, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func envBool(key string, def bool) bool {
	v := strings.ToLower(os.Getenv(key))
	if v == "" {
		return def
	}
	return v == "1" || v == "true" || v == "yes"
}

// ParseKey32 accepts raw 32-byte ASCII, 64-char hex, or standard/raw base64 of 32 bytes.
func ParseKey32(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) == 32 {
		return []byte(raw), nil
	}
	if len(raw) == 64 {
		if b, err := hex.DecodeString(raw); err == nil && len(b) == 32 {
			return b, nil
		}
	}
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil && len(b) == 32 {
		return b, nil
	}
	if b, err := base64.RawStdEncoding.DecodeString(raw); err == nil && len(b) == 32 {
		return b, nil
	}
	return nil, fmt.Errorf("need 32 raw bytes, 64 hex chars, or base64 of 32 bytes (got %d chars)", len(raw))
}

func uniqueBytes(s string) int {
	seen := map[rune]struct{}{}
	for _, r := range s {
		seen[r] = struct{}{}
	}
	return len(seen)
}

// withParseTime garante parseTime=true no DSN do MariaDB (o código lê DATETIME como time.Time).
func withParseTime(dsn string) string {
	if strings.Contains(dsn, "parseTime=") {
		return dsn
	}
	if strings.Contains(dsn, "?") {
		return dsn + "&parseTime=true"
	}
	return dsn + "?parseTime=true"
}
