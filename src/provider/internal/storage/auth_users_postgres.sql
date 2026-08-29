CREATE TABLE IF NOT EXISTS users (
	id VARCHAR(128) PRIMARY KEY,
	external_id VARCHAR(191) NOT NULL,
	account VARCHAR(128) NOT NULL,
	email VARCHAR(320) NULL,
	name VARCHAR(191) NOT NULL,
	role VARCHAR(32) NOT NULL,
	status VARCHAR(32) NOT NULL,
	password_hash VARCHAR(255) NULL,
	created_at TIMESTAMPTZ NOT NULL,
	last_login_at TIMESTAMPTZ NULL,
	CONSTRAINT users_external_id_unique UNIQUE (external_id),
	CONSTRAINT users_account_unique UNIQUE (account),
	CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE INDEX IF NOT EXISTS users_status_idx
	ON users (status);
