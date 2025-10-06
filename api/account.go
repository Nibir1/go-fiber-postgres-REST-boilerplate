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

// listAccountRequest represents query parameters for listing accounts
type listAccountRequest struct {
	PageID   int32 `query:"page_id" binding:"required,min=1"`           // Current page number, min 1
	PageSize int32 `query:"page_size" binding:"required,min=5,max=100"` // Items per page, 5–100
}

// ---------------------------
// Handlers
// ---------------------------

// createAccount handles the POST /accounts route
func (server *Server) createAccount(c *fiber.Ctx) error {
	var req struct {
		Currency string `json:"currency" validate:"required,currency"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	if err := server.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// -------------------
	// Get username from payload
	// -------------------
	payload, ok := c.Locals(authorizationPayloadKey).(*token.Payload)
	if !ok || payload == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(fmt.Errorf("unauthorized")))
	}
	username := payload.Username

	arg := db.CreateAccountParams{
		Owner:    username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(c.Context(), arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

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
	var req struct {
		PageID   int `query:"page_id" validate:"required,min=1"`
		PageSize int `query:"page_size" validate:"required,min=5,max=10"`
	}
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// 2. Validate query params
	if err := server.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// 3. Extract authenticated username
	authPayload := c.Locals(authorizationPayloadKey).(*token.Payload)

	// 4. Prepare DB parameters
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  int64(req.PageSize),
		Offset: int64((req.PageID - 1) * req.PageSize),
	}

	// 5. Execute query
	accounts, err := server.store.ListAccounts(c.Context(), arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 6. Return success response
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
