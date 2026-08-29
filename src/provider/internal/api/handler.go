// Package api provides HTTP handlers for the provider API.
//
// API Layer: request routing, CORS, response formatting.
package api

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
	"github.com/quanttide/qtcloud-asset/provider/internal/service"
	"github.com/quanttide/qtcloud-asset/provider/internal/share"
)

type contextKey string

const userContextKey contextKey = "authUser"

const (
	publicObjectURLExpirySeconds       int64 = 0
	defaultRateLimitRequests                 = 120
	defaultRateLimitWindow                   = time.Minute
	defaultLocalLoginRateLimitRequests       = 5
	managedPasswordHashIterations            = 100000
	minManagedAccountLength                  = 3
	maxManagedAccountLength                  = 128
	minManagedPasswordLength                 = 6
	maxManagedPasswordLength                 = 128
	maxShareTitleLength                      = 120
	maxShareZipObjects                       = 4096
	maxShareZipBytes                   int64 = 512 * 1024 * 1024
)

type inviteUserRequest struct {
	Account string `json:"account"`
	// Email is accepted for compatibility with the pre-account API shape.
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     auth.Role `json:"role"`
	Password string    `json:"password"`
}

type updateUserRoleRequest struct {
	Role auth.Role `json:"role"`
}

type localLoginRequest struct {
	Account string `json:"account"`
	// Email is accepted for compatibility with the pre-account API shape.
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createShareRequest struct {
	Title    string   `json:"title"`
	Bucket   string   `json:"bucket"`
	Prefixes []string `json:"prefixes"`
	Keys     []string `json:"keys"`
}

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	cfg                   *config.Config
	buckets               *service.BucketService
	shares                share.Store
	sessions              *auth.Manager
	identity              auth.IdentityProvider
	localAuthenticator    auth.LocalAuthenticator
	users                 auth.UserStore
	audit                 auth.AuditLogStore
	rateLimiter           *RateLimiter
	localLoginRateLimiter *RateLimiter
}

// New creates a new Handler.
func New(cfg *config.Config, buckets *service.BucketService) *Handler {
	sessions := sessionManagerForConfig(cfg)
	handler := NewWithAuth(cfg, buckets, sessions, auth.NotConfiguredIdentityProvider{})
	handler.localAuthenticator = localAuthenticatorFromConfig(cfg)
	return handler
}

func sessionManagerForConfig(cfg *config.Config) *auth.Manager {
	cookieSecure := cfg != nil && strings.HasPrefix(cfg.BaseURL, "https://")
	return auth.NewManager(auth.ManagerOptions{
		CookieSecure:      cookieSecure,
		SameSite:          sameSiteForSessionCookie(cookieSecure),
		SessionSigningKey: sessionSigningKeyForConfig(cfg),
	})
}

func sameSiteForSessionCookie(cookieSecure bool) http.SameSite {
	if cookieSecure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func sessionSigningKeyForConfig(cfg *config.Config) []byte {
	if cfg == nil || cfg.AuthMode != "local" || cfg.LocalAuthPasswordHash == "" {
		return nil
	}
	return []byte("qtcloud-local-session:" + cfg.LocalAuthPasswordHash)
}

// NewWithAuth creates a Handler with explicit authentication dependencies.
func NewWithAuth(cfg *config.Config, buckets *service.BucketService, sessions *auth.Manager, identity auth.IdentityProvider) *Handler {
	return NewWithStores(cfg, buckets, sessions, identity, auth.NewMemoryUserStore(), defaultAuditLogStore())
}

func defaultAuditLogStore() auth.AuditLogStore {
	return auth.NewMultiAuditLogStore(
		auth.NewMemoryAuditLogStore(),
		auth.NewJSONAuditLogStore(os.Stdout),
	)
}

// NewWithStores creates a Handler with explicit auth storage dependencies.
func NewWithStores(cfg *config.Config, buckets *service.BucketService, sessions *auth.Manager, identity auth.IdentityProvider, users auth.UserStore, audit auth.AuditLogStore) *Handler {
	return NewWithStoresAndShares(
		cfg,
		buckets,
		sessions,
		identity,
		users,
		audit,
		share.NewMemoryStore(),
	)
}

// NewWithStoresAndShares creates a Handler with explicit auth and share stores.
func NewWithStoresAndShares(cfg *config.Config, buckets *service.BucketService, sessions *auth.Manager, identity auth.IdentityProvider, users auth.UserStore, audit auth.AuditLogStore, shares share.Store) *Handler {
	if sessions == nil {
		sessions = sessionManagerForConfig(cfg)
	}
	if identity == nil {
		identity = auth.NotConfiguredIdentityProvider{}
	}
	if users == nil {
		users = auth.NewMemoryUserStore()
	}
	if audit == nil {
		audit = auth.NewMemoryAuditLogStore()
	}
	if shares == nil {
		shares = share.NewMemoryStore()
	}
	return &Handler{
		cfg:                   cfg,
		buckets:               buckets,
		shares:                shares,
		sessions:              sessions,
		identity:              identity,
		localAuthenticator:    localAuthenticatorFromConfig(cfg),
		users:                 users,
		audit:                 audit,
		rateLimiter:           NewRateLimiter(defaultRateLimitRequests, defaultRateLimitWindow),
		localLoginRateLimiter: NewRateLimiter(defaultLocalLoginRateLimitRequests, defaultRateLimitWindow),
	}
}

// RegisterRoutes registers all API routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /", h.root)
	mux.HandleFunc("GET /config", h.config)
	mux.Handle("GET /auth/login", h.rateLimit(http.HandlerFunc(h.authLogin)))
	mux.Handle("POST /auth/login", h.rateLimit(http.HandlerFunc(h.authLocalLogin)))
	mux.Handle("GET /auth/callback", h.rateLimit(http.HandlerFunc(h.authCallback)))
	mux.Handle("GET /auth/me", h.rateLimit(h.requireAuth(http.HandlerFunc(h.authMe))))
	mux.Handle("POST /auth/logout", h.rateLimit(h.requireAuth(http.HandlerFunc(h.authLogout))))
	mux.Handle("GET /admin/users", h.rateLimit(h.requireAdmin(http.HandlerFunc(h.adminUsersList))))
	mux.Handle("POST /admin/users", h.rateLimit(h.requireAdmin(http.HandlerFunc(h.adminUsersCreate))))
	mux.Handle("PATCH /admin/users/{id}/role", h.rateLimit(h.requireAdmin(http.HandlerFunc(h.adminUserRoleUpdate))))
	mux.Handle("POST /admin/users/{id}/disable", h.rateLimit(h.requireAdmin(http.HandlerFunc(h.adminUserDisable))))
	mux.Handle("POST /admin/users/{id}/sessions/revoke", h.rateLimit(h.requireAdmin(http.HandlerFunc(h.adminUserSessionsRevoke))))
	mux.Handle("GET /buckets", h.rateLimit(h.requireAuth(http.HandlerFunc(h.bucketsList))))
	mux.Handle("GET /buckets/{name}/objects", h.rateLimit(h.requireAuth(http.HandlerFunc(h.bucketObjectsList))))
	mux.Handle("GET /buckets/{name}/object-url", h.rateLimit(h.requireAuth(http.HandlerFunc(h.objectURL))))
	mux.Handle("POST /shares", h.rateLimit(h.requireAuth(http.HandlerFunc(h.createShare))))
	mux.Handle("GET /shares", h.rateLimit(h.requireAuth(http.HandlerFunc(h.listShares))))
	mux.Handle("GET /shares/{token}", h.rateLimit(http.HandlerFunc(h.getShare)))
	mux.Handle("GET /shares/{token}/objects", h.rateLimit(http.HandlerFunc(h.listShareObjects)))
	mux.Handle("GET /shares/{token}/object-url", h.rateLimit(http.HandlerFunc(h.shareObjectURL)))
	mux.Handle("GET /shares/{token}/download", h.rateLimit(http.HandlerFunc(h.downloadShare)))
	mux.Handle("DELETE /shares/{token}", h.rateLimit(h.requireAuth(http.HandlerFunc(h.revokeShare))))
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, schema.ErrorResponse{Error: http.StatusText(status), Message: msg})
}

