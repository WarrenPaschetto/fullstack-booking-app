package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func ConnectDB() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	db, err := sql.Open("libsql", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Turso database: %w", err)
	}

	// test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("the Turso database not reachable: %w", err)
	}

	return db, nil
}
