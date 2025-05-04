package db

import (
	"os"
	"testing"
)

func TestConnectDB(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL is not set; skipping database test")
	}

	db, err := ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()
}
