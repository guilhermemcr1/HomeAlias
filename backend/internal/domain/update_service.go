package domain

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/homealias/homealias/backend/internal/auth"
	"github.com/homealias/homealias/backend/internal/dns"
	cf "github.com/homealias/homealias/backend/internal/dns/cloudflare"
	"github.com/jmoiron/sqlx"
)

type UpdateService struct {
	DB       *sqlx.DB
	Provider dns.Provider
	EncKey   []byte
}

type UpdateRequest struct {
	TokenRaw   string
	Hostname   string
	ClientIP   net.IP
	ClientType string
}

type UpdateResult struct {
	Success            bool      `json:"success"`
	Hostname           string    `json:"hostname,omitempty"`
	IPv4               *string   `json:"ipv4,omitempty"`
	IPv6               *string   `json:"ipv6,omitempty"`
	Changed            bool      `json:"changed"`
	UpdatedAt          time.Time `json:"updated_at"`
	Code               string    `json:"code,omitempty"`
	Message            string    `json:"message,omitempty"`
	DynText            string    `json:"-"`
	DetectedIPFamily   string    `json:"detected_ip_family,omitempty"`
	RequiredIPFamilies []string  `json:"required_ip_families,omitempty"`
}

func (s *UpdateService) Handle(ctx context.Context, req UpdateRequest) UpdateResult {
	now := time.Now().UTC()
	if req.ClientIP.To16() == nil {
		return rejected("unauthorized", "missing client ip", "badauth", now)
	}
	family := "v4"
	rrType := "A"
	if req.ClientIP.To4() == nil {
		family = "v6"
		rrType = "AAAA"
	}
	ipStr := req.ClientIP.String()

	hash := auth.HashDDNSToken(req.TokenRaw)
	var tok DDNSToken
	err := s.DB.GetContext(ctx, &tok, `
		SELECT id, owner_id, name, token_hash, token_prefix, expires_at, status, last_used_at, last_used_ip
		FROM ddns_tokens WHERE token_hash = ?`, hash)
	if err == sql.ErrNoRows {
		_ = s.writeEvent(ctx, nil, nil, nil, req.ClientType, family, &ipStr, nil, false, "rejected", strPtr("invalid token"))
		return rejected("unauthorized", "unauthorized", "badauth", now)
	}
	if err != nil {
		return rejected("temporary", "temporary failure", "911", now)
	}
	if tok.Status != "active" {
		_ = s.writeEvent(ctx, nil, &tok.ID, &tok.OwnerID, req.ClientType, family, &ipStr, nil, false, "rejected", strPtr("token not active"))
		return rejected("unauthorized", "unauthorized", "badauth", now)
	}
	if tok.ExpiresAt != nil && now.After(*tok.ExpiresAt) {
		_, _ = s.DB.ExecContext(ctx, `UPDATE ddns_tokens SET status='expired' WHERE id=?`, tok.ID)
		return rejected("unauthorized", "unauthorized", "badauth", now)
	}

	var ownerStatus string
	_ = s.DB.GetContext(ctx, &ownerStatus, `SELECT status FROM users WHERE id=?`, tok.OwnerID)
	if ownerStatus != "active" {
		return rejected("unauthorized", "unauthorized", "badauth", now)
	}

	hostIDs := []string{}
	if err := s.DB.SelectContext(ctx, &hostIDs, `SELECT host_id FROM token_hosts WHERE token_id=?`, tok.ID); err != nil {
		return rejected("temporary", "temporary failure", "911", now)
	}
	if len(hostIDs) == 0 {
		return rejected("not_found", "no hosts", "nohost", now)
	}

	var host Host
	if req.Hostname != "" {
		fqdn := strings.ToLower(strings.TrimSuffix(req.Hostname, "."))
		err = s.DB.GetContext(ctx, &host, `
			SELECT h.* FROM hosts h
			INNER JOIN token_hosts th ON th.host_id = h.id
			WHERE th.token_id = ? AND h.fqdn = ?`, tok.ID, fqdn)
		if err == sql.ErrNoRows {
			_ = s.writeEvent(ctx, nil, &tok.ID, &tok.OwnerID, req.ClientType, family, &ipStr, nil, false, "rejected", strPtr("hostname not linked"))
			return rejected("not_found", "hostname not found", "nohost", now)
		}
		if err != nil {
			return rejected("temporary", "temporary failure", "911", now)
		}
	} else if len(hostIDs) == 1 {
		err = s.DB.GetContext(ctx, &host, `SELECT * FROM hosts WHERE id=?`, hostIDs[0])
		if err != nil {
			return rejected("temporary", "temporary failure", "911", now)
		}
	} else {
		return rejected("bad_request", "hostname required", "nohost", now)
	}

	if mismatch := ipFamilyMismatch(host, req.ClientIP, now); mismatch != nil {
		_ = s.writeEvent(ctx, &host.ID, &tok.ID, &tok.OwnerID, req.ClientType, family, &ipStr, nil, false, "rejected", strPtr("type not enabled"))
		return *mismatch
	}

	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return rejected("temporary", "temporary failure", "911", now)
	}
	defer func() { _ = tx.Rollback() }()

	var locked Host
	if err := tx.GetContext(ctx, &locked, `SELECT * FROM hosts WHERE id=? FOR UPDATE`, host.ID); err != nil {
		return rejected("temporary", "temporary failure", "911", now)
	}

	var prev *string
	if family == "v4" {
		prev = locked.LastIPv4
	} else {
		prev = locked.LastIPv6
	}

	changed := prev == nil || *prev != ipStr
	if !changed {
		if family == "v4" {
			_, _ = tx.ExecContext(ctx, `UPDATE hosts SET last_seen_v4=? WHERE id=?`, now, locked.ID)
		} else {
			_, _ = tx.ExecContext(ctx, `UPDATE hosts SET last_seen_v6=? WHERE id=?`, now, locked.ID)
		}
		_, _ = tx.ExecContext(ctx, `UPDATE ddns_tokens SET last_used_at=?, last_used_ip=? WHERE id=?`, now, ipStr, tok.ID)
		_ = tx.Commit()
		_ = s.writeEvent(ctx, &locked.ID, &tok.ID, &tok.OwnerID, req.ClientType, family, &ipStr, prev, false, "unchanged", nil)
		res := UpdateResult{Success: true, Hostname: locked.FQDN, Changed: false, UpdatedAt: now, DynText: "nochg " + ipStr}
		if family == "v4" {
			res.IPv4 = &ipStr
		} else {
			res.IPv6 = &ipStr
		}
		return res
	}

	var conn Connection
	if err := tx.GetContext(ctx, &conn, `SELECT * FROM connections WHERE id=?`, locked.ConnectionID); err != nil {
		return rejected("temporary", "temporary failure", "911", now)
	}
	plain, err := auth.Decrypt(s.EncKey, conn.APITokenCipher)
	if err != nil {
		return rejected("temporary", "temporary failure", "911", now)
	}
	pctx := cf.WithToken(ctx, plain)
	err = s.Provider.UpsertRecord(pctx, dns.UpsertInput{
		ZoneID:  locked.ZoneID,
		Type:    rrType,
		Name:    locked.FQDN,
		Content: ipStr,
		TTL:     locked.TTL,
		Proxied: locked.Proxied,
	})
	if err != nil {
		msg := err.Error()
		_, _ = tx.ExecContext(ctx, `UPDATE hosts SET last_error=? WHERE id=?`, msg, locked.ID)
		if strings.Contains(strings.ToLower(msg), "authenticat") || strings.Contains(strings.ToLower(msg), "invalid") {
			_, _ = tx.ExecContext(ctx, `UPDATE connections SET status='invalid', last_error=?, last_checked_at=? WHERE id=?`, msg, now, conn.ID)
		}
		_ = tx.Commit()
		_ = s.writeEvent(ctx, &locked.ID, &tok.ID, &tok.OwnerID, req.ClientType, family, &ipStr, prev, false, "error", &msg)
		return rejected("temporary", "upstream failure", "911", now)
	}

	if family == "v4" {
		_, _ = tx.ExecContext(ctx, `UPDATE hosts SET last_ipv4=?, last_seen_v4=?, last_changed_v4=?, last_error=NULL WHERE id=?`, ipStr, now, now, locked.ID)
	} else {
		_, _ = tx.ExecContext(ctx, `UPDATE hosts SET last_ipv6=?, last_seen_v6=?, last_changed_v6=?, last_error=NULL WHERE id=?`, ipStr, now, now, locked.ID)
	}
	_, _ = tx.ExecContext(ctx, `UPDATE ddns_tokens SET last_used_at=?, last_used_ip=? WHERE id=?`, now, ipStr, tok.ID)
	if err := tx.Commit(); err != nil {
		return rejected("temporary", "temporary failure", "911", now)
	}
	_ = s.writeEvent(ctx, &locked.ID, &tok.ID, &tok.OwnerID, req.ClientType, family, &ipStr, prev, true, "success", nil)
	res := UpdateResult{Success: true, Hostname: locked.FQDN, Changed: true, UpdatedAt: now, DynText: "good " + ipStr}
	if family == "v4" {
		res.IPv4 = &ipStr
	} else {
		res.IPv6 = &ipStr
	}
	return res
}

