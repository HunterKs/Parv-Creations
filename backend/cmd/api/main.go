package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/HunterKs/Parv-Creations/backend/internal/auth"
	"github.com/HunterKs/Parv-Creations/backend/internal/database"
	"github.com/HunterKs/Parv-Creations/backend/internal/handlers"
	appMiddleware "github.com/HunterKs/Parv-Creations/backend/internal/middleware"
)

func main() {
	if err := appMiddleware.InitDailyLogger(); err != nil {
		log.Fatalf("Failed to initialize daily logger: %v", err)
	}

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
	permissionColl := db.Collection("permissions")

	database.SeedData(userColl, roleColl, permissionColl)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userColl, roleColl, rememberMeColl)
	userHandler := handlers.NewUserHandler(userColl, roleColl)
	permissionHandler := handlers.NewPermissionHandler(permissionColl)

	// Create a new router
	r := mux.NewRouter()
	r.Use(appMiddleware.LoggerMiddleware)

	// Health check endpoint (public)
	r.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","database":"connected","project":"Parv Creations Engine"}`))
	}).Methods("GET")

	r.PathPrefix("/admin/").Handler(http.StripPrefix("/admin/", http.FileServer(http.Dir("public/admin"))))

	// Public auth endpoints
	r.HandleFunc("/api/v1/admin/auth/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/api/v1/admin/auth/logout", authHandler.Logout).Methods("POST", "OPTIONS")

	// Admin routes (require authentication)
	admin := r.PathPrefix("/api/v1/admin").Subrouter()
	// Apply authentication middleware to all admin routes
	admin.Use(auth.AuthMiddleware)

	// Auth endpoints
	admin.HandleFunc("/auth/status", authHandler.Status).Methods("GET")
	admin.HandleFunc("/auth/me", authHandler.Me).Methods("GET")

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

	// Permission management endpoints
	admin.HandleFunc("/permissions", permissionHandler.ListPermissions).Methods("GET")
	admin.HandleFunc("/permissions", permissionHandler.CreatePermission).Methods("POST")
	admin.HandleFunc("/permissions/{id}", permissionHandler.UpdatePermission).Methods("PUT")
	admin.HandleFunc("/permissions/{id}", permissionHandler.DeletePermission).Methods("DELETE")

	// Unified API permission endpoints for the static admin frontend.
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(auth.AuthMiddleware)
	api.HandleFunc("/permissions", permissionHandler.ListPermissions).Methods("GET")
	api.HandleFunc("/permissions", permissionHandler.CreatePermission).Methods("POST")
	api.HandleFunc("/permissions/{id}", permissionHandler.UpdatePermission).Methods("PUT")
	api.HandleFunc("/permissions/{id}", permissionHandler.DeletePermission).Methods("DELETE")

	r.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/client/about.html")
	}).Methods("GET")
	r.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/client/contact.html")
	}).Methods("GET")
	r.HandleFunc("/privacy-policy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/client/privacy.html")
	}).Methods("GET")
	r.HandleFunc("/terms-of-service", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/client/terms.html")
	}).Methods("GET")

	// Mount static asset route for public visitor view
	clientFileServer := http.FileServer(http.Dir("public/client"))
	r.PathPrefix("/").Handler(http.StripPrefix("/", clientFileServer))

	// Set up HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Parv Creations Core Engine cleanly operational on port %s...", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(r)); err != nil {
		log.Fatalf("Server bootstrap failed: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" && r.URL.Path != "/api/v1/admin/auth/logout" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
