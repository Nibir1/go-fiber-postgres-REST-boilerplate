package api

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"                                               // For Postgres-specific error handling
	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc" // SQLC database package
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"       // Utility functions (e.g., password hashing)
)

// ---------------------------
// Request and Response Structs
// ---------------------------

// createUserRequest represents the expected JSON body for creating a new user
// @Description Create user request payload
type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"` // Alphanumeric username, required
	Password string `json:"password" binding:"required,min=6"`    // Password, min 6 chars
	FullName string `json:"full_name" binding:"required"`         // Full name, required
	Email    string `json:"email" binding:"required,email"`       // Email, required, must be valid
}

// userResponse represents the JSON response returned for a user
// @Description User response payload
type userResponse struct {
	Username          string    `json:"username"`            // Username of user
	FullName          string    `json:"full_name"`           // Full name of user
	Email             string    `json:"email"`               // Email of user
	PasswordChangedAt time.Time `json:"password_changed_at"` // Timestamp of last password change
	CreatedAt         time.Time `json:"created_at"`          // Timestamp of user creation
}

// newUserResponse converts db.User struct into a userResponse for API response
func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

// loginUserRequest represents the expected JSON body for logging in
// @Description Login request payload
type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

// loginUserResponse represents the JSON response for login
// @Description Login response payload
type loginUserResponse struct {
	AccessToken string       `json:"access_token"` // JWT/Paseto token
	User        userResponse `json:"user"`         // User details
}

// ---------------------------
// Handlers
// ---------------------------

// createUser handles POST /users endpoint for creating a new user

// CreateUser godoc
// @Summary      Register a new user
// @Description  Creates a new user with username, password, full name, and email
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body      createUserRequest  true  "User info"
// @Success      200   {object}  userResponse
// @Router       /users [post]
func (server *Server) createUser(c *fiber.Ctx) error {
	// 1. Parse JSON request body into createUserRequest struct
	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		// Invalid JSON → return 400 Bad Request
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// Manual validation
	if req.Username == "" || req.Password == "" || req.FullName == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(errors.New("missing required fields")))
	}

	// 2. Hash password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		// Error during hashing → return 500 Internal Server Error
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 3. Prepare database arguments
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	// 4. Create user in the database
	user, err := server.store.CreateUser(c.Context(), arg)
	if err != nil {
		// Handle Postgres-specific errors (unique constraint violation)
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "unique_violation" {
				return c.Status(fiber.StatusForbidden).JSON(errorResponse(err))
			}
		}
		// Other errors → 500 Internal Server Error
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 5. Build response
	resp := newUserResponse(user)

	// 6. Return 200 OK with user data
	return c.Status(fiber.StatusOK).JSON(resp)
}

// loginUser handles POST /users/login endpoint

// LoginUser godoc
// @Summary      Log in a user
// @Description  Authenticates user credentials and returns a JWT access token
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        credentials  body      loginUserRequest  true  "Login credentials"
// @Success      200  {object}  loginUserResponse
// @Router       /users/login [post]
func (server *Server) loginUser(c *fiber.Ctx) error {
	// 1. Parse request body into loginUserRequest
	var req loginUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// Manual validation
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(errors.New("missing required fields")))
	}

	// 2. Retrieve user from DB by username
	user, err := server.store.GetUser(c.Context(), req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User not found → 404
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		// Other DB error → 500
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 3. Verify password
	if err := util.CheckPassword(req.Password, user.HashedPassword); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
	}

	// 4. Create access token
	accessToken, _, err := server.tokenMaker.CreateToken(user.Username, server.config.AccessTokenDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 5. Build response
	resp := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}

	// 6. Return 200 OK with access token and user info
	return c.Status(fiber.StatusOK).JSON(resp)
}
