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
