-- Plan A Day 3 storage draft for the Provider auth gate.
-- This file is a schema proposal only. Do not execute against production
-- until the platform shared RDS database and migration workflow are confirmed.

CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(128) PRIMARY KEY,
  external_id VARCHAR(191) NOT NULL,
  account VARCHAR(128) NOT NULL,
  email VARCHAR(320) NULL,
  name VARCHAR(191) NOT NULL,
  role VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  password_hash VARCHAR(255) NULL,
  created_at TIMESTAMP NOT NULL,
  last_login_at TIMESTAMP NULL,
  UNIQUE KEY users_external_id_unique (external_id),
  UNIQUE KEY users_account_unique (account),
  UNIQUE KEY users_email_unique (email),
  KEY users_status_idx (status)
);

CREATE TABLE IF NOT EXISTS sessions (
  id VARCHAR(128) PRIMARY KEY,
  user_id VARCHAR(128) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  revoked_at TIMESTAMP NULL,
  ip VARCHAR(64) NOT NULL,
  user_agent VARCHAR(512) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  KEY sessions_user_id_idx (user_id),
  KEY sessions_expires_at_idx (expires_at),
  CONSTRAINT sessions_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id VARCHAR(128) PRIMARY KEY,
  user_id VARCHAR(128) NULL,
  action VARCHAR(64) NOT NULL,
  target VARCHAR(1024) NOT NULL,
  result VARCHAR(32) NOT NULL,
  ip VARCHAR(64) NOT NULL,
  user_agent VARCHAR(512) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  KEY audit_logs_user_id_idx (user_id),
  KEY audit_logs_action_idx (action),
  KEY audit_logs_created_at_idx (created_at),
  CONSTRAINT audit_logs_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS folder_shares (
  token_hash CHAR(64) PRIMARY KEY,
  token_ciphertext VARBINARY(255) NOT NULL,
  title VARCHAR(120) NOT NULL,
  bucket VARCHAR(63) NOT NULL,
  prefixes JSON NOT NULL,
  created_by VARCHAR(128) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  revoked_at TIMESTAMP NULL,
  KEY folder_shares_created_by_idx (created_by),
  KEY folder_shares_created_at_idx (created_at)
);
