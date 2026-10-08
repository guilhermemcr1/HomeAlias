CREATE TABLE instance_settings (
  setting_key VARCHAR(128) NOT NULL PRIMARY KEY,
  setting_value TEXT NOT NULL,
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO instance_settings (setting_key, setting_value) VALUES
  ('status_warning_after_sec', '900'),
  ('status_offline_after_sec', '3600'),
  ('retention_days', '90');
