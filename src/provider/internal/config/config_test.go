package config

import "testing"

func TestLoadUsesProductionGatewayBaseURLByDefault(t *testing.T) {
	t.Setenv("PROVIDER_BASE_URL", "")

	cfg := Load()

	if cfg.BaseURL != "https://api.quanttide.com/qtcloud-asset" {
		t.Fatalf("expected production gateway base URL, got %q", cfg.BaseURL)
	}
}

func TestLoadUsesPostgresRDSDriverByDefault(t *testing.T) {
	t.Setenv("RDS_DRIVER", "")

	cfg := Load()

	if cfg.RDSDriver != "postgres" {
		t.Fatalf("expected PostgreSQL RDS driver by default, got %q", cfg.RDSDriver)
	}
}

func TestLoadUsesFormalStudioOriginAndKeepsCompatibilityOrigin(t *testing.T) {
	t.Setenv("STUDIO_ORIGIN", "")

	cfg := Load()

	if cfg.StudioOrigin != "https://asset.cloud.quanttide.com" {
		t.Fatalf("expected formal Studio origin, got %q", cfg.StudioOrigin)
	}

	if !contains(cfg.StudioOrigins, "https://asset.cloud.quanttide.com") {
		t.Fatal("formal Studio origin is missing from the CORS allowlist")
	}
	if !contains(cfg.StudioOrigins, "https://asset.quanttide.com") {
		t.Fatal("compatibility Studio origin is missing from the CORS allowlist")
	}
}

func TestLoadAllowsStudioOriginOverride(t *testing.T) {
	t.Setenv("STUDIO_ORIGIN", "https://preview.example.com")

	cfg := Load()

	if cfg.StudioOrigin != "https://preview.example.com" {
		t.Fatalf("expected configured Studio origin, got %q", cfg.StudioOrigin)
	}
	if !contains(cfg.StudioOrigins, "https://preview.example.com") {
		t.Fatal("configured Studio origin is missing from the CORS allowlist")
	}
}

func TestLoadReadsLocalAuthConfiguration(t *testing.T) {
	t.Setenv("AUTH_MODE", "local")
	t.Setenv("LOCAL_AUTH_ACCOUNT", "admin")
	t.Setenv("LOCAL_AUTH_EMAIL", "admin@example.com")
	t.Setenv("LOCAL_AUTH_NAME", "Admin User")
	t.Setenv("LOCAL_AUTH_ROLE", "admin")
	t.Setenv("LOCAL_AUTH_PASSWORD_HASH", "pbkdf2_sha256$1000$salt$hash")

	cfg := Load()

	if cfg.AuthMode != "local" || cfg.LocalAuthAccount != "admin" || cfg.LocalAuthEmail != "admin@example.com" {
		t.Fatalf("expected local auth config to load, got %#v", cfg)
	}
	if cfg.LocalAuthName != "Admin User" || cfg.LocalAuthRole != "admin" {
		t.Fatalf("expected local auth identity metadata, got %#v", cfg)
	}
	if cfg.LocalAuthPasswordHash != "pbkdf2_sha256$1000$salt$hash" {
		t.Fatalf("expected local auth password hash to load, got %#v", cfg)
	}
}

func TestLoadKeepsLegacyLocalAuthEmailAsAccountFallback(t *testing.T) {
	t.Setenv("AUTH_MODE", "local")
	t.Setenv("LOCAL_AUTH_ACCOUNT", "")
	t.Setenv("LOCAL_AUTH_EMAIL", "admin@example.com")

	cfg := Load()

	if cfg.LocalAuthAccount != "admin@example.com" {
		t.Fatalf("expected legacy local auth email to populate account fallback, got %#v", cfg)
	}
}

