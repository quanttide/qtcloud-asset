package auth

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	localPasswordHashScheme = "pbkdf2_sha256"
	localPasswordSaltBytes  = 16
	localPasswordKeyBytes   = 32
)

var (
	// ErrInvalidCredentials is returned for unknown local accounts or bad passwords.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrLocalPasswordNotConfigured is returned when local auth is selected but incomplete.
	ErrLocalPasswordNotConfigured = errors.New("local password authenticator is not configured")
)

// LocalPasswordConfig carries the single-account MVP local auth configuration.
type LocalPasswordConfig struct {
	Email        string
	Name         string
	Role         Role
	PasswordHash string
}

// LocalAuthenticator verifies username/password credentials.
type LocalAuthenticator interface {
	Authenticate(ctx context.Context, email, password string) (User, error)
}

// LocalPasswordAuthenticator verifies one configured local account.
type LocalPasswordAuthenticator struct {
	email        string
	name         string
	role         Role
	passwordHash string
}

// NewLocalPasswordAuthenticator creates a single-account password authenticator.
func NewLocalPasswordAuthenticator(cfg LocalPasswordConfig) *LocalPasswordAuthenticator {
	role := cfg.Role
	if role == "" {
		role = RoleAdmin
	}
	return &LocalPasswordAuthenticator{
		email:        normalizeLocalEmail(cfg.Email),
		name:         strings.TrimSpace(cfg.Name),
		role:         role,
		passwordHash: strings.TrimSpace(cfg.PasswordHash),
	}
}

// Authenticate validates local credentials and returns a Provider user identity.
func (a *LocalPasswordAuthenticator) Authenticate(_ context.Context, email, password string) (User, error) {
	if a == nil || a.email == "" || a.passwordHash == "" {
		return User{}, ErrLocalPasswordNotConfigured
	}
	if normalizeLocalEmail(email) != a.email {
		return User{}, ErrInvalidCredentials
	}
	if ok, err := VerifyPasswordPBKDF2(password, a.passwordHash); err != nil {
		return User{}, ErrLocalPasswordNotConfigured
	} else if !ok {
		return User{}, ErrInvalidCredentials
	}

	name := a.name
	if name == "" {
		name = a.email
	}
	return User{
		ExternalID: "local:" + a.email,
		Email:      a.email,
		Name:       name,
		Role:       a.role,
		Status:     UserStatusActive,
	}, nil
}

// HashPasswordPBKDF2 encodes a password hash suitable for LOCAL_AUTH_PASSWORD_HASH.
func HashPasswordPBKDF2(password string, iterations int) (string, error) {
	if iterations <= 0 {
		return "", fmt.Errorf("iterations must be positive")
	}
	salt := make([]byte, localPasswordSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, localPasswordKeyBytes)
	if err != nil {
		return "", fmt.Errorf("derive password hash: %w", err)
	}
	return strings.Join([]string{
		localPasswordHashScheme,
		strconv.Itoa(iterations),
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	}, "$"), nil
}

// VerifyPasswordPBKDF2 checks a password against a pbkdf2_sha256 encoded hash.
func VerifyPasswordPBKDF2(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != localPasswordHashScheme {
		return false, fmt.Errorf("unsupported local password hash format")
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false, fmt.Errorf("invalid local password hash iterations")
	}
	salt, err := decodeLocalPasswordBase64(parts[2])
	if err != nil {
		return false, fmt.Errorf("invalid local password hash salt: %w", err)
	}
	expected, err := decodeLocalPasswordBase64(parts[3])
	if err != nil {
		return false, fmt.Errorf("invalid local password hash value: %w", err)
	}
	actual, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(expected))
	if err != nil {
		return false, fmt.Errorf("derive password hash: %w", err)
	}
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func normalizeLocalEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func decodeLocalPasswordBase64(value string) ([]byte, error) {
	if decoded, err := base64.RawStdEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	return base64.StdEncoding.DecodeString(value)
}
