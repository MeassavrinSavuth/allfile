package routes

import (
	"net/http"
	"social-sync-backend/controllers"
	"social-sync-backend/middleware"

	"github.com/gorilla/mux"
)

func RegisterInvitationRoutes(r *mux.Router) {
	// Get user's pending invitations
	r.Handle("/api/invitations",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.GetInvitations))).Methods("GET")

	// Send invitation to join workspace
	r.Handle("/api/workspaces/{workspaceId}/invite",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.SendInvitation))).Methods("POST")

	// Accept invitation
	r.Handle("/api/invitations/{invitationId}/accept",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.AcceptInvitation))).Methods("POST")

	// Decline invitation
	r.Handle("/api/invitations/{invitationId}/decline",
		middleware.JWTMiddleware(http.HandlerFunc(controllers.DeclineInvitation))).Methods("POST")

	// WebSocket for real-time invitations
	r.HandleFunc("/ws/invitations/{email}", controllers.InvitationWSHandler).Methods("GET")
}
