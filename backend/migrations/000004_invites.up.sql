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