func respondUserStoreError(w http.ResponseWriter, err error, fallback string) {
	if errors.Is(err, auth.ErrUserStoreUnavailable) {
		respondError(w, http.StatusServiceUnavailable, "user persistence is unavailable")
		return
	}
	respondError(w, http.StatusInternalServerError, fallback)
}

// CORSMiddleware wraps a handler with CORS headers.
func CORSMiddleware(origins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if originAllowed(origins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func originAllowed(origins []string, origin string) bool {
	for _, allowed := range origins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func (h *Handler) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed, retryAfter := h.rateLimiter.allow(rateLimitKey(r), time.Now())
		if !allowed {
			w.Header().Set("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
			respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func rateLimitKey(r *http.Request) string {
	return clientIP(r) + "|" + r.Method + "|" + r.URL.Path
}

func (h *Handler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		user, _, ok := h.sessions.Authenticate(r, now)
		if !ok {
			h.recordAudit("", auth.AuditActionAuthFailed, auth.AuditResultDenied, r, now)
			respondError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		latestUser, found, err := getUserByID(h.users, user.ID)
		if err != nil {
			log.Printf("load authenticated user error: %v", err)
			h.recordAudit(user.ID, auth.AuditActionAuthFailed, auth.AuditResultFailure, r, now)
			respondError(w, http.StatusServiceUnavailable, "user store is unavailable")
			return
		}
		if !found {
			if _, durable := h.users.(auth.UserStoreWithErrors); durable {
				h.sessions.RevokeFromRequest(r, now)
				h.recordAudit(user.ID, auth.AuditActionAuthFailed, auth.AuditResultDenied, r, now)
				respondError(w, http.StatusUnauthorized, "authentication required")
				return
			}
		}
		if found {
			if latestUser.Status == auth.UserStatusDisabled {
				h.sessions.RevokeFromRequest(r, now)
				h.recordAudit(user.ID, auth.AuditActionAuthFailed, auth.AuditResultDenied, r, now)
				respondError(w, http.StatusForbidden, "user is disabled")
				return
			}
			user = latestUser
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) requireAdmin(next http.Handler) http.Handler {
	return h.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := userFromContext(r.Context())
		if user.Role != auth.RoleAdmin {
			h.recordAudit(user.ID, adminAuditAction(r), auth.AuditResultDenied, r, time.Now())
			respondError(w, http.StatusForbidden, "admin role required")
			return
		}
		if !h.adminMutationOriginAllowed(r) {
			h.recordAudit(user.ID, adminAuditAction(r), auth.AuditResultDenied, r, time.Now())
			respondError(w, http.StatusForbidden, "origin is not allowed")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func (h *Handler) adminMutationOriginAllowed(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	origins := h.cfg.StudioOrigins
	if len(origins) == 0 && h.cfg.StudioOrigin != "" {
		origins = []string{h.cfg.StudioOrigin}
	}
	return originAllowed(origins, origin)
}

func localAuthenticatorFromConfig(cfg *config.Config) auth.LocalAuthenticator {
	if cfg == nil || cfg.AuthMode != "local" {
		return nil
	}
	role := auth.Role(cfg.LocalAuthRole)
	if !validRole(role) {
		role = auth.RoleAdmin
	}
	return auth.NewLocalPasswordAuthenticator(auth.LocalPasswordConfig{
		Account:      cfg.LocalAuthAccount,
		Email:        cfg.LocalAuthEmail,
		Name:         cfg.LocalAuthName,
		Role:         role,
		PasswordHash: cfg.LocalAuthPasswordHash,
	})
}

func userFromContext(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(userContextKey).(auth.User)
	return user, ok
}

func listUsers(store auth.UserStore) ([]auth.User, error) {
	if storeWithErrors, ok := store.(auth.UserStoreWithErrors); ok {
		return storeWithErrors.ListWithError()
	}
	return store.List(), nil
}

func getUserByID(store auth.UserStore, id string) (auth.User, bool, error) {
	if storeWithErrors, ok := store.(auth.UserStoreWithErrors); ok {
		return storeWithErrors.GetByIDWithError(id)
	}
	user, found := store.GetByID(id)
	return user, found, nil
}

func getUserByAccount(store auth.UserStore, account string) (auth.User, bool, error) {
	if storeWithErrors, ok := store.(auth.UserStoreWithErrors); ok {
		return storeWithErrors.GetByAccountWithError(account)
	}
	user, found := store.GetByAccount(account)
	return user, found, nil
}

func updateUserRole(store auth.UserStore, id string, role auth.Role) (auth.User, bool, error) {
	if storeWithErrors, ok := store.(auth.UserStoreWithErrors); ok {
		return storeWithErrors.UpdateRoleWithError(id, role)
	}
	user, found := store.UpdateRole(id, role)
	return user, found, nil
}

func disableUser(store auth.UserStore, id string, disabledAt time.Time) (bool, error) {
	if storeWithErrors, ok := store.(auth.UserStoreWithErrors); ok {
		return storeWithErrors.DisableWithError(id, disabledAt)
	}
	return store.Disable(id, disabledAt), nil
}

func canAccessObjectMetadata(user auth.User, bucketName string) bool {
	return !service.IsMetadataOnlyBucket(bucketName) || user.Role == auth.RoleAdmin
}

func canGenerateObjectURL(user auth.User, bucketName string) bool {
	return !service.IsMetadataOnlyBucket(bucketName)
}

func adminAuditAction(r *http.Request) auth.AuditAction {
	if r.Method == http.MethodGet && r.URL.Path == "/admin/users" {
		return auth.AuditActionListUsers
	}
	if r.Method == http.MethodPost && r.URL.Path == "/admin/users" {
		return auth.AuditActionInviteUser
	}
	if r.Method == http.MethodPatch && strings.HasSuffix(r.URL.Path, "/role") {
		return auth.AuditActionUpdateUserRole
	}
	if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/disable") {
		return auth.AuditActionDisableUser
	}
	if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/sessions/revoke") {
		return auth.AuditActionRevokeSessions
	}
	return auth.AuditActionAuthFailed
}

func validRole(role auth.Role) bool {
	return role == auth.RoleViewer || role == auth.RoleAdmin
}

func validAccount(account string) bool {
	account = auth.NormalizeAccount(account)
	if len(account) < minManagedAccountLength || len(account) > maxManagedAccountLength {
		return false
	}
	for _, ch := range account {
		if ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '.' || ch == '_' || ch == '-' || ch == '@' {
			continue
		}
		return false
	}
	return true
}

func (h *Handler) authLogin(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	state, err := h.sessions.NewState()
	if err != nil {
		log.Printf("auth state error: %v", err)
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusInternalServerError, "failed to start login")
		return
	}
	loginURL, err := h.identity.LoginURL(state)
	if err != nil {
		log.Printf("auth login provider error: %v", err)
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "identity provider is not configured")
		return
	}
	http.SetCookie(w, h.sessions.LoginStateCookie(state))
	h.recordAudit("", auth.AuditActionLogin, auth.AuditResultSuccess, r, now)
	http.Redirect(w, r, loginURL, http.StatusFound)
}

func (h *Handler) authCallback(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" || !auth.ValidateState(r, state) {
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusBadRequest, "invalid login callback")
		return
	}

	user, err := h.identity.Exchange(r.Context(), code, state)
	if err != nil {
		log.Printf("auth callback provider error: %v", err)
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusUnauthorized, "login failed")
		return
	}
	user, err = h.users.UpsertFromIdentity(user, now)
	if err != nil {
		log.Printf("upsert user error: %v", err)
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultFailure, r, now)
		respondUserStoreError(w, err, "failed to save user")
		return
	}
	if user.Status == auth.UserStatusDisabled {
		h.recordAudit(user.ID, auth.AuditActionLogin, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "user is disabled")
		return
	}
	sessionCookie, err := h.sessions.CreateSessionWithMetadata(user, now, clientIP(r), r.UserAgent())
	if err != nil {
		log.Printf("create session error: %v", err)
		h.recordAudit(user.ID, auth.AuditActionLogin, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	http.SetCookie(w, h.sessions.ClearLoginStateCookie())
	http.SetCookie(w, sessionCookie)
	h.recordAudit(user.ID, auth.AuditActionLogin, auth.AuditResultSuccess, r, now)
	http.Redirect(w, r, h.cfg.StudioOrigin, http.StatusSeeOther)
}

func (h *Handler) authLocalLogin(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	allowed, retryAfter := h.localLoginRateLimiter.allow("local-login|"+clientIP(r), now)
	if !allowed {
		w.Header().Set("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}
	if !h.adminMutationOriginAllowed(r) {
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "origin is not allowed")
		return
	}
	var input localLoginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input); err != nil {
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input.Account = auth.NormalizeAccount(firstRequestValue(input.Account, input.Email))
	if input.Account == "" || input.Password == "" {
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "account and password are required")
		return
	}

	user, err := h.authenticateLocalPasswordUser(r.Context(), input.Account, input.Password)
	if err != nil {
		if errors.Is(err, auth.ErrLocalPasswordNotConfigured) {
			h.recordAudit("", auth.AuditActionLogin, auth.AuditResultFailure, r, now)
			respondError(w, http.StatusServiceUnavailable, "local login is not configured")
			return
		}
		if errors.Is(err, auth.ErrUserStoreUnavailable) {
			h.recordAudit("", auth.AuditActionLogin, auth.AuditResultFailure, r, now)
			respondError(w, http.StatusServiceUnavailable, "user store is unavailable")
			return
		}
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	user, err = h.users.UpsertFromIdentity(user, now)
	if err != nil {
		log.Printf("upsert local user error: %v", err)
		h.recordAudit("", auth.AuditActionLogin, auth.AuditResultFailure, r, now)
		respondUserStoreError(w, err, "failed to save user")
		return
	}
	if user.Status == auth.UserStatusDisabled {
		h.recordAudit(user.ID, auth.AuditActionLogin, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "user is disabled")
		return
	}
	sessionCookie, err := h.sessions.CreateSessionWithMetadata(user, now, clientIP(r), r.UserAgent())
	if err != nil {
		log.Printf("create local session error: %v", err)
		h.recordAudit(user.ID, auth.AuditActionLogin, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	http.SetCookie(w, sessionCookie)
	h.recordAudit(user.ID, auth.AuditActionLogin, auth.AuditResultSuccess, r, now)
	respondJSON(w, http.StatusOK, map[string]auth.User{"user": user})
}

func (h *Handler) authenticateLocalPasswordUser(ctx context.Context, account, password string) (auth.User, error) {
	user, ok, err := getUserByAccount(h.users, account)
	if err != nil {
		return auth.User{}, err
	}
	if ok && user.PasswordHash != "" {
		verified, err := auth.VerifyPasswordPBKDF2(password, user.PasswordHash)
		if err != nil {
			return auth.User{}, auth.ErrLocalPasswordNotConfigured
		}
		if !verified {
			return auth.User{}, auth.ErrInvalidCredentials
		}
		return user, nil
	}
	if h.localAuthenticator == nil {
		return auth.User{}, auth.ErrLocalPasswordNotConfigured
	}
	return h.localAuthenticator.Authenticate(ctx, account, password)
}

func (h *Handler) authMe(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	respondJSON(w, http.StatusOK, map[string]auth.User{"user": user})
}

func (h *Handler) authLogout(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	user, _ := userFromContext(r.Context())
	h.sessions.RevokeFromRequest(r, now)
	http.SetCookie(w, h.sessions.ClearSessionCookie())
	h.recordAudit(user.ID, auth.AuditActionLogout, auth.AuditResultSuccess, r, now)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) adminUsersList(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	now := time.Now()
	users, err := listUsers(h.users)
	if err != nil {
		log.Printf("list managed users error: %v", err)
		h.recordAudit(user.ID, auth.AuditActionListUsers, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "user persistence is unavailable")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"users": users, "total": len(users)})
	h.recordAudit(user.ID, auth.AuditActionListUsers, auth.AuditResultSuccess, r, now)
}

func (h *Handler) adminUsersCreate(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	now := time.Now()
	var input inviteUserRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&input); err != nil {
		h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input.Account = auth.NormalizeAccount(firstRequestValue(input.Account, input.Email))
	input.Email = auth.NormalizeAccount(input.Email)
	input.Name = strings.TrimSpace(input.Name)
	input.Password = strings.TrimSpace(input.Password)
	if !validAccount(input.Account) || input.Name == "" || !validRole(input.Role) {
		h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "account, name, role, and password are required")
		return
	}
	if len(input.Password) < minManagedPasswordLength || len(input.Password) > maxManagedPasswordLength {
		h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "password must be between 6 and 128 characters")
		return
	}

	passwordHash, err := auth.HashPasswordPBKDF2(input.Password, managedPasswordHashIterations)
	if err != nil {
		log.Printf("managed user password hash error: %v", err)
		h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusInternalServerError, "failed to save user")
		return
	}

	user, err := h.users.UpsertManaged(auth.User{
		ExternalID:   "managed:" + input.Account,
		Account:      input.Account,
		Email:        input.Email,
		Name:         input.Name,
		Role:         input.Role,
		Status:       auth.UserStatusActive,
		PasswordHash: passwordHash,
	}, now)
	if err != nil {
		log.Printf("managed user upsert error: %v", err)
		h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultFailure, r, now)
		respondUserStoreError(w, err, "failed to save user")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]auth.User{"user": user})
	h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultSuccess, r, now)
}

func firstRequestValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func (h *Handler) adminUserRoleUpdate(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	now := time.Now()
	id := r.PathValue("id")
	var input updateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || !validRole(input.Role) {
		h.recordAudit(actor.ID, auth.AuditActionUpdateUserRole, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "valid role is required")
		return
	}
	if id == actor.ID && input.Role != auth.RoleAdmin {
		h.recordAudit(actor.ID, auth.AuditActionUpdateUserRole, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusBadRequest, "cannot remove your own admin role")
		return
	}

	user, ok, err := updateUserRole(h.users, id, input.Role)
	if err != nil {
		log.Printf("update managed user role error: %v", err)
		h.recordAudit(actor.ID, auth.AuditActionUpdateUserRole, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "user persistence is unavailable")
		return
	}
	if !ok {
		h.recordAudit(actor.ID, auth.AuditActionUpdateUserRole, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]auth.User{"user": user})
	h.recordAudit(actor.ID, auth.AuditActionUpdateUserRole, auth.AuditResultSuccess, r, now)
}

func (h *Handler) adminUserDisable(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	now := time.Now()
	id := r.PathValue("id")
	if id == actor.ID {
		h.recordAudit(actor.ID, auth.AuditActionDisableUser, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusBadRequest, "cannot disable your own account")
		return
	}
	disabled, err := disableUser(h.users, id, now)
	if err != nil {
		log.Printf("disable managed user error: %v", err)
		h.recordAudit(actor.ID, auth.AuditActionDisableUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "user persistence is unavailable")
		return
	}
	if !disabled {
		h.recordAudit(actor.ID, auth.AuditActionDisableUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	revoked := h.sessions.RevokeUserSessions(id, now)
	user, found, err := getUserByID(h.users, id)
	if err != nil {
		log.Printf("load disabled user error: %v", err)
		h.recordAudit(actor.ID, auth.AuditActionDisableUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "user persistence is unavailable")
		return
	}
	if !found {
		h.recordAudit(actor.ID, auth.AuditActionDisableUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"user": user, "revoked": revoked})
	h.recordAudit(actor.ID, auth.AuditActionDisableUser, auth.AuditResultSuccess, r, now)
}

func (h *Handler) adminUserSessionsRevoke(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	now := time.Now()
	id := r.PathValue("id")
	if _, ok, err := getUserByID(h.users, id); err != nil {
		log.Printf("load user for session revoke error: %v", err)
		h.recordAudit(actor.ID, auth.AuditActionRevokeSessions, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "user persistence is unavailable")
		return
	} else if !ok {
		h.recordAudit(actor.ID, auth.AuditActionRevokeSessions, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	revoked := h.sessions.RevokeUserSessions(id, now)

	respondJSON(w, http.StatusOK, map[string]int{"revoked": revoked})
	h.recordAudit(actor.ID, auth.AuditActionRevokeSessions, auth.AuditResultSuccess, r, now)
}

func (h *Handler) recordAudit(userID string, action auth.AuditAction, result auth.AuditResult, r *http.Request, now time.Time) {
	if h.audit == nil {
		return
	}
	entry := auth.AuditLog{
		UserID:    userID,
		Action:    action,
		Target:    auditTarget(r),
		Result:    result,
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
		CreatedAt: now,
	}
	if err := h.audit.Record(entry); err != nil {
		log.Printf("record audit log error: %v", err)
	}
}

func auditTarget(r *http.Request) string {
	if r.URL.Path == "/auth/callback" {
		return r.URL.Path
	}
	if strings.HasPrefix(r.URL.Path, "/shares/") {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/shares/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			parts[0] = "[redacted]"
			return "/shares/" + strings.Join(parts, "/")
		}
	}
	return r.URL.RequestURI()
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if idx := strings.Index(forwarded, ","); idx >= 0 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// health responds with service health status.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, schema.HealthResponse{
		Status:  "ok",
		Service: "qtcloud-asset-provider",
	})
}

// root responds with service information.
func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, schema.RootResponse{
		Name:        "qtcloud-asset-provider",
		Description: "QtCloud Asset API provider",
		Status:      "ready",
	})
}