func (s *UpdateService) writeEvent(ctx context.Context, hostID, tokenID, ownerID *string, clientType, family string, detected, previous *string, changed bool, result string, reason *string) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO update_events (created_at, host_id, token_id, owner_id, client_type, ip_family, detected_ip, previous_ip, changed, result, error_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		time.Now().UTC(), hostID, tokenID, ownerID, clientType, family, detected, previous, changed, result, reason)
	return err
}

func rejected(code, msg, dyn string, now time.Time) UpdateResult {
	return UpdateResult{Success: false, Code: code, Message: msg, DynText: dyn, UpdatedAt: now}
}

// O IP da conexao so pode atualizar um registro da mesma familia. Nao ha
// conversao automatica entre o IPv4 publico e o IPv6 de um cliente.
func ipFamilyMismatch(host Host, ip net.IP, now time.Time) *UpdateResult {
	if (ip.To4() != nil && host.EnableA) || (ip.To4() == nil && host.EnableAAAA) {
		return nil
	}
	message := "Este host nao possui registros A ou AAAA habilitados."
	required := []string{}
	if host.EnableA {
		message = "Este host aceita IPv4 (A). Use uma conexao IPv4, por exemplo curl -4."
		required = append(required, "ipv4")
	}
	if host.EnableAAAA {
		message = "Este host aceita IPv6 (AAAA). Use uma conexao IPv6, por exemplo curl -6."
		required = append(required, "ipv6")
	}
	res := rejected("bad_request", message, "911", now)
	res.DetectedIPFamily = "ipv6"
	if ip.To4() != nil {
		res.DetectedIPFamily = "ipv4"
	}
	res.RequiredIPFamilies = required
	return &res
}

func strPtr(s string) *string { return &s }

func FQDN(zoneName, name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "@" {
		return strings.ToLower(zoneName)
	}
	return strings.ToLower(fmt.Sprintf("%s.%s", name, zoneName))
}
