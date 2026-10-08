-- HomeAlias: esquema completo (gerado de backend/migrations/*.up.sql). Importe no banco já criado:
--   mariadb -u homealias -p homealias < deploy/sql/schema.sql

-- === 000001_users_sessions.up.sql ===
CREATE TABLE users (
  id CHAR(36) NOT NULL PRIMARY KEY,
  email VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  password_hash VARCHAR(512) NULL,
  role ENUM('admin','user') NOT NULL,
  status ENUM('invite_pending','active','disabled') NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  last_login_at DATETIME(3) NULL,
  UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE sessions (
  id CHAR(64) NOT NULL PRIMARY KEY,
  user_id CHAR(36) NOT NULL,
  expires_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  last_seen_at DATETIME(3) NOT NULL,
  ip VARCHAR(64) NOT NULL DEFAULT '',
  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  KEY idx_sessions_user (user_id),
  KEY idx_sessions_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- === 000002_instance_settings.up.sql ===
CREATE TABLE instance_settings (
  setting_key VARCHAR(128) NOT NULL PRIMARY KEY,
  setting_value TEXT NOT NULL,
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO instance_settings (setting_key, setting_value) VALUES
  ('status_warning_after_sec', '900'),
  ('status_offline_after_sec', '3600'),
  ('retention_days', '90');

-- === 000003_connections_hosts_tokens_events.up.sql ===
CREATE TABLE connections (
  id CHAR(36) NOT NULL PRIMARY KEY,
  owner_id CHAR(36) NOT NULL,
  name VARCHAR(255) NOT NULL,
  provider VARCHAR(64) NOT NULL DEFAULT 'cloudflare',
  api_token_ciphertext TEXT NOT NULL,
  api_token_suffix VARCHAR(32) NOT NULL DEFAULT '',
  account_hint VARCHAR(255) NULL,
  zones_json JSON NULL,
  status ENUM('untested','valid','invalid') NOT NULL DEFAULT 'untested',
  last_checked_at DATETIME(3) NULL,
  last_error TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  CONSTRAINT fk_connections_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
  KEY idx_connections_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE hosts (
  id CHAR(36) NOT NULL PRIMARY KEY,
  owner_id CHAR(36) NOT NULL,
  connection_id CHAR(36) NOT NULL,
  zone_id VARCHAR(64) NOT NULL,
  zone_name VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  fqdn VARCHAR(255) NOT NULL,
  enable_a TINYINT(1) NOT NULL DEFAULT 1,
  enable_aaaa TINYINT(1) NOT NULL DEFAULT 0,
  proxied TINYINT(1) NOT NULL DEFAULT 0,
  ttl INT NOT NULL DEFAULT 1,
  last_ipv4 VARCHAR(64) NULL,
  last_ipv6 VARCHAR(64) NULL,
  last_seen_v4 DATETIME(3) NULL,
  last_seen_v6 DATETIME(3) NULL,
  last_changed_v4 DATETIME(3) NULL,
  last_changed_v6 DATETIME(3) NULL,
  last_error TEXT NULL,
  warning_after_sec INT NULL,
  offline_after_sec INT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  CONSTRAINT fk_hosts_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_hosts_connection FOREIGN KEY (connection_id) REFERENCES connections(id),
  UNIQUE KEY uq_hosts_fqdn (fqdn),
  KEY idx_hosts_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE ddns_tokens (
  id CHAR(36) NOT NULL PRIMARY KEY,
  owner_id CHAR(36) NOT NULL,
  name VARCHAR(255) NOT NULL,
  token_hash CHAR(64) NOT NULL,
  token_prefix VARCHAR(64) NOT NULL,
  expires_at DATETIME(3) NULL,
  status ENUM('active','revoked','expired') NOT NULL DEFAULT 'active',
  last_used_at DATETIME(3) NULL,
  last_used_ip VARCHAR(64) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  CONSTRAINT fk_tokens_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY uq_tokens_hash (token_hash),
  KEY idx_tokens_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE token_hosts (
  token_id CHAR(36) NOT NULL,
  host_id CHAR(36) NOT NULL,
  PRIMARY KEY (token_id, host_id),
  CONSTRAINT fk_th_token FOREIGN KEY (token_id) REFERENCES ddns_tokens(id) ON DELETE CASCADE,
  CONSTRAINT fk_th_host FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE update_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  host_id CHAR(36) NULL,
  token_id CHAR(36) NULL,
  owner_id CHAR(36) NULL,
  client_type ENUM('dyndns','duckdns','agent','windows','unknown') NOT NULL DEFAULT 'unknown',
  ip_family ENUM('v4','v6') NOT NULL,
  detected_ip VARCHAR(64) NULL,
  previous_ip VARCHAR(64) NULL,
  changed TINYINT(1) NOT NULL DEFAULT 0,
  result ENUM('success','unchanged','rejected','error') NOT NULL,
  error_reason TEXT NULL,
  aggregated TINYINT(1) NOT NULL DEFAULT 0,
  KEY idx_events_host_created (host_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- === 000004_invites.up.sql ===
CREATE TABLE invites (
  id CHAR(36) NOT NULL PRIMARY KEY,
  email VARCHAR(255) NOT NULL,
  token_hash CHAR(64) NOT NULL,
  expires_at DATETIME(3) NOT NULL,
  created_by CHAR(36) NOT NULL,
  consumed_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uq_invites_token (token_hash),
  CONSTRAINT fk_invites_creator FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- === 000005_alerts.up.sql ===
CREATE TABLE alert_channels (
  id CHAR(36) NOT NULL PRIMARY KEY,
  owner_id CHAR(36) NULL,
  scope ENUM('user','instance') NOT NULL DEFAULT 'user',
  type ENUM('telegram','email') NOT NULL,
  name VARCHAR(255) NOT NULL,
  destination VARCHAR(512) NOT NULL,
  last_send_status ENUM('ok','error','never') NOT NULL DEFAULT 'never',
  last_send_error TEXT NULL,
  last_send_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_alert_channels_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE alert_rules (
  id CHAR(36) NOT NULL PRIMARY KEY,
  owner_id CHAR(36) NULL,
  created_by_admin TINYINT(1) NOT NULL DEFAULT 0,
  trigger_type ENUM('no_contact','update_failure','ip_changed','token_expiring','connection_invalid') NOT NULL,
  param_json JSON NULL,
  scope_type ENUM('host','connection','user_all','instance') NOT NULL,
  scope_id CHAR(36) NULL,
  min_interval_sec INT NOT NULL DEFAULT 300,
  state ENUM('normal','fired') NOT NULL DEFAULT 'normal',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_alert_rules_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE rule_channels (
  rule_id CHAR(36) NOT NULL,
  channel_id CHAR(36) NOT NULL,
  PRIMARY KEY (rule_id, channel_id),
  CONSTRAINT fk_rc_rule FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE,
  CONSTRAINT fk_rc_channel FOREIGN KEY (channel_id) REFERENCES alert_channels(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- === 000006_audit_log.up.sql ===
CREATE TABLE audit_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  actor_id CHAR(36) NULL,
  actor_role VARCHAR(32) NOT NULL DEFAULT '',
  action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(64) NOT NULL DEFAULT '',
  resource_owner_id CHAR(36) NULL,
  ip VARCHAR(64) NOT NULL DEFAULT '',
  summary TEXT NOT NULL,
  KEY idx_audit_created (created_at),
  KEY idx_audit_actor (actor_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- === 000007_client_types.up.sql ===
ALTER TABLE update_events
  MODIFY client_type ENUM('dyndns','duckdns','agent','windows','shell','docker','unknown') NOT NULL DEFAULT 'unknown';
