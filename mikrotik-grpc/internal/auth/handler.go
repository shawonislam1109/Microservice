package auth

import (
	"context"
	"fmt"
	"isp-management-system/internal/cache"
	"isp-management-system/internal/db"
	"isp-management-system/internal/models"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/vendors/mikrotik"
)

// Handler handles RADIUS authentication requests.
type Handler struct {
	repo  db.Repository
	cache cache.Cache
}

// NewHandler creates a new authentication handler.
func NewHandler(repo db.Repository, cache cache.Cache) *Handler {
	return &Handler{repo: repo, cache: cache}
}

// HandleAccessRequest processes a RADIUS Access-Request.
func (h *Handler) HandleAccessRequest(w radius.ResponseWriter, r *radius.Request) {
	username := rfc2865.UserName_GetString(r.Packet)
	password := rfc2865.UserPassword_GetString(r.Packet)
	nasIP := rfc2865.NASIPAddress_Get(r.Packet)

	log.Printf("RADIUS: Received auth request for user '%s' from NAS %s", username, nasIP)

	// --- 1. Fetch user data (Cache -> DB) ---
	user, err := h.getUser(r.Context(), username)
	if err != nil {
		log.Printf("RADIUS: Auth failed for user '%s': %v", username, err)
		h.sendReject(w, r)
		return
	}

	// --- 2. Validate Password ---
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Printf("RADIUS: Auth failed for user '%s': Invalid password", username)
		h.sendReject(w, r)
		return
	}

	// --- 3. Authorization Checks ---
	if user.Status != "active" {
		log.Printf("RADIUS: Auth failed for user '%s': User is not active (status: %s)", username, user.Status)
		h.sendReject(w, r)
		return
	}
	if time.Now().After(user.ExpiryDate) {
		log.Printf("RADIUS: Auth failed for user '%s': Package has expired on %s", username, user.ExpiryDate)
		h.sendReject(w, r)
		return
	}

	// --- 4. Send Access-Accept ---
	h.sendAccept(w, r, user)
	log.Printf("RADIUS: Successfully authenticated user '%s'", username)
}

// getUser is a helper to get user from cache or database.
func (h *Handler) getUser(ctx context.Context, username string) (*models.User, error) {
	// Try to get from cache first
	cachedUser, err := h.cache.GetUser(ctx, username)
	if err == nil {
		log.Printf("RADIUS: Cache hit for user '%s'", username)
		return cachedUser, nil
	}

	log.Printf("RADIUS: Cache miss for user '%s', fetching from DB", username)
	// If cache miss, get from DB
	dbUser, err := h.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err // Propagate "user not found" or other DB errors
	}

	// Store the fresh user data in cache for next time
	if err := h.cache.SetUser(ctx, dbUser); err != nil {
		log.Printf("RADIUS: Warning: Failed to cache user '%s': %v", username, err)
	}

	return dbUser, nil
}

// sendAccept builds and sends an Access-Accept packet.
func (h *Handler) sendAccept(w radius.ResponseWriter, r *radius.Request, user *models.User) {
	pkt := r.Response(radius.CodeAccessAccept)

	// Add attributes based on the user's package
	if user.PackageRateLimit != "" {
		mikrotik.MikrotikRateLimit_SetString(pkt, user.PackageRateLimit)
	}

	// Example of setting a session timeout (e.g., 24 hours)
	rfc2865.SessionTimeout_Set(pkt, 86400)

	if err := w.Write(pkt); err != nil {
		log.Printf("RADIUS: Failed to send Access-Accept: %v", err)
	}
}

// sendReject builds and sends an Access-Reject packet.
func (h *Handler) sendReject(w radius.ResponseWriter, r *radius.Request) {
	pkt := r.Response(radius.CodeAccessReject)
	if err := w.Write(pkt); err != nil {
		log.Printf("RADIUS: Failed to send Access-Reject: %v", err)
	}
}