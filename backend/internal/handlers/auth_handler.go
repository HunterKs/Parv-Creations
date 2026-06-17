package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/HunterKs/Parv-Creations/backend/internal/auth"
	"github.com/HunterKs/Parv-Creations/backend/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	userColl          *mongo.Collection
	roleColl          *mongo.Collection
	rememberMeColl    *mongo.Collection
	sessionCookieName string
}

// NewAuthHandler creates a new AuthHandler with the given collections.
func NewAuthHandler(userColl, roleColl, rememberMeColl *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		userColl:          userColl,
		roleColl:          roleColl,
		rememberMeColl:    rememberMeColl,
		sessionCookieName: "session_token",
	}
}

// Login handles the login request.
// Expected JSON body: { "email": "...", "password": "...", "remember": true/false }
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("AuthHandler.Login start")

	var creds models.Credentials
	if err := decodeJSON(r, &creds); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Find the user by email
	var user models.User
	err := h.userColl.FindOne(r.Context(), bson.M{"email": creds.Email}).Decode(&user)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Check password
	if !auth.CheckPasswordHash(creds.Password, user.PasswordHash) {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	var role models.Role
	if err := h.roleColl.FindOne(r.Context(), bson.M{"_id": user.RoleID}).Decode(&role); err != nil {
		respondError(w, http.StatusForbidden, "Assigned role not found")
		return
	}
	user.Role = &role

	// Generate session token (JWT)
	sessionToken, err := auth.GenerateJWT(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate session token")
		return
	}

	// Set session cookie
	setSessionCookie(w, h.sessionCookieName, sessionToken, auth.TokenDuration)

	// If remember me is requested, generate a remember-me token
	if creds.Remember {
		plainToken, tokenHash, err := auth.GenerateRememberMeToken()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to generate remember-me token")
			return
		}
		// Store the remember-me token hash in the database
		rememberMeToken := models.RememberMeToken{
			UserID:    user.ID,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().Add(auth.RememberMeTokenDuration),
			CreatedAt: time.Now(),
		}
		_, err = h.rememberMeColl.InsertOne(r.Context(), rememberMeToken)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to store remember-me token")
			return
		}
		// Set remember-me cookie (HttpOnly, secure in production)
		http.SetCookie(w, &http.Cookie{
			Name:     "remember_me_token",
			Value:    plainToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			MaxAge:   int(auth.RememberMeTokenDuration.Seconds()),
		})
	}

	// Update last login time
	_, err = h.userColl.UpdateOne(
		r.Context(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"last_login_at": time.Now(), "updated_at": time.Now()}},
	)
	if err != nil {
		// Log error but don't fail the login
	}

	// Respond with user info (without password hash)
	user.PasswordHash = ""
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"token":   sessionToken,
		"user":    user,
	})
}

// Logout handles the logout request.
// Clears the session cookie and removes the remember-me token if present.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Println("AuthHandler.Logout start")

	if claims, ok := auth.GetClaimsFromContext(r); ok {
		if userID, err := bson.ObjectIDFromHex(claims.UserID); err == nil {
			if _, err := h.rememberMeColl.DeleteMany(r.Context(), bson.M{"user_id": userID}); err != nil {
				log.Printf("AuthHandler.Logout remember-me cleanup failed: %v", err)
			}
		}
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     h.sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Clear remember-me cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "remember_me_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Logged out successfully"}`))
}

// Status confirms that the current request has a valid authenticated session.
func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	log.Printf("AuthHandler.Status start")

	claims, ok := auth.GetClaimsFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": true,
		"email":         claims.Email,
	})
}

// Me returns the current authenticated user without password material.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	log.Printf("AuthHandler.Me start")

	claims, ok := auth.GetClaimsFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := bson.ObjectIDFromHex(claims.UserID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	if err := h.userColl.FindOne(r.Context(), bson.M{"_id": userID}).Decode(&user); err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	user.PasswordHash = ""
	respondJSON(w, http.StatusOK, user)
}

// decodeJSON is a helper to decode JSON request body into a struct.
func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// respondJSON is a helper to send a JSON response.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// respondError is a helper to send an error JSON response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// setSessionCookie sets an HttpOnly cookie with the given token and duration.
func setSessionCookie(w http.ResponseWriter, name, token string, duration time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		MaxAge:   int(duration.Seconds()),
	})
}
