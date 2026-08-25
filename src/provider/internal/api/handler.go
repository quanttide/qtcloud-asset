// Package api provides HTTP handlers for the provider API.
//
// API Layer: request routing, CORS, response formatting.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
	"github.com/quanttide/qtcloud-asset/provider/internal/service"
)

type contextKey string

const userContextKey contextKey = "authUser"

const (
	publicObjectURLExpirySeconds         int64 = 0
	defaultPrivateObjectURLExpirySeconds int64 = 86400
	maxPrivateObjectURLExpirySeconds     int64 = 604800
	defaultRateLimitRequests                   = 120
	defaultRateLimitWindow                     = time.Minute
)

type inviteUserRequest struct {
	Email string    `json:"email"`
	Name  string    `json:"name"`
	Role  auth.Role `json:"role"`
}

type updateUserRoleRequest struct {
	Role auth.Role `json:"role"`
}

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	cfg         *config.Config
	buckets     *service.BucketService
	sessions    *auth.Manager
	identity    auth.IdentityProvider
	users       auth.UserStore
	audit       auth.AuditLogStore
	rateLimiter *RateLimiter
}

// New creates a new Handler.
func New(cfg *config.Config, buckets *service.BucketService) *Handler {
	sessions := auth.NewManager(auth.ManagerOptions{
		CookieSecure: strings.HasPrefix(cfg.BaseURL, "https://"),
	})
	return NewWithAuth(cfg, buckets, sessions, auth.NotConfiguredIdentityProvider{})
}

// NewWithAuth creates a Handler with explicit authentication dependencies.
func NewWithAuth(cfg *config.Config, buckets *service.BucketService, sessions *auth.Manager, identity auth.IdentityProvider) *Handler {
	return NewWithStores(cfg, buckets, sessions, identity, auth.NewMemoryUserStore(), auth.NewMemoryAuditLogStore())
}

// NewWithStores creates a Handler with explicit auth storage dependencies.
func NewWithStores(cfg *config.Config, buckets *service.BucketService, sessions *auth.Manager, identity auth.IdentityProvider, users auth.UserStore, audit auth.AuditLogStore) *Handler {
	if sessions == nil {
		sessions = auth.NewManager(auth.ManagerOptions{})
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
	return &Handler{
		cfg:         cfg,
		buckets:     buckets,
		sessions:    sessions,
		identity:    identity,
		users:       users,
		audit:       audit,
		rateLimiter: NewRateLimiter(defaultRateLimitRequests, defaultRateLimitWindow),
	}
}

// RegisterRoutes registers all API routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /", h.root)
	mux.HandleFunc("GET /config", h.config)
	mux.Handle("GET /auth/login", h.rateLimit(http.HandlerFunc(h.authLogin)))
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
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
		if latestUser, found := h.users.GetByID(user.ID); found {
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

func userFromContext(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(userContextKey).(auth.User)
	return user, ok
}

func canAccessObjectMetadata(user auth.User, bucketName string) bool {
	return !service.IsMetadataOnlyBucket(bucketName) || user.Role == auth.RoleAdmin
}

func canGenerateObjectURL(user auth.User, bucketName string) bool {
	return !service.IsMetadataOnlyBucket(bucketName) || user.Role == auth.RoleAdmin
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

func validEmail(email string) bool {
	email = strings.TrimSpace(email)
	at := strings.Index(email, "@")
	return at > 0 && at < len(email)-1 && strings.Contains(email[at+1:], ".")
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
		respondError(w, http.StatusInternalServerError, "failed to save user")
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
	users := h.users.List()
	respondJSON(w, http.StatusOK, map[string]any{"users": users, "total": len(users)})
	h.recordAudit(user.ID, auth.AuditActionListUsers, auth.AuditResultSuccess, r, now)
}

func (h *Handler) adminUsersCreate(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	now := time.Now()
	var input inviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Name = strings.TrimSpace(input.Name)
	if !validEmail(input.Email) || input.Name == "" || !validRole(input.Role) {
		h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, "email, name, and role are required")
		return
	}

	user, err := h.users.UpsertManaged(auth.User{
		ExternalID: "managed:" + input.Email,
		Email:      input.Email,
		Name:       input.Name,
		Role:       input.Role,
		Status:     auth.UserStatusActive,
	}, now)
	if err != nil {
		log.Printf("managed user upsert error: %v", err)
		h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusInternalServerError, "failed to save user")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]auth.User{"user": user})
	h.recordAudit(actor.ID, auth.AuditActionInviteUser, auth.AuditResultSuccess, r, now)
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

	user, ok := h.users.UpdateRole(id, input.Role)
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
	if !h.users.Disable(id, now) {
		h.recordAudit(actor.ID, auth.AuditActionDisableUser, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	revoked := h.sessions.RevokeUserSessions(id, now)
	user, _ := h.users.GetByID(id)

	respondJSON(w, http.StatusOK, map[string]any{"user": user, "revoked": revoked})
	h.recordAudit(actor.ID, auth.AuditActionDisableUser, auth.AuditResultSuccess, r, now)
}

func (h *Handler) adminUserSessionsRevoke(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	now := time.Now()
	id := r.PathValue("id")
	if _, ok := h.users.GetByID(id); !ok {
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
		respondError(w, http.StatusInternalServerError, "failed to list objects in bucket")
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

	expiresIn, err := parseObjectURLExpiry(name, r.URL.Query().Get("expires"))
	if err != nil {
		h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultFailure, r, now)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	url, err := h.buckets.ObjectURLAuthorized(name, key, expiresIn)
	if err != nil {
		if errors.Is(err, service.ErrMetadataOnlyBucket) {
			h.recordAudit(user.ID, auth.AuditActionObjectURL, auth.AuditResultDenied, r, now)
			respondError(w, http.StatusForbidden, "bucket object URLs are disabled")
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

func parseObjectURLExpiry(bucketName, raw string) (int64, error) {
	var requestedExpiry int64
	if raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n <= 0 {
			return 0, errors.New("expires must be a positive integer")
		}
		requestedExpiry = n
	}

	if !service.IsMetadataOnlyBucket(bucketName) {
		return publicObjectURLExpirySeconds, nil
	}

	expiresIn := defaultPrivateObjectURLExpirySeconds
	if requestedExpiry > 0 {
		expiresIn = requestedExpiry
	}
	if expiresIn > maxPrivateObjectURLExpirySeconds {
		return 0, errors.New("expires exceeds maximum of 604800 seconds")
	}
	return expiresIn, nil
}
