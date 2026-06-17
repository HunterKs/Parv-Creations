package handlers

import (
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/HunterKs/Parv-Creations/backend/internal/auth"
	"github.com/HunterKs/Parv-Creations/backend/internal/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// UserHandler handles user and role CRUD operations.
type UserHandler struct {
	userColl *mongo.Collection
	roleColl *mongo.Collection
}

// NewUserHandler creates a new UserHandler with the given collections.
func NewUserHandler(userColl, roleColl *mongo.Collection) *UserHandler {
	return &UserHandler{
		userColl: userColl,
		roleColl: roleColl,
	}
}

func parseObjectID(hex string) (bson.ObjectID, bool) {
	objectID, err := bson.ObjectIDFromHex(hex)
	return objectID, err == nil
}

// GetUsers handles GET /users
// Returns a paginated list of users (without password hashes).
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	log.Printf("UserHandler.GetUsers start")

	query := r.URL.Query()
	page := parsePositiveInt(query.Get("page"), 1)
	limit := parsePositiveInt(query.Get("limit"), 10)
	if limit > 100 {
		limit = 100
	}

	filter, ok := buildUserFilter(query.Get("search"), query.Get("role_id"), query.Get("is_active"))
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid user filter")
		return
	}

	sortBy := normalizeUserSortField(query.Get("sort_by"))
	sortOrder := parseSortOrder(query.Get("sort_order"))
	skip := int64((page - 1) * limit)

	total, err := h.userColl.CountDocuments(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: sortBy, Value: sortOrder}})

	cursor, err := h.userColl.Find(r.Context(), filter, findOptions)
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
	if users == nil {
		users = []models.User{}
	}

	// Remove password hashes from the response
	for i := range users {
		users[i].PasswordHash = ""
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"users": users,
			"pagination": map[string]interface{}{
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
			},
		},
	})
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func buildUserFilter(search, roleID, isActive string) (bson.M, bool) {
	filter := bson.M{}

	search = strings.TrimSpace(search)
	if search != "" {
		regex := bson.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"email": regex},
			bson.M{"first_name": regex},
			bson.M{"last_name": regex},
		}
	}

	roleID = strings.TrimSpace(roleID)
	if roleID != "" {
		objectID, ok := parseObjectID(roleID)
		if !ok {
			return nil, false
		}
		filter["role_id"] = objectID
	}

	isActive = strings.TrimSpace(strings.ToLower(isActive))
	if isActive != "" {
		active, ok := parseBoolFilter(isActive)
		if !ok {
			return nil, false
		}
		filter["is_active"] = active
	}

	return filter, true
}

func parseBoolFilter(value string) (bool, bool) {
	switch value {
	case "true", "1", "active":
		return true, true
	case "false", "0", "inactive":
		return false, true
	default:
		return false, false
	}
}

func normalizeUserSortField(sortBy string) string {
	candidate := strings.TrimSpace(sortBy)
	switch candidate {
	case "email", "first_name", "last_name", "role_id", "is_active", "last_login_at", "created_at", "updated_at":
		return candidate
	default:
		return "created_at"
	}
}

func parseSortOrder(value string) int {
	if value == "1" {
		return 1
	}
	return -1
}

// GetUserByID handles GET /users/{id}
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("UserHandler.GetUserByID start")

	vars := mux.Vars(r)
	id := vars["id"]
	objectID, ok := parseObjectID(id)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

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
	log.Printf("UserHandler.CreateUser start")

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
	roleID, ok := parseObjectID(input.RoleID)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

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
	log.Printf("UserHandler.UpdateUser start")

	vars := mux.Vars(r)
	id := vars["id"]
	objectID, ok := parseObjectID(id)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

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
		roleID, ok := parseObjectID(input.RoleID)
		if !ok {
			respondError(w, http.StatusBadRequest, "Invalid role ID")
			return
		}
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
	log.Printf("UserHandler.DeleteUser start")

	vars := mux.Vars(r)
	id := vars["id"]
	objectID, ok := parseObjectID(id)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

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
	log.Printf("UserHandler.GetRoles start")

	query := r.URL.Query()
	page := parsePositiveInt(query.Get("page"), 1)
	limit := parsePositiveInt(query.Get("limit"), 10)
	if limit > 100 {
		limit = 100
	}

	filter := buildRoleFilter(query.Get("search"))
	sortBy := normalizeRoleSortField(query.Get("sort_by"))
	sortOrder := parseSortOrder(query.Get("sort_order"))
	skip := int64((page - 1) * limit)

	total, err := h.roleColl.CountDocuments(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: sortBy, Value: sortOrder}})

	cursor, err := h.roleColl.Find(r.Context(), filter, findOptions)
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
	if roles == nil {
		roles = []models.Role{}
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"roles": roles,
			"pagination": map[string]interface{}{
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
			},
		},
	})
}

func buildRoleFilter(search string) bson.M {
	filter := bson.M{}

	search = strings.TrimSpace(search)
	if search != "" {
		regex := bson.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"name": regex},
			bson.M{"description": regex},
		}
	}

	return filter
}

func normalizeRoleSortField(sortBy string) string {
	candidate := strings.TrimSpace(sortBy)
	switch candidate {
	case "name", "description", "created_at", "updated_at":
		return candidate
	default:
		return "created_at"
	}
}

// GetRoleByID handles GET /roles/{id}
func (h *UserHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("UserHandler.GetRoleByID start")

	vars := mux.Vars(r)
	id := vars["id"]
	objectID, ok := parseObjectID(id)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

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
	log.Printf("UserHandler.CreateRole start")

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
	var permissions []models.PermissionKey
	for _, p := range input.Permissions {
		permissions = append(permissions, models.PermissionKey(p))
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
	log.Printf("UserHandler.UpdateRole start")

	vars := mux.Vars(r)
	id := vars["id"]
	objectID, ok := parseObjectID(id)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

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
		var permissions []models.PermissionKey
		for _, p := range input.Permissions {
			permissions = append(permissions, models.PermissionKey(p))
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
	log.Printf("UserHandler.DeleteRole start")

	vars := mux.Vars(r)
	id := vars["id"]
	objectID, ok := parseObjectID(id)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

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
