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

// transferRequest represents the expected JSON body for creating a transfer
type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"` // ID of sender account, must be > 0
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`   // ID of receiver account, must be > 0
	Amount        int64  `json:"amount" binding:"required,gt=0"`           // Transfer amount, must be > 0
	Currency      string `json:"currency" binding:"required,currency"`     // Currency, validated by custom validator
}

// ---------------------------
// Handlers
// ---------------------------

// createTransfer handles POST /transfers endpoint
func (server *Server) createTransfer(c *fiber.Ctx) error {
	// 1. Parse JSON request body into transferRequest struct
	var req transferRequest
	if err := c.BodyParser(&req); err != nil {
		// Invalid request JSON → 400 Bad Request
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// 2. Validate "from" account exists and currency matches
	fromAccount, valid := server.validAccount(c, req.FromAccountID, req.Currency)
	if !valid {
		// Error handled in validAccount → return immediately
		return nil
	}

	// 3. Retrieve authenticated user from context
	authPayload := c.Locals(authorizationPayloadKey).(*token.Payload)

	// 4. Ensure the "from" account belongs to the authenticated user
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
	}

	// 5. Validate "to" account exists and currency matches
	_, valid = server.validAccount(c, req.ToAccountID, req.Currency)
	if !valid {
		return nil // Error handled in validAccount
	}

	// 6. Prepare arguments for the transfer transaction
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	// 7. Execute transfer transaction using SQLC
	result, err := server.store.TransferTx(c.Context(), arg)
	if err != nil {
		// Transaction error → 500 Internal Server Error
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 8. Return 200 OK with transaction result
	return c.Status(fiber.StatusOK).JSON(result)
}

// ---------------------------
// Helper Functions
// ---------------------------

// validAccount validates that an account exists and has the correct currency
func (server *Server) validAccount(c *fiber.Ctx, accountID int64, currency string) (db.Account, bool) {
	// 1. Fetch account from the database
	account, err := server.store.GetAccount(c.Context(), accountID)
	if err != nil {
		// Account not found → 404 Not Found
		if errors.Is(err, sql.ErrNoRows) {
			_ = c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return account, false
		}

		// Other DB errors → 500 Internal Server Error
		_ = c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return account, false
	}

	// 2. Check if account currency matches the requested currency
	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		_ = c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return account, false
	}

	// 3. Account is valid
	return account, true
}
