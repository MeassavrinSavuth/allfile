// lib/database.go
package lib

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func ConnectDB() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	log.Printf("INFO: DATABASE_URL being used: %s", dbURL)

	var err error
	DB, err = sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Error connecting to DB:", err)
	}

	// Configure connection pool to avoid stale prepared statements error
	DB.SetMaxOpenConns(10)                 // max open connections (adjust as needed)
	DB.SetMaxIdleConns(5)                  // max idle connections
	DB.SetConnMaxLifetime(1 * time.Minute) // recycle connections every 5 minutes

	err = DB.PingContext(context.Background())
	if err != nil {
		log.Fatal("Error pinging DB:", err)
	}

}

func GetDB() *sql.DB {
	return DB
}