// config responds with provider configuration.
func (h *Handler) config(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, schema.ConfigResponse{
		ProviderBaseURL: h.cfg.BaseURL,
		StudioOrigin:    h.cfg.StudioOrigin,
		CORS:            "enabled",
	})
}

// bucketsList responds with the discovered OSS buckets (read-only).
// Query params: sort (name/created), order (asc/desc).
func (h *Handler) bucketsList(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	now := time.Now()
	if h.buckets == nil {
		h.recordAudit(user.ID, auth.AuditActionListBuckets, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}

	buckets, err := h.buckets.ListBuckets()
	if err != nil {
		log.Printf("list buckets error: %v", err)
		h.recordAudit(user.ID, auth.AuditActionListBuckets, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusInternalServerError, "failed to list OSS buckets")
		return
	}

	buckets = visibleBucketsForUser(user, buckets)
	sortBuckets(buckets, r.URL.Query().Get("sort"), r.URL.Query().Get("order"))

	respondJSON(w, http.StatusOK, schema.BucketListResponse{
		Buckets: buckets,
		Total:   len(buckets),
	})
	h.recordAudit(user.ID, auth.AuditActionListBuckets, auth.AuditResultSuccess, r, now)
}

func visibleBucketsForUser(user auth.User, buckets []schema.Bucket) []schema.Bucket {
	if user.Role == auth.RoleAdmin {
		return buckets
	}
	visible := make([]schema.Bucket, 0, len(buckets))
	for _, bucket := range buckets {
		if !service.IsMetadataOnlyBucket(bucket.Name) {
			visible = append(visible, bucket)
		}
	}
	return visible
}

// sortBuckets sorts buckets in memory by the given field and order.
func sortBuckets(buckets []schema.Bucket, sortKey, order string) {
	if sortKey == "" {
		return
	}
	desc := order == "desc"
	switch sortKey {
	case "name":
		sort.Slice(buckets, func(i, j int) bool {
			if desc {
				return buckets[i].Name > buckets[j].Name
			}
			return buckets[i].Name < buckets[j].Name
		})
	case "created":
		sort.Slice(buckets, func(i, j int) bool {
			if desc {
				return buckets[i].CreatedAt > buckets[j].CreatedAt
			}
			return buckets[i].CreatedAt < buckets[j].CreatedAt
		})
	}
}

// bucketObjectsList responds with objects inside a bucket (read-only).
// Query params: prefix, sort (key/size/date), order (asc/desc), limit, marker.
func (h *Handler) bucketObjectsList(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	now := time.Now()
	name := r.PathValue("name")
	if name == "" {
		h.recordAudit(user.ID, auth.AuditActionListObjects, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "bucket name is required")
		return
	}
	if !canAccessObjectMetadata(user, name) {
		h.recordAudit(user.ID, auth.AuditActionListObjects, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "bucket object listing is disabled")
		return
	}
	if h.buckets == nil {
		h.recordAudit(user.ID, auth.AuditActionListObjects, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}

	q := r.URL.Query()
	params := schema.ListObjectsParams{
		Prefix: q.Get("prefix"),
		Sort:   q.Get("sort"),
		Order:  q.Get("order"),
		Marker: q.Get("marker"),
	}

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			h.recordAudit(user.ID, auth.AuditActionListObjects, auth.AuditResultFailure, r, now)
			respondError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		params.Limit = n
	}

	objects, nextMarker, truncated, err := h.buckets.ListObjectsAuthorized(name, params)
	if err != nil {
		if errors.Is(err, service.ErrMetadataOnlyBucket) {
			h.recordAudit(user.ID, auth.AuditActionListObjects, auth.AuditResultDenied, r, now)
			respondError(w, http.StatusForbidden, "bucket object listing is disabled")
			return
		}
		log.Printf("list objects in %s error: %v", name, err)
		h.recordAudit(user.ID, auth.AuditActionListObjects, auth.AuditResultFailure, r, now)
		status, message := classifyObjectListError(err)
		respondError(w, status, message)
		return
	}

	respondJSON(w, http.StatusOK, schema.ObjectListResponse{
		Bucket:     name,
		Objects:    objects,
		Total:      len(objects),
		NextMarker: nextMarker,
		Truncated:  truncated,
	})
	h.recordAudit(user.ID, auth.AuditActionListObjects, auth.AuditResultSuccess, r, now)
}

// objectURL responds with an access URL for an object.
// Query params: key (object key), expires (seconds).
func (h *Handler) objectURL(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	now := time.Now()
	name := r.PathValue("name")
	key := r.URL.Query().Get("key")
	if name == "" || key == "" {
		h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "bucket name and object key are required")
		return
	}
	if !canGenerateObjectURL(user, name) {
		h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "bucket object URLs are disabled")
		return
	}
	if h.buckets == nil {
		h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}

	expiresIn, err := parseObjectURLExpiry(r.URL.Query().Get("expires"))
	if err != nil {
		h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	url, err := h.buckets.ObjectURLAuthorized(name, key, expiresIn)
	if err != nil {
		if errors.Is(err, service.ErrMetadataOnlyBucket) || errors.Is(err, service.ErrBucketNotPublic) {
			h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultDenied, r, now)
			respondError(w, http.StatusForbidden, "bucket object URLs are disabled")
			return
		}
		if errors.Is(err, service.ErrBucketACLUnavailable) {
			h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultFailure, r, now)
			respondError(w, http.StatusServiceUnavailable, "bucket access policy could not be verified")
			return
		}
		log.Printf("build url for %s/%s error: %v", name, key, err)
		h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusInternalServerError, "failed to build object url")
		return
	}

	respondJSON(w, http.StatusOK, schema.ObjectURLResponse{
		Bucket:    name,
		Key:       key,
		URL:       url,
		ExpiresIn: expiresIn,
	})
	h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultSuccess, r, now)
}

