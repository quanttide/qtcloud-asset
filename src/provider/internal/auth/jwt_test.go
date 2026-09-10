package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestJWTVerifierAcceptsRS256JWKAndReturnsClaims(t *testing.T) {
	privateKey, jwk := newJWTTestKey(t)
	verifier, err := NewJWTVerifier(jwk)
	if err != nil {
		t.Fatalf("create JWT verifier: %v", err)
	}

	now := time.Unix(1_757_500_000, 0)
	token := signJWTTestToken(t, privateKey, map[string]any{
		"sub":   "user-uuid",
		"email": "user@example.com",
		"name":  "Test User",
		"exp":   now.Add(time.Hour).Unix(),
	})

	claims, err := verifier.Verify(token, now)
	if err != nil {
		t.Fatalf("verify JWT: %v", err)
	}
	if claims.Subject != "user-uuid" || claims.Email != "user@example.com" || claims.Name != "Test User" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJWTVerifierRejectsInvalidTokens(t *testing.T) {
	privateKey, jwk := newJWTTestKey(t)
	verifier, err := NewJWTVerifier(jwk)
	if err != nil {
		t.Fatalf("create JWT verifier: %v", err)
	}

	now := time.Unix(1_757_500_000, 0)
	tests := []struct {
		name    string
		header  map[string]any
		payload map[string]any
		mutate  func(string) string
	}{
		{
			name:   "missing subject",
			header: map[string]any{"alg": "RS256", "typ": "JWT"},
			payload: map[string]any{
				"exp": now.Add(time.Hour).Unix(),
			},
		},
		{
			name:   "expired",
			header: map[string]any{"alg": "RS256", "typ": "JWT"},
			payload: map[string]any{
				"sub": "user-uuid",
				"exp": now.Add(-time.Minute).Unix(),
			},
		},
		{
			name:   "wrong algorithm",
			header: map[string]any{"alg": "HS256", "typ": "JWT"},
			payload: map[string]any{
				"sub": "user-uuid",
				"exp": now.Add(time.Hour).Unix(),
			},
		},
		{
			name:   "trailing payload data",
			header: map[string]any{"alg": "RS256", "typ": "JWT"},
			payload: map[string]any{
				"sub": "user-uuid",
				"exp": now.Add(time.Hour).Unix(),
			},
			mutate: func(token string) string {
				parts := strings.Split(token, ".")
				parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user-uuid","exp":1799999999} {}`))
				signingInput := parts[0] + "." + parts[1]
				digest := sha256.Sum256([]byte(signingInput))
				signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
				if err != nil {
					t.Fatalf("sign trailing payload token: %v", err)
				}
				parts[2] = base64.RawURLEncoding.EncodeToString(signature)
				return strings.Join(parts, ".")
			},
		},
		{
			name:   "tampered signature",
			header: map[string]any{"alg": "RS256", "typ": "JWT"},
			payload: map[string]any{
				"sub": "user-uuid",
				"exp": now.Add(time.Hour).Unix(),
			},
			mutate: func(token string) string {
				parts := strings.Split(token, ".")
				signature, err := base64.RawURLEncoding.DecodeString(parts[2])
				if err != nil {
					t.Fatalf("decode signature: %v", err)
				}
				signature[0] ^= 1
				parts[2] = base64.RawURLEncoding.EncodeToString(signature)
				return strings.Join(parts, ".")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := signJWTTestTokenWithHeader(t, privateKey, tt.header, tt.payload)
			if tt.mutate != nil {
				token = tt.mutate(token)
			}
			if _, err := verifier.Verify(token, now); err == nil {
				t.Fatal("expected token verification to fail")
			}
		})
	}
}

func TestNewJWTVerifierRejectsMalformedJWK(t *testing.T) {
	for _, raw := range []string{
		"",
		`{"kty":"EC","n":"bad","e":"AQAB"}`,
		`{"kty":"RSA","n":"!","e":"AQAB"}`,
	} {
		if _, err := NewJWTVerifier(raw); err == nil {
			t.Fatalf("expected malformed JWK %q to fail", raw)
		}
	}
}

func newJWTTestKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA test key: %v", err)
	}
	modulus := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())
	jwk, err := json.Marshal(map[string]string{
		"kty": "RSA",
		"n":   modulus,
		"e":   "AQAB",
		"alg": "RS256",
		"use": "sig",
	})
	if err != nil {
		t.Fatalf("marshal JWK: %v", err)
	}
	return privateKey, string(jwk)
}

func signJWTTestToken(t *testing.T, privateKey *rsa.PrivateKey, payload map[string]any) string {
	t.Helper()
	return signJWTTestTokenWithHeader(t, privateKey, map[string]any{
		"alg": "RS256",
		"typ": "JWT",
	}, payload)
}

func signJWTTestTokenWithHeader(t *testing.T, privateKey *rsa.PrivateKey, header, payload map[string]any) string {
	t.Helper()
	encode := func(value any) string {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal JWT part: %v", err)
		}
		return base64.RawURLEncoding.EncodeToString(raw)
	}
	signingInput := encode(header) + "." + encode(payload)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}
