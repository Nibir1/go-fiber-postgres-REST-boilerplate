package api

import (
	"fmt"

	"github.com/go-playground/validator/v10"        // For custom request validation
	"github.com/gofiber/fiber/v2"                   // Fiber web framework
	"github.com/gofiber/fiber/v2/middleware/cors"   // ✅ CORS middleware
	"github.com/gofiber/fiber/v2/middleware/logger" // Request logging middleware

	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"

	"github.com/gofiber/swagger"                                  // swagger handler
	_ "github.com/nibir1/go-fiber-postgres-REST-boilerplate/docs" // generated docs
)

// ---------------------------
// Swagger Security Definition
// ---------------------------

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

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
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	// Create a new Fiber app
	app := fiber.New(fiber.Config{})

	// ---------------------------
	// Global Middlewares
	// ---------------------------

	// Request logger (prints each request)
	app.Use(logger.New())

	// ✅ Enable CORS for frontend communication
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.AllowedOrigins, // ✅ loaded from env/config
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length, Content-Type",
		AllowCredentials: true, // ✅ now safe, no wildcard
	}))

	// Initialize validator for request validation
	validate := validator.New()

	// Register custom validation for currency fields
	validate.RegisterValidation("currency", validCurrency)

	// Initialize server instance
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

	// Swagger UI route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// ---------------------
	// PUBLIC ROUTES
	// ---------------------
	app.Post("/users", server.createUser)
	app.Post("/users/login", server.loginUser)

	// ---------------------
	// PROTECTED ROUTES
	// ---------------------
	auth := app.Group("/", authMiddlewareFiber(server.tokenMaker))

	// Account-related endpoints
	auth.Post("/accounts", server.createAccount)
	auth.Get("/accounts/:id", server.getAccount)
	auth.Get("/accounts", server.listAccount)
	auth.Delete("/accounts/:id", server.deleteAccount)

	// Transfer-related endpoint
	auth.Post("/transfers", server.createTransfer)
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
