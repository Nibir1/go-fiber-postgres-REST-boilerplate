package api

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc" // SQLC database package
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"      // Token handling (JWT/Paseto)
)

// ---------------------------
// Request Structs
// ---------------------------

// createAccountRequest represents the JSON body for creating a new account
type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required,alphanum"`    // Account owner username
	Currency string `json:"currency" binding:"required,currency"` // Currency for the account (validated)
}

// listAccountRequest represents query parameters for listing accounts
type listAccountRequest struct {
	PageID   int32 `query:"page_id" binding:"required,min=1"`           // Current page number, min 1
	PageSize int32 `query:"page_size" binding:"required,min=5,max=100"` // Items per page, 5–100
}

// ---------------------------
// Handlers
// ---------------------------

// createAccount handles POST /accounts endpoint
func (server *Server) createAccount(c *fiber.Ctx) error {
	// 1. Parse JSON body
	var req createAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// 2. Extract authenticated username from token payload
	authPayload := c.Locals(authorizationPayloadKey).(*token.Payload)
	if req.Owner != authPayload.Username {
		// Ensure user can only create account for themselves
		err := errors.New("cannot create account for another user")
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
	}

	// 3. Prepare arguments for SQLC
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0, // Default balance
	}

	// 4. Execute DB query
	account, err := server.store.CreateAccount(c.Context(), arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 5. Return 200 OK with created account
	return c.Status(fiber.StatusOK).JSON(account)
}

// getAccount handles GET /accounts/:id endpoint
func (server *Server) getAccount(c *fiber.Ctx) error {
	// 1. Parse account ID from URL
	accountID, err := c.ParamsInt("id")
	if err != nil || accountID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(fmt.Errorf("invalid account ID")))
	}

	// 2. Fetch account from database
	account, err := server.store.GetAccount(c.Context(), int64(accountID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 3. Ensure authenticated user owns this account
	authPayload := c.Locals(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(errors.New("account does not belong to user")))
	}

	// 4. Return account
	return c.Status(fiber.StatusOK).JSON(account)
}

// listAccount handles GET /accounts endpoint
func (server *Server) listAccount(c *fiber.Ctx) error {
	// 1. Parse query parameters
	var req listAccountRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// 2. Extract authenticated username
	authPayload := c.Locals(authorizationPayloadKey).(*token.Payload)

	// 3. Calculate DB offset for pagination
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  int64(req.PageSize),
		Offset: int64((req.PageID - 1) * req.PageSize),
	}

	// 4. Execute DB query
	accounts, err := server.store.ListAccounts(c.Context(), arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 5. Return list of accounts
	return c.Status(fiber.StatusOK).JSON(accounts)
}

// deleteAccount handles DELETE /accounts/:id endpoint
func (server *Server) deleteAccount(c *fiber.Ctx) error {
	// 1. Parse account ID from URL
	accountID, err := c.ParamsInt("id")
	if err != nil || accountID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(fmt.Errorf("invalid account ID")))
	}

	// 2. Fetch account to ensure ownership
	account, err := server.store.GetAccount(c.Context(), int64(accountID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	authPayload := c.Locals(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(errors.New("account does not belong to user")))
	}

	// 3. Delete account
	if err := server.store.DeleteAccount(c.Context(), int64(accountID)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 4. Return success
	return c.SendStatus(fiber.StatusOK)
}
