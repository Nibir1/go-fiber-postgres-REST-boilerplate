package api

import (
	"fmt"

	"github.com/go-playground/validator/v10"        // For custom request validation
	"github.com/gofiber/fiber/v2"                   // Fiber web framework
	"github.com/gofiber/fiber/v2/middleware/logger" // Request logging middleware
	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
)

// ---------------------------
// Server Struct
// ---------------------------

// Server holds all dependencies and configuration for the API server
type Server struct {
	config     util.Config         // Application configuration (DB, token keys, etc.)
	store      db.Store            // Database access layer (SQLC queries)
	tokenMaker token.Maker         // Token maker for JWT/Paseto
	app        *fiber.App          // Fiber app instance (routes + middleware)
	validate   *validator.Validate // Validator for custom request validations
}

// ---------------------------
// NewServer
// ---------------------------

// NewServer initializes a new API server with Fiber framework
func NewServer(config util.Config, store db.Store) (*Server, error) {
	// Initialize token maker with symmetric key from config
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey) // token.NewJWTMaker(config.TokenSymmetricKey) --- Is another alternative ---
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		// You can add custom Fiber configurations here (timeouts, error handler, etc.)
	})

	// Add a logger middleware that prints request info to console
	app.Use(logger.New())

	// Initialize validator for request validation
	validate := validator.New()

	// Register a custom validation for currency fields
	validate.RegisterValidation("currency", validCurrency)

	// Initialize server instance with all dependencies
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		app:        app,
		validate:   validate,
	}

	// Setup all API routes (public and protected)
	server.setUpRoutes()

	return server, nil
}

// ---------------------------
// Route Setup
// ---------------------------

// setUpRoutes defines all public and protected routes for the API server
func (server *Server) setUpRoutes() {
	app := server.app

	// ---------------------
	// PUBLIC ROUTES
	// ---------------------
	// No authentication required
	app.Post("/users", server.createUser)      // Create new user
	app.Post("/users/login", server.loginUser) // Login user and get access token

	// ---------------------
	// PROTECTED ROUTES
	// ---------------------
	// Group of routes that require valid JWT/Paseto token
	auth := app.Group("/", authMiddlewareFiber(server.tokenMaker)) // Use our Fiber auth middleware

	// Account-related endpoints
	auth.Post("/accounts", server.createAccount)       // Create a new bank account
	auth.Get("/accounts/:id", server.getAccount)       // Get account details by ID
	auth.Get("/accounts", server.listAccount)          // List all accounts
	auth.Delete("/accounts/:id", server.deleteAccount) // Delete account by ID

	// Transfer-related endpoint
	auth.Post("/transfers", server.createTransfer) // Make a transfer between accounts
}

// ---------------------------
// Start Server
// ---------------------------

// Start runs the Fiber server on the given address (e.g., ":8080")
func (server *Server) Start(address string) error {
	return server.app.Listen(address)
}

// ---------------------------
// Helper: Error Response
// ---------------------------

// errorResponse standardizes JSON error responses for Fiber
func errorResponse(err error) fiber.Map {
	return fiber.Map{
		"error": err.Error(),
	}
}