func (h *Handler) createShare(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	now := time.Now()
	if !h.adminMutationOriginAllowed(r) {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "origin is not allowed")
		return
	}
	if h.shares == nil {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "share service is not configured")
		return
	}
	if h.buckets == nil {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}

	var input createShareRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024)).Decode(&input); err != nil {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Bucket = strings.TrimSpace(input.Bucket)
	if len([]rune(input.Title)) > maxShareTitleLength {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "share title is too long")
		return
	}
	if input.Title == "" {
		input.Title = "公开文件夹"
	}
	if input.Bucket == "" {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "bucket is required")
		return
	}
	if service.IsMetadataOnlyBucket(input.Bucket) {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "metadata-only buckets cannot be shared")
		return
	}
	buckets, err := h.buckets.ListBuckets()
	if err != nil {
		log.Printf("validate share bucket %s error: %v", input.Bucket, err)
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusInternalServerError, "failed to validate bucket")
		return
	}
	foundBucket := false
	for _, bucket := range buckets {
		if bucket.Name == input.Bucket {
			foundBucket = true
			break
		}
	}
	if !foundBucket {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusNotFound, "bucket not found")
		return
	}
	if !h.isShareableBucket(input.Bucket) {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "bucket is not enabled for sharing")
		return
	}
	public, err := h.buckets.IsPublicBucket(input.Bucket)
	if err != nil {
		log.Printf("read share bucket ACL %s error: %v", input.Bucket, err)
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "bucket access policy could not be verified")
		return
	}
	if !public {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "only public-read buckets can be shared")
		return
	}
	prefixes, err := share.NormalizePrefixes(input.Prefixes)
	if err != nil {
		if len(input.Prefixes) > 0 {
			h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		prefixes = nil
	}
	keys, err := share.NormalizeKeys(input.Keys)
	if err != nil {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(prefixes)+len(keys) == 0 {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "at least one folder prefix or file key is required")
		return
	}
	if len(prefixes)+len(keys) > 128 {
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "at most 128 files or folders can be shared")
		return
	}

	record, err := h.shares.Create(share.Record{
		Title:     input.Title,
		Bucket:    input.Bucket,
		Prefixes:  prefixes,
		Keys:      keys,
		CreatedBy: actor.ID,
		CreatedAt: now,
	})
	if err != nil {
		if errors.Is(err, share.ErrStoreUnavailable) {
			h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
			respondError(w, http.StatusServiceUnavailable, "share persistence is unavailable")
			return
		}
		log.Printf("create share error: %v", err)
		h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusInternalServerError, "failed to create share")
		return
	}

	respondJSON(w, http.StatusCreated, schema.FolderShareEnvelope{
		Share: h.shareResponse(record),
	})
	h.recordAudit(actor.ID, auth.AuditActionCreateShare, auth.AuditResultSuccess, r, now)
}

