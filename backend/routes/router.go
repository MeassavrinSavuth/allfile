package routes

import "github.com/gorilla/mux"

func InitRoutes() *mux.Router {
	r := mux.NewRouter()

	AuthRoutes(r)
	RegisterUserRoutes(r)
	RegisterWorkspaceRoutes(r)
	RegisterInvitationRoutes(r)
	RegisterTaskRoutes(r)
	RegisterDraftPostRoutes(r)
	RegisterMediaRoutes(r)
	// Add more like RegisterPostRoutes(r), etc.

	return r
}
