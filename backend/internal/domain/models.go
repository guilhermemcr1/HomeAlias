package domain

import "time"

type Connection struct {
	ID               string     `db:"id" json:"id"`
	OwnerID          string     `db:"owner_id" json:"owner_id"`
	Name             string     `db:"name" json:"name"`
	Provider         string     `db:"provider" json:"provider"`
	APITokenCipher   string     `db:"api_token_ciphertext" json:"-"`
	APITokenSuffix   string     `db:"api_token_suffix" json:"api_token_suffix"`
	AccountHint      *string    `db:"account_hint" json:"account_hint,omitempty"`
	ZonesJSON        *string    `db:"zones_json" json:"zones,omitempty"`
	Status           string     `db:"status" json:"status"`
	LastCheckedAt    *time.Time `db:"last_checked_at" json:"last_checked_at,omitempty"`
	LastError        *string    `db:"last_error" json:"last_error,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"-"`
	UpdatedAt        time.Time  `db:"updated_at" json:"-"`
}

type Host struct {
	ID              string     `db:"id" json:"id"`
	OwnerID         string     `db:"owner_id" json:"owner_id"`
	ConnectionID    string     `db:"connection_id" json:"connection_id"`
	ZoneID          string     `db:"zone_id" json:"zone_id"`
	ZoneName        string     `db:"zone_name" json:"zone_name"`
	Name            string     `db:"name" json:"name"`
	FQDN            string     `db:"fqdn" json:"fqdn"`
	EnableA         bool       `db:"enable_a" json:"enable_a"`
	EnableAAAA      bool       `db:"enable_aaaa" json:"enable_aaaa"`
	Proxied         bool       `db:"proxied" json:"proxied"`
	TTL             int        `db:"ttl" json:"ttl"`
	LastIPv4        *string    `db:"last_ipv4" json:"last_ipv4,omitempty"`
	LastIPv6        *string    `db:"last_ipv6" json:"last_ipv6,omitempty"`
	LastSeenV4      *time.Time `db:"last_seen_v4" json:"last_seen_v4,omitempty"`
	LastSeenV6      *time.Time `db:"last_seen_v6" json:"last_seen_v6,omitempty"`
	LastChangedV4   *time.Time `db:"last_changed_v4" json:"last_changed_v4,omitempty"`
	LastChangedV6   *time.Time `db:"last_changed_v6" json:"last_changed_v6,omitempty"`
	LastError       *string    `db:"last_error" json:"last_error,omitempty"`
	WarningAfterSec *int       `db:"warning_after_sec" json:"warning_after_sec,omitempty"`
	OfflineAfterSec *int       `db:"offline_after_sec" json:"offline_after_sec,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"-"`
	UpdatedAt       time.Time  `db:"updated_at" json:"-"`
}

type DDNSToken struct {
	ID         string     `db:"id" json:"id"`
	OwnerID    string     `db:"owner_id" json:"owner_id"`
	Name       string     `db:"name" json:"name"`
	TokenHash  string     `db:"token_hash" json:"-"`
	Prefix     string     `db:"token_prefix" json:"token_prefix"`
	ExpiresAt  *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	Status     string     `db:"status" json:"status"`
	LastUsedAt *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
	LastUsedIP *string    `db:"last_used_ip" json:"last_used_ip,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"-"`
}

type UpdateEvent struct {
	ID          int64      `db:"id" json:"id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	HostID      *string    `db:"host_id" json:"host_id,omitempty"`
	TokenID     *string    `db:"token_id" json:"token_id,omitempty"`
	OwnerID     *string    `db:"owner_id" json:"owner_id,omitempty"`
	ClientType  string     `db:"client_type" json:"client_type"`
	IPFamily    string     `db:"ip_family" json:"ip_family"`
	DetectedIP  *string    `db:"detected_ip" json:"detected_ip,omitempty"`
	PreviousIP  *string    `db:"previous_ip" json:"previous_ip,omitempty"`
	Changed     bool       `db:"changed" json:"changed"`
	Result      string     `db:"result" json:"result"`
	ErrorReason *string    `db:"error_reason" json:"error_reason,omitempty"`
	Aggregated  bool       `db:"aggregated" json:"aggregated"`
}
