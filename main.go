// @title Go-Fiber-Postgres-REST-Boilerplate
// @version 1.0
// @description A lightweight boilerplate for building RESTful APIs with Golang (Fiber) and PostgreSQL.
// @description This project provides a clean and modular backend setup with Docker support for easy local development.
// @description Designed as a starting point for rapid prototyping or learning, without cloud deployment overhead.
// @description
// @description **Developer:** Nahasat Nibir (Software Developer)
// @description **LinkedIn:** https://www.linkedin.com/in/nibir-1/
// @description **Portfolio:** https://github.com/nibir1 | https://www.artstation.com/nibir
// @contact.name Nahasat Nibir
// @contact.email nahasat.nibir@gmail.com
// @host 127.0.0.1:8080
// @BasePath /
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

// ---------------------------
// Swagger Security Definition
// ---------------------------

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

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
