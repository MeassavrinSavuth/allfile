package controllers

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type userWSClient struct {
	conn *websocket.Conn
}

type userWSHub struct {
	clients map[string]map[*userWSClient]bool // email -> clients
	lock    sync.RWMutex
}

var userHub = &userWSHub{
	clients: make(map[string]map[*userWSClient]bool),
}

func (h *userWSHub) addClient(email string, client *userWSClient) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.clients[email] == nil {
		h.clients[email] = make(map[*userWSClient]bool)
	}
	h.clients[email][client] = true
}

func (h *userWSHub) removeClient(email string, client *userWSClient) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.clients[email] != nil {
		delete(h.clients[email], client)
		if len(h.clients[email]) == 0 {
			delete(h.clients, email)
		}
	}
}

func (h *userWSHub) broadcast(email string, msg []byte) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	for client := range h.clients[email] {
		client.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

// InvitationWSHandler handles per-user invitation WebSocket connections
func InvitationWSHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &userWSClient{conn: conn}
	userHub.addClient(email, client)
	defer func() {
		userHub.removeClient(email, client)
		conn.Close()
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// This server is broadcast-only; ignore client messages
	}
}
