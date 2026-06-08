package handlers

import (
	"net/http"
	"time"

	"github.com/HunterKs/Parv-Creations/backend/internal/auth"
	"github.com/HunterKs/Parv-Creations/backend/internal/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler handles user and role CRUD operations.
type UserHandler struct {
	userColl   *mongo.Collection
	roleColl   *mongo.Collection
}

// NewUserHandler creates a new UserHandler with the given collections.
func NewUserHandler(userColl, roleColl *mongo.Collection) *UserHandler {
	return &UserHandler{
		userColl: userColl,
		roleColl: roleColl,
	}
}

// GetUsers handles GET /users
// Returns a list of users (without password hashes)
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	cursor, err := h.userColl.Find(r.Context(), bson.M{})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer cursor.Close(r.Context())

	var users []models.User
	if err = cursor.All(r.Context(), &users); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Remove password hashes from the response
	for i := range users {
		users[i].PasswordHash = ""
	}
	respondJSON(w, http.StatusOK, users)
}

// GetUserByID handles GET /users/{id}
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !primitive.IsObjectIDHex(id) {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	objectID := primitive.ObjectIDHex(id)

	var user models.User
	err := h.userColl.FindOne(r.Context(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	user.PasswordHash = ""
	respondJSON(w, http.StatusOK, user)
}

// CreateUser handles POST /users
// Expected JSON body: { "email": "...", "password": "...", "first_name": "...", "last_name": "...", "role_id": "...", "is_active": true/false }
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		RoleID    string `json:"role_id" binding:"required"`
		IsActive  bool   `json:"is_active"`
	}
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate role ID
	if !primitive.IsObjectIDHex(input.RoleID) {
		respondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}
	roleID := primitive.ObjectIDHex(input.RoleID)

	// Check if role exists
	var role models.Role
	err := h.roleColl.FindOne(r.Context(), bson.M{"_id": roleID}).Decode(&role)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Role not found")
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create the user
	user := models.User{
		Email:        input.Email,
		PasswordHash: hashedPassword,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		RoleID:       roleID,
		IsActive:     input.IsActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	result, err := h.userColl.InsertOne(r.Context(), user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.ID = result.InsertedID.(bson.ObjectID)
	user.PasswordHash = "" // Remove hash from response
	respondJSON(w, http.StatusCreated, user)
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !primitive.IsObjectIDHex(id) {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	objectID := primitive.ObjectIDHex(id)

	var input struct {
		Email     string `json:"email"`
		Password  string `json:"password"` // If provided, we will hash and update
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		RoleID    string `json:"role_id"`
		IsActive  *bool  `json:"is_active"` // Pointer to distinguish between unset and false
	}
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Find the existing user
	var user models.User
	err := h.userColl.FindOne(r.Context(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Prepare update fields
	update := bson.M{}
	if input.Email != "" {
		update["email"] = input.Email
	}
	if input.Password != "" {
		hashedPassword, err := auth.HashPassword(input.Password)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		update["password_hash"] = hashedPassword
	}
	if input.FirstName != "" {
		update["first_name"] = input.FirstName
	}
	if input.LastName != "" {
		update["last_name"] = input.LastName
	}
	if input.RoleID != "" {
		if !primitive.IsObjectIDHex(input.RoleID) {
			respondError(w, http.StatusBadRequest, "Invalid role ID")
			return
		}
		roleID := primitive.ObjectIDHex(input.RoleID)
		// Check if role exists
		var role models.Role
		err := h.roleColl.FindOne(r.Context(), bson.M{"_id": roleID}).Decode(&role)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Role not found")
			return
		}
		update["role_id"] = roleID
	}
	if input.IsActive != nil {
		update["is_active"] = *input.IsActive
	}
	update["updated_at"] = time.Now()

	if len(update) == 0 {
		respondError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	// Update the user
	_, err = h.userColl.UpdateOne(r.Context(), bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch and return the updated user (without password hash)
	err = h.userColl.FindOne(r.Context(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.PasswordHash = ""
	respondJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !primitive.IsObjectIDHex(id) {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	objectID := primitive.ObjectIDHex(id)

	// Delete the user
	result, err := h.userColl.DeleteOne(r.Context(), bson.M{"_id": objectID})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result.DeletedCount == 0 {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "User deleted"})
}

// GetRoles handles GET /roles
func (h *UserHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	cursor, err := h.roleColl.Find(r.Context(), bson.M{})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer cursor.Close(r.Context())

	var roles []models.Role
	if err = cursor.All(r.Context(), &roles); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, roles)
}

// GetRoleByID handles GET /roles/{id}
func (h *UserHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !primitive.IsObjectIDHex(id) {
		respondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}
	objectID := primitive.ObjectIDHex(id)

	var role models.Role
	err := h.roleColl.FindOne(r.Context(), bson.M{"_id": objectID}).Decode(&role)
	if err != nil {
		respondError(w, http.StatusNotFound, "Role not found")
		return
	}
	respondJSON(w, http.StatusOK, role)
}

// CreateRole handles POST /roles
// Expected JSON body: { "name": "...", "description": "...", "permissions": ["perm1", "perm2", ...] }
func (h *UserHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Convert permission strings to Permission type
	var permissions []models.Permission
	for _, p := range input.Permissions {
		permissions = append(permissions, models.Permission(p))
	}

	role := models.Role{
		Name:        input.Name,
		Description: input.Description,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	result, err := h.roleColl.InsertOne(r.Context(), role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	role.ID = result.InsertedID.(bson.ObjectID)
	respondJSON(w, http.StatusCreated, role)
}

// UpdateRole handles PUT /roles/{id}
func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !primitive.IsObjectIDHex(id) {
		respondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}
	objectID := primitive.ObjectIDHex(id)

	var input struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Find the existing role
	var role models.Role
	err := h.roleColl.FindOne(r.Context(), bson.M{"_id": objectID}).Decode(&role)
	if err != nil {
		respondError(w, http.StatusNotFound, "Role not found")
		return
	}

	// Prepare update fields
	update := bson.M{}
	if input.Name != "" {
		update["name"] = input.Name
	}
	if input.Description != "" {
		update["description"] = input.Description
	}
	if len(input.Permissions) > 0 {
		var permissions []models.Permission
		for _, p := range input.Permissions {
			permissions = append(permissions, models.Permission(p))
		}
		update["permissions"] = permissions
	}
	update["updated_at"] = time.Now()

	if len(update) == 0 {
		respondError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	// Update the role
	_, err = h.roleColl.UpdateOne(r.Context(), bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch and return the updated role
	err = h.roleColl.FindOne(r.Context(), bson.M{"_id": objectID}).Decode(&role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, role)
}

// DeleteRole handles DELETE /roles/{id}
func (h *UserHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !primitive.IsObjectIDHex(id) {
		respondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}
	objectID := primitive.ObjectIDHex(id)

	// Delete the role
	result, err := h.roleColl.DeleteOne(r.Context(), bson.M{"_id": objectID})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result.DeletedCount == 0 {
		respondError(w, http.StatusNotFound, "Role not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "Role deleted"})
}