func TestLoadUsesFCExecutionRoleCredentialsWhenExplicitCredentialsAreMissing(t *testing.T) {
	t.Setenv("OSS_ACCESS_KEY_ID", "")
	t.Setenv("OSS_ACCESS_KEY_SECRET", "")
	t.Setenv("OSS_SESSION_TOKEN", "")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "role-ak")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "role-secret")
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "role-token")

	cfg := Load()

	if cfg.OSSAccessKeyID != "role-ak" {
		t.Fatalf("expected FC role access key ID, got %q", cfg.OSSAccessKeyID)
	}
	if cfg.OSSAccessKeySecret != "role-secret" {
		t.Fatalf("expected FC role access key secret, got %q", cfg.OSSAccessKeySecret)
	}
	if cfg.OSSSecurityToken != "role-token" {
		t.Fatalf("expected FC role security token, got %q", cfg.OSSSecurityToken)
	}
}

func TestLoadPrefersExplicitOSSSecretOverFCExecutionRoleCredentials(t *testing.T) {
	t.Setenv("OSS_ACCESS_KEY_ID", "explicit-ak")
	t.Setenv("OSS_ACCESS_KEY_SECRET", "explicit-secret")
	t.Setenv("OSS_SESSION_TOKEN", "explicit-token")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "role-ak")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "role-secret")
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "role-token")

	cfg := Load()

	if cfg.OSSAccessKeyID != "explicit-ak" ||
		cfg.OSSAccessKeySecret != "explicit-secret" ||
		cfg.OSSSecurityToken != "explicit-token" {
		t.Fatalf("expected explicit OSS credentials to take precedence, got %#v", cfg)
	}
}

func TestLoadReadsSharePersistenceAndAllowlistConfiguration(t *testing.T) {
	t.Setenv("RDS_DRIVER", "mysql")
	t.Setenv("RDS_CONNECTION_STRING", "user:password@tcp(rds.internal:3306)/asset?parseTime=true")
	t.Setenv("SHARE_STORE", "rds")
	t.Setenv("SHARE_MIGRATION", "folder-shares-postgres-v1")
	t.Setenv("SHARE_TOKEN_ENCRYPTION_KEY", "base64-key-material")
	t.Setenv("SHAREABLE_BUCKETS", "qtcloud-asset-studio, qtcloud-data-studio,qtcloud-asset-studio")

	cfg := Load()

	if cfg.RDSDriver != "mysql" || cfg.RDSConnectionString == "" {
		t.Fatalf("expected RDS configuration, got %#v", cfg)
	}
	if cfg.ShareStoreMode != "rds" || cfg.ShareTokenEncryptionKey != "base64-key-material" {
		t.Fatalf("expected share persistence configuration, got %#v", cfg)
	}
	if cfg.ShareMigration != "folder-shares-postgres-v1" {
		t.Fatalf("expected share migration configuration, got %#v", cfg)
	}
	if len(cfg.ShareableBuckets) != 2 ||
		cfg.ShareableBuckets[0] != "qtcloud-asset-studio" ||
		cfg.ShareableBuckets[1] != "qtcloud-data-studio" {
		t.Fatalf("expected deduplicated share bucket allowlist, got %#v", cfg.ShareableBuckets)
	}
}

func TestLoadReadsUserPersistenceConfiguration(t *testing.T) {
	t.Setenv("USER_STORE", "memory")
	t.Setenv("USER_MIGRATION", "users-postgres-v1")

	cfg := Load()

	if cfg.UserStoreMode != "memory" {
		t.Fatalf("expected user store mode, got %q", cfg.UserStoreMode)
	}
	if cfg.UserMigration != "users-postgres-v1" {
		t.Fatalf("expected user migration, got %q", cfg.UserMigration)
	}
}

func TestLoadDoesNotMixPartialExplicitCredentialsWithRoleCredentials(t *testing.T) {
	t.Setenv("OSS_ACCESS_KEY_ID", "partial-ak")
	t.Setenv("OSS_ACCESS_KEY_SECRET", "")
	t.Setenv("OSS_SESSION_TOKEN", "partial-token")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "role-ak")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "role-secret")
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "role-token")

	cfg := Load()

	if cfg.OSSAccessKeyID != "role-ak" ||
		cfg.OSSAccessKeySecret != "role-secret" ||
		cfg.OSSSecurityToken != "role-token" {
		t.Fatalf("expected complete role credentials after partial explicit config, got %#v", cfg)
	}
}
