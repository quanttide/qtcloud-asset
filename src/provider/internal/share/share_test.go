package share

import (
	"encoding/base64"
	"testing"
	"time"
)

func TestMemoryStoreCreatesAndReadsShare(t *testing.T) {
	store := NewMemoryStore()
	now := time.Date(2026, time.August, 27, 10, 0, 0, 0, time.UTC)

	created, err := store.Create(Record{
		Title:     "设计稿",
		Bucket:    "qtcloud-asset-studio",
		Prefixes:  []string{"design/", "brand/"},
		CreatedBy: "user-1",
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("create share: %v", err)
	}
	if created.Token == "" {
		t.Fatal("created share should have an opaque token")
	}
	if created.CreatedAt != now {
		t.Fatalf("created timestamp changed: got %v want %v", created.CreatedAt, now)
	}

	loaded, ok, err := store.Get(created.Token)
	if err != nil {
		t.Fatalf("get share: %v", err)
	}
	if !ok {
		t.Fatal("created share should be readable")
	}
	if loaded.Title != "设计稿" || loaded.Bucket != "qtcloud-asset-studio" {
		t.Fatalf("unexpected loaded share: %+v", loaded)
	}
	if len(loaded.Prefixes) != 2 || loaded.Prefixes[0] != "design/" {
		t.Fatalf("unexpected loaded prefixes: %+v", loaded.Prefixes)
	}
}

func TestMemoryStoreRevokeMarksShareInactive(t *testing.T) {
	store := NewMemoryStore()
	created, err := store.Create(Record{
		Bucket:    "qtcloud-asset-studio",
		Prefixes:  []string{"design/"},
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("create share: %v", err)
	}

	revokedAt := time.Date(2026, time.August, 27, 11, 0, 0, 0, time.UTC)
	revoked, ok, err := store.Revoke(created.Token, revokedAt)
	if err != nil {
		t.Fatalf("revoke share: %v", err)
	}
	if !ok {
		t.Fatal("existing share should be revocable")
	}
	if revoked.RevokedAt == nil || !revoked.RevokedAt.Equal(revokedAt) {
		t.Fatalf("unexpected revoked timestamp: %+v", revoked.RevokedAt)
	}

	loaded, ok, err := store.Get(created.Token)
	if err != nil {
		t.Fatalf("get revoked share: %v", err)
	}
	if !ok || loaded.RevokedAt == nil {
		t.Fatalf("revoked share should remain readable for audit: %+v", loaded)
	}
}

func TestMemoryStoreListsSharesByOwnerNewestFirst(t *testing.T) {
	store := NewMemoryStore()
	firstAt := time.Date(2026, time.August, 27, 10, 0, 0, 0, time.UTC)
	secondAt := firstAt.Add(time.Hour)
	for _, record := range []Record{
		{Token: "first", Bucket: "bucket", Prefixes: []string{"a/"}, CreatedBy: "owner", CreatedAt: firstAt},
		{Token: "second", Bucket: "bucket", Prefixes: []string{"b/"}, CreatedBy: "owner", CreatedAt: secondAt},
		{Token: "other", Bucket: "bucket", Prefixes: []string{"c/"}, CreatedBy: "other", CreatedAt: secondAt},
	} {
		if _, err := store.Create(record); err != nil {
			t.Fatalf("create share: %v", err)
		}
	}

	records, err := store.ListByOwner("owner")
	if err != nil {
		t.Fatalf("list owner shares: %v", err)
	}
	if len(records) != 2 || records[0].Token != "second" || records[1].Token != "first" {
		t.Fatalf("unexpected owner shares: %+v", records)
	}
}

func TestMemoryStoreRejectsDuplicateTokens(t *testing.T) {
	store := NewMemoryStore()
	record := Record{
		Token:    "fixed-token",
		Bucket:   "qtcloud-asset-studio",
		Prefixes: []string{"design/"},
	}
	if _, err := store.Create(record); err != nil {
		t.Fatalf("create first share: %v", err)
	}
	if _, err := store.Create(record); err == nil {
		t.Fatal("expected duplicate token to be rejected")
	}
}

func TestNormalizePrefixesCanonicalizesAndRejectsUnsafePaths(t *testing.T) {
	prefixes, err := NormalizePrefixes([]string{"docs/", "design/", "docs/"})
	if err != nil {
		t.Fatalf("normalize prefixes: %v", err)
	}
	if len(prefixes) != 2 || prefixes[0] != "design/" || prefixes[1] != "docs/" {
		t.Fatalf("unexpected normalized prefixes: %+v", prefixes)
	}

	for _, input := range [][]string{
		{"docs"},
		{""},
		{"../private/"},
		{"/absolute/"},
	} {
		if _, err := NormalizePrefixes(input); err == nil {
			t.Fatalf("expected unsafe prefix to be rejected: %+v", input)
		}
	}
}

func TestAllowsKeyUsesFolderPrefixBoundary(t *testing.T) {
	if !AllowsKey([]string{"docs/"}, "docs/readme.md") {
		t.Fatal("expected key under docs/ to be allowed")
	}
	if AllowsKey([]string{"docs/"}, "docs-private/readme.md") {
		t.Fatal("docs/ must not allow docs-private/")
	}
}

func TestNormalizeKeysCanonicalizesAndRejectsUnsafePaths(t *testing.T) {
	keys, err := NormalizeKeys([]string{" docs/readme.md ", "docs/readme.md", "images/logo.svg"})
	if err != nil {
		t.Fatalf("normalize keys: %v", err)
	}
	if len(keys) != 2 || keys[0] != "docs/readme.md" || keys[1] != "images/logo.svg" {
		t.Fatalf("unexpected normalized keys: %#v", keys)
	}

	for _, input := range [][]string{
		{""},
		{"/absolute.txt"},
		{"docs/../secret.txt"},
	} {
		if _, err := NormalizeKeys(input); err == nil {
			t.Fatalf("expected unsafe key to be rejected: %+v", input)
		}
	}
}

func TestAllowsObjectAndPrefixForExplicitFiles(t *testing.T) {
	if !AllowsObject([]string{"docs/"}, []string{"root.txt"}, "docs/readme.md") {
		t.Fatal("expected folder object to be allowed")
	}
	if !AllowsObject(nil, []string{"root.txt"}, "root.txt") {
		t.Fatal("expected explicitly shared file to be allowed")
	}
	if AllowsObject(nil, []string{"root.txt"}, "other.txt") {
		t.Fatal("did not expect unshared file to be allowed")
	}
	if !AllowsPrefix(nil, []string{"docs/readme.md"}, "docs/") {
		t.Fatal("expected prefix containing explicitly shared file to be allowed")
	}
	if AllowsPrefix(nil, []string{"docs/readme.md"}, "private/") {
		t.Fatal("did not expect unrelated prefix to be allowed")
	}
}

func TestTokenEncryptionKeyRequiresSupportedLength(t *testing.T) {
	for _, raw := range []string{"", "short", base64.StdEncoding.EncodeToString([]byte("short"))} {
		if _, err := ParseTokenEncryptionKey(raw); err == nil {
			t.Fatalf("expected invalid encryption key %q to be rejected", raw)
		}
	}
	key := []byte("01234567890123456789012345678901")
	if parsed, err := ParseTokenEncryptionKey(base64.StdEncoding.EncodeToString(key)); err != nil || string(parsed) != string(key) {
		t.Fatalf("expected valid AES-256 key to parse, got %v", err)
	}
}

func TestTokenEncryptionRoundTripDoesNotExposePlaintext(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encrypted, err := encryptToken(key, "opaque-share-token")
	if err != nil {
		t.Fatalf("encrypt token: %v", err)
	}
	if string(encrypted) == "opaque-share-token" {
		t.Fatal("encrypted token must not equal plaintext")
	}
	decrypted, err := decryptToken(key, encrypted)
	if err != nil {
		t.Fatalf("decrypt token: %v", err)
	}
	if decrypted != "opaque-share-token" {
		t.Fatalf("unexpected decrypted token %q", decrypted)
	}
}
