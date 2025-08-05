package main

import (
	"log"
	"net/http"
	"os"

	"social-sync-backend/lib"
	"social-sync-backend/routes"
	"social-sync-backend/utils"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
)

// CORSMiddleware sets CORS headers.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("❌ Error loading .env file: %v", err)
	}

	// Initialize database
	lib.ConnectDB()
	defer func() {
		if lib.DB != nil {
			if err := lib.DB.Close(); err != nil {
				log.Printf("❌ Error closing database: %v", err)
			} else {
				log.Println("✅ Database connection closed.")
			}
		}
	}()
	log.Println("✅ Connected to PostgreSQL DB!")

	// Initialize Cloudinary
	if err := lib.InitCloudinary(); err != nil {
		log.Fatalf("❌ Failed to initialize Cloudinary: %v", err)
	}
	log.Println("✅ Cloudinary initialized!")

	// Setup cron job for social account sync
	c := cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger),
		cron.DelayIfStillRunning(cron.DefaultLogger),
	))
	if _, err := c.AddFunc("@every 24h", func() {
		log.Println("🔁 Running scheduled social account sync...")
		utils.SyncAllSocialAccountsTask(lib.DB)
	}); err != nil {
		log.Fatalf("❌ Failed to schedule cron: %v", err)
	}
	c.Start()
	defer c.Stop()
	log.Println("✅ Cron job started (every 12h).")

	// Setup routes and middleware
	r := routes.InitRoutes()
	handler := CORSMiddleware(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running at: http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
