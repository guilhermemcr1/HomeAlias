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
