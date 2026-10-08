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
