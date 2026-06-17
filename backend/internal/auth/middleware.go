package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/HunterKs/Parv-Creations/backend/internal/models"
)

// AuthMiddleware validates the JWT token from the HttpOnly cookie and sets user info in context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerTokenFromHeader(r.Header.Get("Authorization"))

		// Get the session cookie
		if token == "" {
			cookie, err := r.Cookie("session_token")
			if err == nil && cookie.Value != "" {
				token = cookie.Value
			}
		}
		if token == "" {
			// No session cookie, unauthorized
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse and validate the JWT token
		claims, err := ParseJWT(token)
		if err != nil {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// Attach claims to the request context for downstream handlers
		ctx := context.WithValue(r.Context(), userClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerTokenFromHeader(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
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
func PermissionMiddleware(requiredPermission models.PermissionKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaimsFromContext(r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.RoleName == "SuperAdmin" {
				next.ServeHTTP(w, r)
				return
			}

			if requiredPermission == "" || hasPermission(claims.Permissions, requiredPermission) {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

func hasPermission(permissions []models.PermissionKey, requiredPermission models.PermissionKey) bool {
	for _, permission := range permissions {
		if permission == requiredPermission {
			return true
		}
	}
	return false
}
