package main

import (
	"database/sql" // For database connectivity
	"log"          // For logging errors

	// Database driver for PostgreSQL
	_ "github.com/lib/pq"

	// Import our own packages
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/api"        // API layer
	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc" // SQLC-generated database queries
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"       // Utilities for config, password hashing, etc.
)

func main() {
	// Load configuration from current directory
	config, err := util.LoadConfig(".")
	if err != nil {
		// If config fails to load, terminate program with error
		log.Fatal("cannot load configuration:", err)
	}

	// Open a connection to the database
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to db:", err)
	}

	// Create a store (wrapper around SQLC queries)
	store := db.NewStore(conn)

	// Initialize the API server with configuration and store
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	// Start listening on the configured server address
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}
