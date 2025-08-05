package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"social-sync-backend/lib"
)

// type contextKey string

// const UserIDKey = contextKey("userID")

func EnableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		authHeader := r.Header.Get("Authorization")
		fmt.Println("[JWT DEBUG] Authorization header:", authHeader)
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
			fmt.Println("[JWT DEBUG] Extracted token:", token)
		} else if t := r.URL.Query().Get("token"); t != "" {
			token = t
			fmt.Println("[JWT DEBUG] Token from query param:", token)
		} else {
			fmt.Println("[JWT ERROR] No token found in header or query param")
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		claims, err := lib.VerifyToken(token, os.Getenv("JWT_SECRET"))
		if err != nil {
			fmt.Println("[JWT ERROR] Invalid token:", err)
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			fmt.Println("[JWT ERROR] Invalid user ID in token")
			http.Error(w, "Unauthorized: invalid user ID", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
