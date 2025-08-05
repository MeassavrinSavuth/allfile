package routes

import (
	"social-sync-backend/controllers"
	"social-sync-backend/middleware"

	"github.com/gorilla/mux"
)

func RegisterDraftPostRoutes(r *mux.Router) {
	drafts := r.PathPrefix("/api/workspaces/{workspaceId}/drafts").Subrouter()
	drafts.Use(middleware.JWTMiddleware)
	drafts.HandleFunc("", controllers.ListDraftPosts).Methods("GET")
	drafts.HandleFunc("", controllers.CreateDraftPost).Methods("POST")
	drafts.HandleFunc("/{draftId}", controllers.UpdateDraftPost).Methods("PATCH")
	drafts.HandleFunc("/{draftId}", controllers.DeleteDraftPost).Methods("DELETE")
	drafts.HandleFunc("/{draftId}/publish", controllers.PublishDraftPost).Methods("POST")
}
