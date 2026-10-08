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
