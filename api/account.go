package api

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2" // Import for handling PostgreSQL-specific errors
	"github.com/lib/pq"
	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc" // SQLC database package
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"      // Token handling (JWT/Paseto)
)

// ---------------------------
// Request Structs
// ---------------------------

// **createAccountRequest** defines the structure for incoming JSON data in the `createAccount` handler.
type createAccountRequest struct {
	Currency string `json:"currency" validate:"required,currency"` // This field is required and must be a valid currency code.
}

// **getAccountRequest** defines the structure for path parameters in the `getAccount` handler.
type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"` // This field is required and must be a positive integer (minimum value of 1).
}

// listAccountRequest represents query parameters for listing accounts
type listAccountRequest struct {
	PageID   int `query:"page_id" validate:"required,min=1"`          // Current page number, min 1
	PageSize int `query:"page_size" validate:"required,min=5,max=10"` // Items per page, 5–10
}

// **deleteAccountRequest** defines the structure for path parameters in the `deleteAccount` handler.
type deleteAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"` // This field is required and must be a positive integer (minimum value of 1).
}

// ---------------------------
// Handlers
// ---------------------------

// ------------ **API Functionality for Creating Accounts** ------------

// **(server *Server) createAccount** POST. Handles creating a new account.
func (server *Server) createAccount(c *fiber.Ctx) error {
	// 1. Parse Request Body
	var req createAccountRequest

	// Attempt to parse the request body to the `createAccountRequest` struct.
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	if err := server.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// 2. Extract Authentication Information
	// Retrieve the authentication payload from the context.
	// This assumes the context has a key named `authorizationPayloadKey` containing the token payload.
	payload, ok := c.Locals(authorizationPayloadKey).(*token.Payload)
	if !ok || payload == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(fmt.Errorf("unauthorized")))
	}
	// Get username from payload
	username := payload.Username

	// 3. Prepare Database Arguments
	// Create a `db.CreateAccountParams` struct to hold arguments for the database call.
	arg := db.CreateAccountParams{
		Owner:    username,     // Set the owner based on the authenticated user.
		Currency: req.Currency, // Set the currency from the request.
		Balance:  0,            // Initial balance is set to 0.
	}

	// 4. Call Database Function
	// Call the `CreateAccount` function from the `server.store` object (assuming it interacts with the database).
	// This function likely creates a new account in the database based on the provided arguments.
	account, err := server.store.CreateAccount(c.Context(), arg)

	// Handle Database Errors
	if err != nil {
		// Check if the error is a specific type (`*pq.Error`) indicating a PostgreSQL error.
		if pqErr, ok := err.(*pq.Error); ok {
			// If it's a PostgreSQL error, check the error code.
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				// Handle foreign key or unique constraint violations (e.g., invalid owner or duplicate currency).
				return c.Status(fiber.StatusForbidden).JSON(errorResponse(err)) // Send a forbidden response with the error message.
			}
		}
	}

	// If it's not a specific PostgreSQL error, handle it as a generic internal server error.
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err)) // Send an internal server error response with the error message.
	}

	// 5. Success Response
	// If everything is successful, send a JSON response with the created account data.
	return c.Status(fiber.StatusOK).JSON(account)
}

// ------------ **API Functionality for Getting an Account** ------------

// **(server *Server) getAccount** GET. Handles retrieving an account by /accounts/:id endpoint.
func (server *Server) getAccount(c *fiber.Ctx) error {
	// 1. Parse account ID from URL parameters
	var req getAccountRequest
	if err := c.ParamsParser(&req); err != nil || req.ID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(fmt.Errorf("invalid account ID")))
	}

	// 2. Retrieve account from database
	account, err := server.store.GetAccount(c.Context(), req.ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(fmt.Errorf("account not found")))
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		}
	}

	// 3. Extract authentication payload
	payload, ok := c.Locals(authorizationPayloadKey).(*token.Payload)
	if !ok || payload == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(fmt.Errorf("unauthorized")))
	}

	// 4. Check ownership
	if account.Owner != payload.Username {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(fmt.Errorf("account doesn't belong to the authenticated user")))
	}

	// 5. Return successful response
	return c.Status(fiber.StatusOK).JSON(account)
}

// ------------ **API Functionality for Listing Accounts** ------------

// **(server *Server) listAccount** GET. Handles retrieving a paginated list of accounts for the authenticated user.
func (server *Server) listAccount(c *fiber.Ctx) error {
	// 1. Parse query parameters
	var req listAccountRequest

	// Attempt to parse query parameters to the `listAccountRequest` struct.
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}
	// Validate query params
	if err := server.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// 2. Extract Authentication Information
	// Retrieve the authentication payload from the context.
	// This assumes the context has a key named `authorizationPayloadKey` containing the token payload.
	payload, ok := c.Locals(authorizationPayloadKey).(*token.Payload)
	if !ok || payload == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(fmt.Errorf("unauthorized")))
	}
	// Get username from payload
	username := payload.Username

	// 3. Prepare Database Arguments
	// Create a `db.ListAccountsParams` struct to hold arguments for the database call.
	arg := db.ListAccountsParams{
		Owner:  username,
		Limit:  int64(req.PageSize),
		Offset: int64((req.PageID - 1) * req.PageSize),
	}

	// 4. Call Database Function
	// Call the `ListAccounts` function from the `server.store` object (assuming it interacts with the database).
	// This function likely retrieves a list of accounts for the authenticated user with pagination.
	accounts, err := server.store.ListAccounts(c.Context(), arg)

	// 5. Handle Database Errors
	if err != nil {
		// Check if the error is a specific error indicating no rows found (`sql.ErrNoRows`).
		if err == sql.ErrNoRows {
			// Handle the case where no account is found.
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err)) // Send a not found response with the error message.
		}

		// If it's not a specific error, handle it as a generic internal server error.
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err)) // Send an internal server error response with the error message.
	}

	// 6. Success Response
	// If everything is successful, send a JSON response with the list of accounts.
	return c.Status(fiber.StatusOK).JSON(accounts)
}

// ------------ **API Functionality for Deleting an Account** ------------

// deleteAccount handles DELETE /accounts/:id endpoint
func (server *Server) deleteAccount(c *fiber.Ctx) error {
	// 1. Parse account ID from URL path
	var req deleteAccountRequest
	if err := c.ParamsParser(&req); err != nil || req.ID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(fmt.Errorf("invalid account ID")))
	}

	// 2. Retrieve the account from DB to confirm existence and ownership
	account, err := server.store.GetAccount(c.Context(), req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(fmt.Errorf("account not found")))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 3. Retrieve authentication payload from context
	authPayload, ok := c.Locals(authorizationPayloadKey).(*token.Payload)
	if !ok || authPayload == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(fmt.Errorf("unauthorized")))
	}

	// 4. Ensure account belongs to the authenticated user
	if account.Owner != authPayload.Username {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(fmt.Errorf("account doesn't belong to the authenticated user")))
	}

	// 5. Perform account deletion
	if err := server.store.DeleteAccount(c.Context(), req.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 6. Respond with success message
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Account deleted successfully",
		"accountID": req.ID,
	})
}
