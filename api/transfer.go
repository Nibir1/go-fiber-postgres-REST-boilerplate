package api

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc" // SQLC database package
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"      // Token handling (JWT/Paseto)
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
)

// ---------------------------
// Request Structs
// ---------------------------

// transferRequest represents the expected JSON body for creating a transfer
// @Description Transfer request payload
type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"` // ID of sender account, must be > 0
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`   // ID of receiver account, must be > 0
	Amount        int64  `json:"amount" binding:"required,gt=0"`           // Transfer amount, must be > 0
	Currency      string `json:"currency" binding:"required,currency"`     // Currency, validated by custom validator
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

// ---------------------------
// Handlers
// ---------------------------

// createTransfer handles POST /transfers endpoint

// CreateTransfer godoc
// @Summary      Create a new money transfer
// @Description  Transfers a specified amount between two accounts if they belong to the authenticated user and share the same currency.
// @Tags         Transfers
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        transfer  body      transferRequest  true  "Transfer details"
// @Success      200       {object}  db.TransferTxResult
// @Router       /transfers [post]
func (server *Server) createTransfer(c *fiber.Ctx) error {
	// 1. Parse JSON request body into transferRequest struct
	var req transferRequest
	if err := c.BodyParser(&req); err != nil {
		// Invalid request JSON → 400 Bad Request
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// 2. Validate input fields before any DB access
	if req.Amount <= 0 {
		err := errors.New("amount must be greater than zero")
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	if !util.IsSupportedCurrency(req.Currency) {
		err := fmt.Errorf("unsupported currency: %s", req.Currency)
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	// 3. Validate "from" account exists and currency matches
	fromAccount, valid := server.validAccount(c, req.FromAccountID, req.Currency)
	if !valid {
		// Error handled in validAccount → return immediately
		return nil
	}

	// 4. Retrieve authenticated user from payload
	payload, ok := c.Locals(authorizationPayloadKey).(*token.Payload)
	if !ok || payload == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(fmt.Errorf("unauthorized")))
	}
	username := payload.Username

	// 5. Ensure the "from" account belongs to the authenticated user
	if fromAccount.Owner != username {
		err := errors.New("from account doesn't belong to the authenticated user")
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
	}

	// 6. Validate "to" account exists and currency matches
	_, valid = server.validAccount(c, req.ToAccountID, req.Currency)
	if !valid {
		return nil // Error handled in validAccount
	}

	// 7. Prepare arguments for the transfer transaction
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	// 8. Execute transfer transaction using SQLC
	result, err := server.store.TransferTx(c.Context(), arg)
	if err != nil {
		// Transaction error → 500 Internal Server Error
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	// 9. Return 200 OK with transaction result
	return c.Status(fiber.StatusOK).JSON(result)
}
