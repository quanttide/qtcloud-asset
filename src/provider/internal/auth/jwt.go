package auth

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"
)

// JWTClaims contains the account-system identity fields used by Provider.
type JWTClaims struct {
	Subject string
	Email   string
	Name    string
}

// JWTVerifier verifies account-system RS256 access tokens.
type JWTVerifier struct {
	keys       map[string]*rsa.PublicKey
	defaultKey *rsa.PublicKey
}

type jwtJSONWebKey struct {
	KTY string `json:"kty"`
	KID string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwtJSONWebKeySet struct {
	Keys []jwtJSONWebKey `json:"keys"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	KID string `json:"kid"`
}

type jwtPayload struct {
	Subject string      `json:"sub"`
	Email   string      `json:"email"`
	Name    string      `json:"name"`
	Exp     json.Number `json:"exp"`
	NBF     json.Number `json:"nbf"`
}

// NewJWTVerifier parses a single RSA JWK or a JWK set.
func NewJWTVerifier(raw string) (*JWTVerifier, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("AUTH_JWT_PUBLIC_JWK is missing")
	}

	var single jwtJSONWebKey
	if err := json.Unmarshal([]byte(raw), &single); err == nil && single.KTY != "" {
		return newJWTVerifier([]jwtJSONWebKey{single})
	}

	var set jwtJSONWebKeySet
	if err := json.Unmarshal([]byte(raw), &set); err != nil {
		return nil, fmt.Errorf("parse AUTH_JWT_PUBLIC_JWK: %w", err)
	}
	if len(set.Keys) == 0 {
		return nil, errors.New("AUTH_JWT_PUBLIC_JWK contains no keys")
	}
	return newJWTVerifier(set.Keys)
}

func newJWTVerifier(keys []jwtJSONWebKey) (*JWTVerifier, error) {
	verifier := &JWTVerifier{keys: make(map[string]*rsa.PublicKey, len(keys))}
	for _, key := range keys {
		if key.KTY != "RSA" {
			return nil, errors.New("AUTH_JWT_PUBLIC_JWK must contain RSA keys")
		}
		if key.Alg != "" && key.Alg != "RS256" {
			return nil, errors.New("AUTH_JWT_PUBLIC_JWK only supports RS256")
		}
		if key.Use != "" && key.Use != "sig" {
			return nil, errors.New("AUTH_JWT_PUBLIC_JWK key use must be sig")
		}
		publicKey, err := parseRSAPublicKey(key.N, key.E)
		if err != nil {
			return nil, err
		}
		if key.KID == "" {
			if verifier.defaultKey != nil {
				return nil, errors.New("AUTH_JWT_PUBLIC_JWK has multiple keys without kid")
			}
			verifier.defaultKey = publicKey
			continue
		}
		if _, exists := verifier.keys[key.KID]; exists {
			return nil, fmt.Errorf("AUTH_JWT_PUBLIC_JWK contains duplicate kid %q", key.KID)
		}
		verifier.keys[key.KID] = publicKey
	}
	if verifier.defaultKey == nil && len(verifier.keys) == 0 {
		return nil, errors.New("AUTH_JWT_PUBLIC_JWK contains no usable keys")
	}
	return verifier, nil
}

func parseRSAPublicKey(modulus, exponent string) (*rsa.PublicKey, error) {
	nBytes, err := decodeJWTBase64(modulus)
	if err != nil || len(nBytes) == 0 {
		return nil, errors.New("AUTH_JWT_PUBLIC_JWK has invalid RSA modulus")
	}
	eBytes, err := decodeJWTBase64(exponent)
	if err != nil || len(eBytes) == 0 || len(eBytes) > 4 {
		return nil, errors.New("AUTH_JWT_PUBLIC_JWK has invalid RSA exponent")
	}
	var e int
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}
	if e < 3 || e%2 == 0 {
		return nil, errors.New("AUTH_JWT_PUBLIC_JWK has invalid RSA exponent")
	}
	publicKey := &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}
	if publicKey.N.BitLen() < 2048 {
		return nil, errors.New("AUTH_JWT_PUBLIC_JWK RSA modulus must be at least 2048 bits")
	}
	return publicKey, nil
}

// Verify validates an RS256 token and returns its stable account identity.
func (v *JWTVerifier) Verify(token string, now time.Time) (JWTClaims, error) {
	if v == nil {
		return JWTClaims{}, errors.New("JWT verifier is not configured")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return JWTClaims{}, errors.New("JWT must have three segments")
	}
	headerBytes, err := decodeJWTBase64(parts[0])
	if err != nil {
		return JWTClaims{}, errors.New("JWT header is invalid")
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Alg != "RS256" {
		return JWTClaims{}, errors.New("JWT algorithm must be RS256")
	}
	publicKey := v.defaultKey
	if header.KID != "" {
		publicKey = v.keys[header.KID]
	} else if publicKey == nil {
		return JWTClaims{}, errors.New("JWT key id is required")
	}
	if publicKey == nil {
		return JWTClaims{}, errors.New("JWT key id is unknown")
	}

	signature, err := decodeJWTBase64(parts[2])
	if err != nil {
		return JWTClaims{}, errors.New("JWT signature is invalid")
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature); err != nil {
		return JWTClaims{}, errors.New("JWT signature verification failed")
	}

	payloadBytes, err := decodeJWTBase64(parts[1])
	if err != nil {
		return JWTClaims{}, errors.New("JWT payload is invalid")
	}
	var payload jwtPayload
	decoder := json.NewDecoder(bytes.NewReader(payloadBytes))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return JWTClaims{}, errors.New("JWT payload is invalid")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return JWTClaims{}, errors.New("JWT payload is invalid")
	}
	if strings.TrimSpace(payload.Subject) == "" {
		return JWTClaims{}, errors.New("JWT subject is missing")
	}
	exp, err := jwtNumericDate(payload.Exp)
	if err != nil || !now.Before(exp) {
		return JWTClaims{}, errors.New("JWT is expired")
	}
	if payload.NBF != "" {
		notBefore, err := jwtNumericDate(payload.NBF)
		if err != nil || now.Before(notBefore) {
			return JWTClaims{}, errors.New("JWT is not active")
		}
	}
	return JWTClaims{
		Subject: strings.TrimSpace(payload.Subject),
		Email:   strings.TrimSpace(payload.Email),
		Name:    strings.TrimSpace(payload.Name),
	}, nil
}

func jwtNumericDate(value json.Number) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("JWT numeric date is missing")
	}
	seconds, err := value.Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, 0), nil
}

func decodeJWTBase64(value string) ([]byte, error) {
	if decoded, err := base64.RawURLEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	return base64.URLEncoding.DecodeString(value)
}
