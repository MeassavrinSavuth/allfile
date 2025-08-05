package routes

import (
	"social-sync-backend/controllers"
	"social-sync-backend/middleware"

	"github.com/gorilla/mux"
)

func RegisterMediaRoutes(r *mux.Router) {
	media := r.PathPrefix("/api/workspaces/{workspaceId}/media").Subrouter()
	media.Use(middleware.JWTMiddleware)

	media.HandleFunc("", controllers.UploadMedia).Methods("POST")
	media.HandleFunc("", controllers.ListMedia).Methods("GET")
	media.HandleFunc("/{mediaId}", controllers.DeleteMedia).Methods("DELETE")
	media.HandleFunc("/{mediaId}/tags", controllers.UpdateMediaTags).Methods("PATCH")
}
