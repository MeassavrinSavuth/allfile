package routes

import (
	"net/http"
	"social-sync-backend/controllers"
	"social-sync-backend/middleware"

	"github.com/gorilla/mux"
)

func RegisterWorkspaceRoutes(r *mux.Router) {
	r.Handle("/api/workspaces",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.ListWorkspaces))).Methods("GET")
	r.Handle("/api/workspaces",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.CreateWorkspace))).Methods("POST")
	r.Handle("/api/workspaces/{workspaceId}/members",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.ListWorkspaceMembers))).Methods("GET")
	r.Handle("/api/workspaces/{workspaceId}/leave",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.LeaveWorkspace))).Methods("POST")
	r.Handle("/api/workspaces/{workspaceId}/members/{memberId}",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.RemoveWorkspaceMember))).Methods("DELETE")
	r.Handle("/api/workspaces/{workspaceId}",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.DeleteWorkspace))).Methods("DELETE")
	r.Handle("/api/workspaces/{workspaceId}/members/{memberId}/role",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.ChangeMemberRole))).Methods("PATCH")
	r.HandleFunc("/ws/{workspaceId}", controllers.WorkspaceWSHandler).Methods("GET")
}