func (h *Handler) listShares(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	now := time.Now()
	if h.shares == nil {
		h.recordAudit(actor.ID, auth.AuditActionListShares, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "share service is not configured")
		return
	}
	records, err := h.shares.ListByOwner(actor.ID)
	if err != nil {
		h.recordAudit(actor.ID, auth.AuditActionListShares, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "share persistence is unavailable")
		return
	}
	shares := make([]schema.FolderShareResponse, 0, len(records))
	for _, record := range records {
		if record.RevokedAt != nil {
			continue
		}
		shares = append(shares, h.shareResponse(record))
	}
	respondJSON(w, http.StatusOK, schema.FolderShareListResponse{
		Shares: shares,
		Total:  len(shares),
	})
	h.recordAudit(actor.ID, auth.AuditActionListShares, auth.AuditResultSuccess, r, now)
}

func (h *Handler) getShare(w http.ResponseWriter, r *http.Request) {
	record, ok, err := h.activeShare(r)
	if err != nil {
		h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultFailure, r, time.Now())
		if errors.Is(err, share.ErrStoreUnavailable) {
			respondError(w, http.StatusServiceUnavailable, "share persistence is unavailable")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load share")
		}
		return
	}
	if !ok {
		h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultDenied, r, time.Now())
		respondError(w, http.StatusNotFound, "share not found")
		return
	}
	if err := h.ensureShareBucketPublic(record.Bucket); err != nil {
		h.respondShareAccessError(w, r, auth.AuditActionViewShare, err)
		return
	}
	respondJSON(w, http.StatusOK, schema.FolderShareEnvelope{
		Share: h.shareResponse(record),
	})
	h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultSuccess, r, time.Now())
}

func (h *Handler) listShareObjects(w http.ResponseWriter, r *http.Request) {
	record, ok, err := h.activeShare(r)
	if err != nil {
		h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultFailure, r, time.Now())
		if errors.Is(err, share.ErrStoreUnavailable) {
			respondError(w, http.StatusServiceUnavailable, "share persistence is unavailable")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load share")
		}
		return
	}
	if !ok {
		h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultDenied, r, time.Now())
		respondError(w, http.StatusNotFound, "share not found")
		return
	}
	if h.buckets == nil {
		h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}
	if err := h.ensureShareBucketPublic(record.Bucket); err != nil {
		h.respondShareAccessError(w, r, auth.AuditActionViewShare, err)
		return
	}

	query := r.URL.Query()
	prefix := query.Get("prefix")
	if !query.Has("prefix") && prefix == "" && len(record.Prefixes) > 0 {
		prefix = record.Prefixes[0]
	}
	if !share.AllowsPrefix(record.Prefixes, record.Keys, prefix) {
		h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultDenied, r, time.Now())
		respondError(w, http.StatusForbidden, "prefix is outside the shared folders")
		return
	}
	params := schema.ListObjectsParams{
		Prefix: prefix,
		Sort:   query.Get("sort"),
		Order:  query.Get("order"),
		Marker: query.Get("marker"),
	}
	if rawLimit := query.Get("limit"); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 {
			h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultFailure, r, time.Now())
			respondError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		params.Limit = limit
	}

	objects, nextMarker, truncated, err := h.buckets.ListObjectsAuthorized(record.Bucket, params)
	if err != nil {
		log.Printf("list shared objects in %s error: %v", record.Bucket, err)
		h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusInternalServerError, "failed to list shared objects")
		return
	}
	filtered := objects[:0]
	for _, object := range objects {
		if share.AllowsObject(record.Prefixes, record.Keys, object.Key) {
			filtered = append(filtered, object)
		}
	}
	respondJSON(w, http.StatusOK, schema.ObjectListResponse{
		Bucket:     record.Bucket,
		Objects:    filtered,
		Total:      len(filtered),
		NextMarker: nextMarker,
		Truncated:  truncated,
	})
	h.recordAudit("", auth.AuditActionViewShare, auth.AuditResultSuccess, r, time.Now())
}

