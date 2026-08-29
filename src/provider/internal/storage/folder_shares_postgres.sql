CREATE TABLE IF NOT EXISTS folder_shares (
	token_hash CHAR(64) PRIMARY KEY,
	token_ciphertext BYTEA NOT NULL,
	title VARCHAR(120) NOT NULL,
	bucket VARCHAR(63) NOT NULL,
	prefixes JSONB NOT NULL,
	created_by VARCHAR(128) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	revoked_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS folder_shares_created_by_idx
	ON folder_shares (created_by);

CREATE INDEX IF NOT EXISTS folder_shares_created_at_idx
	ON folder_shares (created_at);
