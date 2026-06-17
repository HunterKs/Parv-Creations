package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/HunterKs/Parv-Creations/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTSecret is the secret used to sign JWT tokens.
// In production, this should be loaded from an environment variable.
var JWTSecret = []byte("your-super-secret-key-change-in-production")

// TokenDuration is how long a session token is valid.
const TokenDuration = 24 * time.Hour

// RememberMeTokenDuration is how long a remember-me token is valid.
const RememberMeTokenDuration = 30 * 24 * time.Hour

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plaintext password with a hashed password.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT creates a signed JWT token for the given user.
func GenerateJWT(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role_id": user.RoleID.Hex(),
		"email":   user.Email,
		"exp":     time.Now().Add(TokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}
	if user.Role != nil {
		claims["role_name"] = user.Role.Name
		claims["permissions"] = permissionKeysToStrings(user.Role.Permissions)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func permissionKeysToStrings(permissions []models.PermissionKey) []string {
	values := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		values = append(values, string(permission))
	}
	return values
}

// ParseJWT validates and parses a JWT token string into SessionClaims.
func ParseJWT(tokenString string) (*models.SessionClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Convert MapClaims to SessionClaims
		var sessionClaims models.SessionClaims
		if uid, ok := claims["user_id"].(string); ok {
			sessionClaims.UserID = uid
		}
		if rid, ok := claims["role_id"].(string); ok {
			sessionClaims.RoleID = rid
		}
		if roleName, ok := claims["role_name"].(string); ok {
			sessionClaims.RoleName = roleName
		}
		if permissions, ok := claims["permissions"].([]interface{}); ok {
			for _, permission := range permissions {
				if value, ok := permission.(string); ok {
					sessionClaims.Permissions = append(sessionClaims.Permissions, models.PermissionKey(value))
				}
			}
		}
		if permissions, ok := claims["permissions"].([]string); ok {
			for _, permission := range permissions {
				sessionClaims.Permissions = append(sessionClaims.Permissions, models.PermissionKey(permission))
			}
		}
		if email, ok := claims["email"].(string); ok {
			sessionClaims.Email = email
		}
		if exp, ok := claims["exp"].(float64); ok {
			sessionClaims.ExpiresAt = int64(exp)
		}
		if iat, ok := claims["iat"].(float64); ok {
			sessionClaims.IssuedAt = int64(iat)
		}
		return &sessionClaims, nil
	}
	return nil, errors.New("invalid token")
}

// GenerateRememberMeToken creates a random token and its hash for remember-me functionality.
// Returns the plain token (to be sent to the user) and the hash (to be stored in DB).
func GenerateRememberMeToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	hash, err := HashPassword(token)
	if err != nil {
		return "", "", err
	}
	return token, hash, nil
}

// CompareRememberMeToken compares a plain token with its hash.
func CompareRememberMeToken(token, hash string) bool {
	return CheckPasswordHash(token, hash)
}
