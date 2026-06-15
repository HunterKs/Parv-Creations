package database

import (
	"context"
	"log"
	"time"

	"github.com/HunterKs/Parv-Creations/backend/internal/auth"
	"github.com/HunterKs/Parv-Creations/backend/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	superAdminEmail    = "admin@parvcreations.com"
	superAdminPassword = "password"
	superAdminRoleName = "SuperAdmin"
)

// SeedData provisions the baseline SuperAdmin role and user if they do not exist.
func SeedData(userColl, roleColl *mongo.Collection) {
	log.Printf("database.SeedData start")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

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

	roleID, ok := ensureSuperAdminRole(ctx, roleColl)
	if !ok {
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

func ensureSuperAdminRole(ctx context.Context, roleColl *mongo.Collection) (bson.ObjectID, bool) {
	now := time.Now()
	permissions := models.AllSystemPermissions()

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