func (h *Handler) shareObjectURL(w http.ResponseWriter, r *http.Request) {
	record, ok, err := h.activeShare(r)
	if err != nil {
		h.recordAudit("", auth.AuditActionShareObjectURL, auth.AuditResultFailure, r, time.Now())
		if errors.Is(err, share.ErrStoreUnavailable) {
			respondError(w, http.StatusServiceUnavailable, "share persistence is unavailable")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load share")
		}
		return
	}
	if !ok {
		h.recordAudit("", auth.AuditActionShareObjectURL, auth.AuditResultDenied, r, time.Now())
		respondError(w, http.StatusNotFound, "share not found")
		return
	}
	if !h.isShareableBucket(record.Bucket) {
		h.recordAudit("", auth.AuditActionShareObjectURL, auth.AuditResultDenied, r, time.Now())
		respondError(w, http.StatusNotFound, "share not found")
		return
	}
	key := r.URL.Query().Get("key")
	if !share.AllowsObject(record.Prefixes, record.Keys, key) {
		h.recordAudit("", auth.AuditActionShareObjectURL, auth.AuditResultDenied, r, time.Now())
		respondError(w, http.StatusForbidden, "object is outside the shared folders")
		return
	}
	if h.buckets == nil {
		h.recordAudit("", auth.AuditActionShareObjectURL, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}
	url, err := h.buckets.ObjectURLAuthorized(record.Bucket, key, publicObjectURLExpirySeconds)
	if err != nil {
		if errors.Is(err, service.ErrBucketNotPublic) {
			h.recordAudit("", auth.AuditActionShareObjectURL, auth.AuditResultDenied, r, time.Now())
			respondError(w, http.StatusNotFound, "share not found")
			return
		}
		if errors.Is(err, service.ErrBucketACLUnavailable) {
			h.recordAudit("", auth.AuditActionShareObjectURL, auth.AuditResultFailure, r, time.Now())
			respondError(w, http.StatusServiceUnavailable, "bucket access policy could not be verified")
			return
		}
		log.Printf("build shared url for %s/%s error: %v", record.Bucket, key, err)
		h.recordAudit("", auth.AuditActionShareObjectURL, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusInternalServerError, "failed to build shared object url")
		return
	}
	respondJSON(w, http.StatusOK, schema.ObjectURLResponse{
		Bucket:    record.Bucket,
		Key:       key,
		URL:       url,
		ExpiresIn: publicObjectURLExpirySeconds,
	})
	h.recordAudit("", auth.AuditActionShareObjectURL, auth.AuditResultSuccess, r, time.Now())
}

func (h *Handler) downloadShare(w http.ResponseWriter, r *http.Request) {
	record, ok, err := h.activeShare(r)
	if err != nil {
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
		if errors.Is(err, share.ErrStoreUnavailable) {
			respondError(w, http.StatusServiceUnavailable, "share persistence is unavailable")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load share")
		}
		return
	}
	if !ok {
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultDenied, r, time.Now())
		respondError(w, http.StatusNotFound, "share not found")
		return
	}
	if h.buckets == nil {
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}
	if err := h.ensureShareBucketPublic(record.Bucket); err != nil {
		h.respondShareAccessError(w, r, auth.AuditActionDownloadShare, err)
		return
	}

	objects, err := h.collectShareObjects(record)
	if err != nil {
		log.Printf("collect shared objects for %s error: %v", record.Bucket, err)
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusInternalServerError, "failed to list shared objects")
		return
	}

	tempFile, err := os.CreateTemp("", "qtcloud-share-*.zip")
	if err != nil {
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusInternalServerError, "failed to prepare share download")
		return
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	archive := zip.NewWriter(tempFile)
	var totalBytes int64
	for _, object := range objects {
		if object.Size > 0 && object.Size > maxShareZipBytes-totalBytes {
			h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
			respondError(w, http.StatusRequestEntityTooLarge, "shared files are too large to archive")
			return
		}

		reader, err := h.buckets.GetObjectAuthorized(record.Bucket, object.Key)
		if err != nil {
			h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
			if errors.Is(err, service.ErrObjectReaderUnavailable) {
				respondError(w, http.StatusServiceUnavailable, "share download is not configured")
			} else {
				log.Printf("read shared object in %s error: %v", record.Bucket, err)
				respondError(w, http.StatusInternalServerError, "failed to read shared object")
			}
			return
		}

		header := &zip.FileHeader{
			Name:   object.Key,
			Method: zip.Store,
		}
		entry, err := archive.CreateHeader(header)
		if err != nil {
			_ = reader.Close()
			h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
			respondError(w, http.StatusInternalServerError, "failed to prepare share archive")
			return
		}
		remaining := maxShareZipBytes - totalBytes
		copied, copyErr := io.Copy(entry, io.LimitReader(reader, remaining+1))
		closeErr := reader.Close()
		if copyErr != nil || closeErr != nil {
			h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
			log.Printf("archive shared object in %s error: %v", record.Bucket, firstError(copyErr, closeErr))
			respondError(w, http.StatusInternalServerError, "failed to archive shared object")
			return
		}
		if copied > remaining {
			h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
			respondError(w, http.StatusRequestEntityTooLarge, "shared files are too large to archive")
			return
		}
		totalBytes += copied
	}

	if err := archive.Close(); err != nil {
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusInternalServerError, "failed to finalize share archive")
		return
	}
	if err := tempFile.Sync(); err != nil {
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusInternalServerError, "failed to finalize share archive")
		return
	}
	info, err := tempFile.Stat()
	if err != nil {
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusInternalServerError, "failed to inspect share archive")
		return
	}
	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
		respondError(w, http.StatusInternalServerError, "failed to read share archive")
		return
	}

	filename := shareArchiveFilename(record.Title)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="share.zip"; filename*=UTF-8''`+url.PathEscape(filename))
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, tempFile); err != nil {
		log.Printf("send shared archive in %s error: %v", record.Bucket, err)
		h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultFailure, r, time.Now())
		return
	}
	h.recordAudit("", auth.AuditActionDownloadShare, auth.AuditResultSuccess, r, time.Now())
}

func (h *Handler) collectShareObjects(record share.Record) ([]schema.Object, error) {
	objectsByKey := make(map[string]schema.Object)
	for _, prefix := range record.Prefixes {
		marker := ""
		for {
			objects, nextMarker, truncated, err := h.buckets.ListObjectsAuthorized(
				record.Bucket,
				schema.ListObjectsParams{Prefix: prefix, Marker: marker},
			)
			if err != nil {
				return nil, err
			}
			for _, object := range objects {
				if !share.AllowsObject(record.Prefixes, record.Keys, object.Key) ||
					object.Type == "Directory" ||
					strings.HasSuffix(object.Key, "/") ||
					!validShareZipEntryName(object.Key) {
					continue
				}
				if _, exists := objectsByKey[object.Key]; exists {
					continue
				}
				if len(objectsByKey) >= maxShareZipObjects {
					return nil, fmt.Errorf("share contains more than %d files", maxShareZipObjects)
				}
				objectsByKey[object.Key] = object
			}
			if !truncated {
				break
			}
			if nextMarker == "" || nextMarker == marker {
				return nil, errors.New("shared object listing pagination did not advance")
			}
			marker = nextMarker
		}
	}
	for _, key := range record.Keys {
		if !validShareZipEntryName(key) {
			return nil, errors.New("shared object key cannot be archived")
		}
		if _, exists := objectsByKey[key]; exists {
			continue
		}
		if len(objectsByKey) >= maxShareZipObjects {
			return nil, fmt.Errorf("share contains more than %d files", maxShareZipObjects)
		}
		objectsByKey[key] = schema.Object{Key: key}
	}

	objects := make([]schema.Object, 0, len(objectsByKey))
	for _, object := range objectsByKey {
		objects = append(objects, object)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})
	return objects, nil
}

func validShareZipEntryName(key string) bool {
	if key == "" || strings.HasPrefix(key, "/") || strings.HasSuffix(key, "/") ||
		strings.ContainsRune(key, '\\') || strings.ContainsRune(key, '\x00') {
		return false
	}
	for _, r := range key {
		if r < 0x20 {
			return false
		}
	}
	for _, segment := range strings.Split(key, "/") {
		if segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func shareArchiveFilename(title string) string {
	title = strings.TrimSpace(title)
	title = strings.Map(func(r rune) rune {
		if r < 0x20 || strings.ContainsRune(`/\:*?"<>|`, r) {
			return '_'
		}
		return r
	}, title)
	title = strings.Trim(title, ". ")
	if title == "" {
		title = "share"
	}
	return title + ".zip"
}

