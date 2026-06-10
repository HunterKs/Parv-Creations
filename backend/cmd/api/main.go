package main

import (
	"log"
	"net/http"
	"os"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/HunterKs/Parv-Creations/backend/internal/auth"
	"github.com/HunterKs/Parv-Creations/backend/internal/database"
	"github.com/HunterKs/Parv-Creations/backend/internal/handlers"
)

func main() {
	// Initialize the MongoDB Atlas pool and get the client and database
	client, db, err := database.ConnectProductionCluster()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(nil)

	// Get collections
	userColl := db.Collection("users")
	roleColl := db.Collection("roles")
	rememberMeColl := db.Collection("remember_me_tokens")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userColl, roleColl, rememberMeColl)
	userHandler := handlers.NewUserHandler(userColl, roleColl)

	// Create a new router
	r := mux.NewRouter()

	// Health check endpoint (public)
	r.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","database":"connected","project":"Parv Creations Engine"}`))
	}).Methods("GET")

	r.PathPrefix("/admin/").Handler(http.StripPrefix("/admin/", http.FileServer(http.Dir("public/admin"))))

	// Admin routes (require authentication)
	admin := r.PathPrefix("/api/v1/admin").Subrouter()
	// Apply authentication middleware to all admin routes
	admin.Use(auth.AuthMiddleware)

	// Auth endpoints (login/logout)
	admin.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	admin.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")

	// User management endpoints
	admin.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	admin.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	admin.HandleFunc("/users/{id}", userHandler.GetUserByID).Methods("GET")
	admin.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	admin.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Role management endpoints
	admin.HandleFunc("/roles", userHandler.GetRoles).Methods("GET")
	admin.HandleFunc("/roles", userHandler.CreateRole).Methods("POST")
	admin.HandleFunc("/roles/{id}", userHandler.GetRoleByID).Methods("GET")
	admin.HandleFunc("/roles/{id}", userHandler.UpdateRole).Methods("PUT")
	admin.HandleFunc("/roles/{id}", userHandler.DeleteRole).Methods("DELETE")

	// Wrap the router with CORS and logging middleware (optional but useful for development)
	// In production, you might want to adjust CORS settings.
	headersOk := gorillaHandlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := gorillaHandlers.AllowedOrigins([]string{"*"}) // Be more restrictive in production
	methodsOk := gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	// Set up HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Parv Creations Core Engine cleanly operational on port %s...", port)
	if err := http.ListenAndServe(":"+port, gorillaHandlers.CORS(headersOk, originsOk, methodsOk)(r)); err != nil {
		log.Fatalf("Server bootstrap failed: %v", err)
	}
}
