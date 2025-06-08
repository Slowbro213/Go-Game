// auth.go
package middleware

import (
	"context"
	"fmt"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

// AuthService handles authentication middleware
type AuthService struct {
	store *sessions.CookieStore
	log   *log.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(store *sessions.CookieStore, logger *log.Logger) *AuthService {
	return &AuthService{
		store: store,
		log:   logger,
	}
}

// AuthResult represents authentication result
type AuthResult struct {
	Authenticated bool
	UserID        string
	Error         error
}

const (
	sessionName           = "poked-cookie"
	sessionAuthenticated  = "authenticated"
	ContextPlayerID       = "player_id"
)

// AuthMiddleware verifies session authentication
func (a *AuthService) AuthMiddleware() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			result := a.checkAuth(r)
			
			if !result.Authenticated {
				if result.Error != nil {
					a.log.Printf("Unauthenticated request: %v", result.Error)
					http.Error(w, "Please login first", http.StatusUnauthorized)
					return
				}

				a.log.Printf("Authentication failed: %v", result.Error)
				http.ServeFile(w,r,"../views/errors/duplicate.html")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(r.Context(), ContextPlayerID, result.UserID)
			r = r.WithContext(ctx)

			next(w, r)
		}
	}
}

// checkAuth performs session validation
func (a *AuthService) checkAuth(r *http.Request) AuthResult {
	session, err := a.store.Get(r, sessionName)
	if err != nil {
		return AuthResult{Error: fmt.Errorf("session retrieval failed: %w", err)}
	}

	auth, ok := session.Values[sessionAuthenticated].(bool)
	if !ok || !auth {
		return AuthResult{Error: errors.New("not authenticated")}
	}

	userID, ok := session.Values[ContextPlayerID].(string)
	if !ok {
		return AuthResult{Error: errors.New("invalid user ID in session")}
	}

	return AuthResult{
		Authenticated: true,
		UserID:        userID,
	}
}

// RequireRole checks user permissions (example implementation)
func (a *AuthService) RequireRole(role string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("userID").(int)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Simplified role check - replace with your actual logic
			hasRole := a.checkUserRole(userID, role)
			if !hasRole {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}

// checkUserRole - implement your actual role checking logic here
func (a *AuthService) checkUserRole(userID int, requiredRole string) bool {
	// In a real application, you would:
	// 1. Look up the user in your database
	// 2. Check their roles/permissions
	// 3. Return true if they have the required role
	
	// This is a placeholder implementation
	return true // Change to your actual logic
}
