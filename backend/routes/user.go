package routes

import (
	"net/http"
	"social-sync-backend/controllers"
	"social-sync-backend/middleware"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router) {
	// Dashboard route
	r.Handle("/api/dashboard",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.DashboardHandler)),
	).Methods("GET")

	// Profile CRUD route
	r.Handle("/api/profile",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.ProfileHandler)),
	).Methods("GET", "PUT", "DELETE", "OPTIONS")

	// Upload profile image
	r.Handle("/api/profile/image",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.ProfileImageHandler)),
	).Methods("POST", "OPTIONS")

	// Serve profile image publicly (GET doesn't need auth middleware here unless required)
	r.Handle("/api/profile/image/{userID}",
		http.HandlerFunc(controllers.ProfileImageHandler),
	).Methods("GET")

	r.Handle("/api/profile/password", 
		middleware.JWTMiddleware(http.HandlerFunc(controllers.ProfilePasswordHandler))).Methods("PUT", "OPTIONS")

	r.Handle("/api/upload", middleware.EnableCORS(middleware.JWTMiddleware(http.HandlerFunc(controllers.UploadImageHandler)))).Methods("POST", "OPTIONS")


}
