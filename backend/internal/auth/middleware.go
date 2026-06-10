package auth

import (
	"context"
	"net/http"

	"github.com/HunterKs/Parv-Creations/backend/internal/models"
)

// AuthMiddleware validates the JWT token from the HttpOnly cookie and sets user info in context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the session cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// No session cookie, unauthorized
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse and validate the JWT token
		claims, err := ParseJWT(cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// Attach claims to the request context for downstream handlers
		ctx := context.WithValue(r.Context(), userClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// contextKey is a type to avoid collisions in context.
type contextKey string

const userClaimsKey contextKey = "user_claims"

// GetClaimsFromContext retrieves the session claims from the request context.
func GetClaimsFromContext(r *http.Request) (*models.SessionClaims, bool) {
	val := r.Context().Value(userClaimsKey)
	if val == nil {
		return nil, false
	}
	claims, ok := val.(*models.SessionClaims)
	return claims, ok
}

// PermissionMiddleware checks if the user has the required permission.
// It should be used after AuthMiddleware.
func PermissionMiddleware(requiredPermission models.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := GetClaimsFromContext(r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// TODO: Implement actual permission check by fetching user role and permissions from DB.
			// For now, we'll allow the request to proceed and note that this is a placeholder.
			// In a real implementation, we would:
			// 1. Fetch the user from the DB using claims.UserID
			// 2. Fetch the user's role
			// 3. Check if the requiredPermission is in the role's permissions
			// 4. If not, return http.StatusForbidden
			_ = requiredPermission // Avoid unused parameter error
			next.ServeHTTP(w, r)
		})
	}
}
