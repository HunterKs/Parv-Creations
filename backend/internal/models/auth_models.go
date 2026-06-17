package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// PermissionKey is the stable machine-readable identifier for an action.
type PermissionKey string

const (
	PermissionViewProduct       PermissionKey = "view_product"
	PermissionAddProduct        PermissionKey = "add_product"
	PermissionEditProduct       PermissionKey = "edit_product"
	PermissionDeleteProduct     PermissionKey = "delete_product"
	PermissionViewUsers         PermissionKey = "view_users"
	PermissionManageUsers       PermissionKey = "manage_users"
	PermissionViewRoles         PermissionKey = "view_roles"
	PermissionManageRoles       PermissionKey = "manage_roles"
	PermissionViewPermissions   PermissionKey = "view_permissions"
	PermissionManagePermissions PermissionKey = "manage_permissions"
)

// Permission represents a system permission document.
type Permission struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key         PermissionKey `bson:"key" json:"key" binding:"required"`
	Name        string        `bson:"name" json:"name" binding:"required"`
	Description string        `bson:"description" json:"description"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}

// AllSystemPermissions returns every built-in permission key assigned to SuperAdmin.
func AllSystemPermissions() []PermissionKey {
	return []PermissionKey{
		PermissionViewProduct,
		PermissionAddProduct,
		PermissionEditProduct,
		PermissionDeleteProduct,
		PermissionViewUsers,
		PermissionManageUsers,
		PermissionViewRoles,
		PermissionManageRoles,
		PermissionViewPermissions,
		PermissionManagePermissions,
	}
}

// Role defines a set of permissions that can be assigned to a user.
type Role struct {
	ID          bson.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string          `bson:"name" json:"name" binding:"required"`
	Description string          `bson:"description" json:"description"`
	Permissions []PermissionKey `bson:"permissions" json:"permissions"`
	CreatedAt   time.Time       `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `bson:"updated_at" json:"updated_at"`
}

// User represents an administrator in the system.
// Password is stored as a bcrypt hash.
// RememberMeTokens are stored in a separate collection (see RememberMeToken).
type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email        string        `bson:"email" json:"email" binding:"required,email"`
	PasswordHash string        `bson:"password_hash" json:"-"` // Never exposed in JSON
	FirstName    string        `bson:"first_name" json:"first_name" binding:"required"`
	LastName     string        `bson:"last_name" json:"last_name" binding:"required"`
	RoleID       bson.ObjectID `bson:"role_id" json:"role_id" binding:"required"` // Reference to Role
	Role         *Role         `bson:"-" json:"role,omitempty"`
	IsActive     bool          `bson:"is_active" json:"is_active"`
	LastLoginAt  time.Time     `bson:"last_login_at" json:"last_login_at"`
	CreatedAt    time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at" json:"updated_at"`
}

// Credentials is used for login requests.
type Credentials struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember"` // Whether to issue a remember-me token
}

// RememberMeToken stores a long-lived token for "remember me" functionality.
// We store a bcrypt hash of the token, and the token is sent to the user as a cookie.
type RememberMeToken struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    bson.ObjectID `bson:"user_id" json:"user_id" binding:"required"`
	TokenHash string        `bson:"token_hash" json:"-"` // bcrypt hash of the token
	ExpiresAt time.Time     `bson:"expires_at" json:"expires_at"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

// Session represents a session stored in a signed cookie.
// We'll use a signed cookie (e.g., using gorilla/sessions or custom) but for simplicity,
// we can store a session ID and look it up in a sessions collection, or we can use JWT.
// However, the requirement says: "Sessions must be tracked securely via HttpOnly cookies using signed state tokens."
// We'll implement a simple session struct that can be signed and validated.
// For now, we'll just define the claims we want in the token.
type SessionClaims struct {
	UserID      string          `json:"user_id"`
	RoleID      string          `json:"role_id"`
	RoleName    string          `json:"role_name"`
	Permissions []PermissionKey `json:"permissions"`
	Email       string          `json:"email"`
	ExpiresAt   int64           `json:"exp"` // Standard JWT exp claim
	IssuedAt    int64           `json:"iat"` // Standard JWT iat claim
}
