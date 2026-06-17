package handlers

import (
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/HunterKs/Parv-Creations/backend/internal/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// PermissionHandler handles permission CRUD operations.
type PermissionHandler struct {
	permissionColl *mongo.Collection
}

// NewPermissionHandler creates a new PermissionHandler.
func NewPermissionHandler(permissionColl *mongo.Collection) *PermissionHandler {
	return &PermissionHandler{permissionColl: permissionColl}
}

// ListPermissions handles GET /permissions.
func (h *PermissionHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	log.Printf("PermissionHandler.ListPermissions start")

	query := r.URL.Query()
	page := parsePositiveInt(query.Get("page"), 1)
	limit := parsePositiveInt(query.Get("limit"), 10)
	if limit > 100 {
		limit = 100
	}

	filter := buildPermissionFilter(query.Get("search"))
	sortBy := normalizePermissionSortField(query.Get("sort_by"))
	sortOrder := parseSortOrder(query.Get("sort_order"))
	skip := int64((page - 1) * limit)

	total, err := h.permissionColl.CountDocuments(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: sortBy, Value: sortOrder}})

	cursor, err := h.permissionColl.Find(r.Context(), filter, findOptions)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer cursor.Close(r.Context())

	var permissions []models.Permission
	if err := cursor.All(r.Context(), &permissions); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if permissions == nil {
		permissions = []models.Permission{}
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"permissions": permissions,
			"pagination": map[string]interface{}{
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
			},
		},
	})
}

func buildPermissionFilter(search string) bson.M {
	filter := bson.M{}

	search = strings.TrimSpace(search)
	if search != "" {
		regex := bson.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"key": regex},
			bson.M{"name": regex},
			bson.M{"description": regex},
		}
	}

	return filter
}

func normalizePermissionSortField(sortBy string) string {
	candidate := strings.TrimSpace(sortBy)
	switch candidate {
	case "key", "name", "description", "created_at", "updated_at":
		return candidate
	default:
		return "created_at"
	}
}

// CreatePermission handles POST /permissions.
func (h *PermissionHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	log.Printf("PermissionHandler.CreatePermission start")

	var input struct {
		Key         string `json:"key" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.Key == "" || input.Name == "" {
		respondError(w, http.StatusBadRequest, "Permission key and name are required")
		return
	}

	var existing models.Permission
	err := h.permissionColl.FindOne(r.Context(), bson.M{"key": input.Key}).Decode(&existing)
	if err == nil {
		respondError(w, http.StatusConflict, "Permission already exists")
		return
	}
	if err != mongo.ErrNoDocuments {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now()
	permission := models.Permission{
		Key:         models.PermissionKey(input.Key),
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := h.permissionColl.InsertOne(r.Context(), permission)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if id, ok := result.InsertedID.(bson.ObjectID); ok {
		permission.ID = id
	}

	respondJSON(w, http.StatusCreated, permission)
}

// UpdatePermission handles PUT /permissions/{id}.
func (h *PermissionHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	log.Printf("PermissionHandler.UpdatePermission start")

	objectID, ok := parseObjectID(mux.Vars(r)["id"])
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	var input struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	update := bson.M{}
	if input.Key != "" {
		update["key"] = models.PermissionKey(input.Key)
	}
	if input.Name != "" {
		update["name"] = input.Name
	}
	if input.Description != "" {
		update["description"] = input.Description
	}
	update["updated_at"] = time.Now()

	if len(update) == 1 {
		respondError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	result, err := h.permissionColl.UpdateOne(r.Context(), bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result.MatchedCount == 0 {
		respondError(w, http.StatusNotFound, "Permission not found")
		return
	}

	var permission models.Permission
	if err := h.permissionColl.FindOne(r.Context(), bson.M{"_id": objectID}).Decode(&permission); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, permission)
}

// DeletePermission handles DELETE /permissions/{id}.
func (h *PermissionHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	log.Printf("PermissionHandler.DeletePermission start")

	objectID, ok := parseObjectID(mux.Vars(r)["id"])
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	result, err := h.permissionColl.DeleteOne(r.Context(), bson.M{"_id": objectID})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result.DeletedCount == 0 {
		respondError(w, http.StatusNotFound, "Permission not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Permission deleted"})
}
