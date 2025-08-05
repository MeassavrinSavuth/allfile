package routes

import (
	"social-sync-backend/controllers"
	"social-sync-backend/middleware"

	"github.com/gorilla/mux"
)

func RegisterTaskRoutes(router *mux.Router) {
	tasks := router.PathPrefix("/api/workspaces/{workspaceId}/tasks").Subrouter()
	tasks.Use(middleware.JWTMiddleware)

	tasks.HandleFunc("", controllers.ListTasks).Methods("GET")
	tasks.HandleFunc("", controllers.CreateTask).Methods("POST")
	tasks.HandleFunc("/{taskId}", controllers.UpdateTask).Methods("PATCH")
	tasks.HandleFunc("/{taskId}", controllers.DeleteTask).Methods("DELETE")

	// Comments
	comments := tasks.PathPrefix("/{taskId}/comments").Subrouter()
	comments.HandleFunc("", controllers.ListComments).Methods("GET")
	comments.HandleFunc("", controllers.AddComment).Methods("POST")
	comments.HandleFunc("/{commentId}", controllers.DeleteComment).Methods("DELETE")

	// Reactions
	reactions := tasks.PathPrefix("/{taskId}/reactions").Subrouter()
	reactions.HandleFunc("", controllers.GetTaskReactions).Methods("GET")
	reactions.HandleFunc("", controllers.ToggleReaction).Methods("POST")
	reactions.HandleFunc("/user", controllers.GetUserReactions).Methods("GET")
}
