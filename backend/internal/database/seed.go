package database

import (
	"context"
	"log"
	"time"

	"github.com/HunterKs/Parv-Creations/backend/internal/auth"
	"github.com/HunterKs/Parv-Creations/backend/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	superAdminEmail    = "admin@parvcreations.com"
	superAdminPassword = "password"
	superAdminRoleName = "SuperAdmin"
)

// SeedData provisions baseline permissions, the SuperAdmin role, and the bootstrap user.
func SeedData(userColl, roleColl, permissionColl *mongo.Collection) {
	log.Printf("database.SeedData start")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	permissions := systemPermissionDefinitions()
	if !seedSystemPermissions(ctx, permissionColl, permissions) {
		return
	}

	roleID, ok := ensureSuperAdminRole(ctx, roleColl, systemPermissionKeys(permissions))
	if !ok {
		return
	}

	var existingUser models.User
	err := userColl.FindOne(ctx, bson.M{"email": superAdminEmail}).Decode(&existingUser)
	if err == nil {
		log.Printf("database.SeedData super admin already exists email=%s", superAdminEmail)
		return
	}
	if err != mongo.ErrNoDocuments {
		log.Printf("database.SeedData user lookup failed: %v", err)
		return
	}

	passwordHash, err := auth.HashPassword(superAdminPassword)
	if err != nil {
		log.Printf("database.SeedData password hash failed: %v", err)
		return
	}

	now := time.Now()
	user := models.User{
		Email:        superAdminEmail,
		PasswordHash: passwordHash,
		FirstName:    "Super",
		LastName:     "Admin",
		RoleID:       roleID,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if _, err := userColl.InsertOne(ctx, user); err != nil {
		log.Printf("database.SeedData super admin insert failed: %v", err)
		return
	}

	log.Printf("database.SeedData super admin provisioned email=%s", superAdminEmail)
}

func systemPermissionDefinitions() []models.Permission {
	return []models.Permission{
		{Key: models.PermissionViewProduct, Name: "View Product", Description: "View product records and product listing data."},
		{Key: models.PermissionAddProduct, Name: "Add Product", Description: "Create new product records."},
		{Key: models.PermissionEditProduct, Name: "Edit Product", Description: "Update existing product records."},
		{Key: models.PermissionDeleteProduct, Name: "Delete Product", Description: "Delete product records."},
		{Key: models.PermissionViewUsers, Name: "View Users", Description: "View user management records."},
		{Key: models.PermissionManageUsers, Name: "Manage Users", Description: "Create, update, and delete user records."},
		{Key: models.PermissionViewRoles, Name: "View Roles", Description: "View role management records."},
		{Key: models.PermissionManageRoles, Name: "Manage Roles", Description: "Create, update, and delete role records."},
		{Key: models.PermissionViewPermissions, Name: "View Permissions", Description: "View permission management records."},
		{Key: models.PermissionManagePermissions, Name: "Manage Permissions", Description: "Create, update, and delete permission records."},
	}
}

func systemPermissionKeys(permissions []models.Permission) []models.PermissionKey {
	keys := make([]models.PermissionKey, 0, len(permissions))
	for _, permission := range permissions {
		keys = append(keys, permission.Key)
	}
	return keys
}

func seedSystemPermissions(ctx context.Context, permissionColl *mongo.Collection, permissions []models.Permission) bool {
	now := time.Now()
	for _, permission := range permissions {
		_, err := permissionColl.UpdateOne(
			ctx,
			bson.M{"key": permission.Key},
			bson.M{
				"$set": bson.M{
					"key":         permission.Key,
					"name":        permission.Name,
					"description": permission.Description,
					"updated_at":  now,
				},
				"$setOnInsert": bson.M{
					"created_at": now,
				},
			},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			log.Printf("database.SeedData permission upsert failed key=%s error=%v", permission.Key, err)
			return false
		}
	}
	return true
}

func ensureSuperAdminRole(ctx context.Context, roleColl *mongo.Collection, permissions []models.PermissionKey) (bson.ObjectID, bool) {
	now := time.Now()

	var role models.Role
	err := roleColl.FindOne(ctx, bson.M{"name": superAdminRoleName}).Decode(&role)
	if err == nil {
		_, updateErr := roleColl.UpdateOne(
			ctx,
			bson.M{"_id": role.ID},
			bson.M{"$set": bson.M{
				"permissions": permissions,
				"updated_at":  now,
			}},
		)
		if updateErr != nil {
			log.Printf("database.SeedData super admin role update failed: %v", updateErr)
			return bson.NilObjectID, false
		}
		return role.ID, true
	}
	if err != mongo.ErrNoDocuments {
		log.Printf("database.SeedData role lookup failed: %v", err)
		return bson.NilObjectID, false
	}

	role = models.Role{
		Name:        superAdminRoleName,
		Description: "Full system access for platform bootstrap administration.",
		Permissions: permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := roleColl.InsertOne(ctx, role)
	if err != nil {
		log.Printf("database.SeedData super admin role insert failed: %v", err)
		return bson.NilObjectID, false
	}

	roleID, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		log.Printf("database.SeedData unexpected role id type %T", result.InsertedID)
		return bson.NilObjectID, false
	}

	return roleID, true
}