func firstError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) revokeShare(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	now := time.Now()
	if !h.adminMutationOriginAllowed(r) {
		h.recordAudit(actor.ID, auth.AuditActionRevokeShare, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "origin is not allowed")
		return
	}
	token := r.PathValue("token")
	record, ok, err := h.shares.Get(token)
	if err != nil {
		h.recordAudit(actor.ID, auth.AuditActionRevokeShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "share persistence is unavailable")
		return
	}
	if !ok {
		h.recordAudit(actor.ID, auth.AuditActionRevokeShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusNotFound, "share not found")
		return
	}
	if record.CreatedBy != actor.ID && actor.Role != auth.RoleAdmin {
		h.recordAudit(actor.ID, auth.AuditActionRevokeShare, auth.AuditResultDenied, r, now)
		respondError(w, http.StatusForbidden, "share owner or admin role required")
		return
	}
	if _, ok, err := h.shares.Revoke(token, now); err != nil {
		h.recordAudit(actor.ID, auth.AuditActionRevokeShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusServiceUnavailable, "share persistence is unavailable")
		return
	} else if !ok {
		h.recordAudit(actor.ID, auth.AuditActionRevokeShare, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusNotFound, "share not found")
		return
	}
	h.recordAudit(actor.ID, auth.AuditActionRevokeShare, auth.AuditResultSuccess, r, now)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) activeShare(r *http.Request) (share.Record, bool, error) {
	if h.shares == nil {
		return share.Record{}, false, share.ErrStoreUnavailable
	}
	record, ok, err := h.shares.Get(r.PathValue("token"))
	if err != nil {
		return share.Record{}, false, err
	}
	if !ok || record.RevokedAt != nil {
		return share.Record{}, false, nil
	}
	return record, true, nil
}

func (h *Handler) isShareableBucket(bucketName string) bool {
	if h.cfg == nil {
		return false
	}
	for _, allowed := range h.cfg.ShareableBuckets {
		if allowed == bucketName {
			return true
		}
	}
	return false
}

func (h *Handler) ensureShareBucketPublic(bucketName string) error {
	if !h.isShareableBucket(bucketName) {
		return service.ErrBucketNotPublic
	}
	if h.buckets == nil {
		return service.ErrBucketACLUnavailable
	}
	public, err := h.buckets.IsPublicBucket(bucketName)
	if err != nil {
		return err
	}
	if !public {
		return service.ErrBucketNotPublic
	}
	return nil
}

func (h *Handler) respondShareAccessError(w http.ResponseWriter, r *http.Request, action auth.AuditAction, err error) {
	if errors.Is(err, service.ErrBucketNotPublic) {
		h.recordAudit("", action, auth.AuditResultDenied, r, time.Now())
		respondError(w, http.StatusNotFound, "share not found")
		return
	}
	h.recordAudit("", action, auth.AuditResultFailure, r, time.Now())
	respondError(w, http.StatusServiceUnavailable, "bucket access policy could not be verified")
}

func (h *Handler) shareResponse(record share.Record) schema.FolderShareResponse {
	base := strings.TrimRight(h.cfg.BaseURL, "/")
	studioOrigin := strings.TrimRight(h.cfg.StudioOrigin, "/")
	urlBase := studioOrigin
	if urlBase == "" {
		urlBase = base
	}
	return schema.FolderShareResponse{
		Token:     record.Token,
		Title:     record.Title,
		Bucket:    record.Bucket,
		Prefixes:  append([]string(nil), record.Prefixes...),
		Keys:      append([]string(nil), record.Keys...),
		URL:       urlBase + "/#/share/" + url.PathEscape(record.Token),
		CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func parseObjectURLExpiry(raw string) (int64, error) {
	if raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n <= 0 {
			return 0, errors.New("expires must be a positive integer")
		}
	}

	return publicObjectURLExpirySeconds, nil
}
