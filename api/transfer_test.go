package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	mockdb "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/mock"
	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
	"github.com/stretchr/testify/require"
)

// ---------------------------
// Helper Function
// ---------------------------

// addAuthorizationFiber is defined in middleware_test.go and reused here.

// createTransferFiber handles transfer requests in Fiber.
func (server *Server) createTransferFiber(c *fiber.Ctx) error {
	// This is a placeholder implementation for testing purposes.
	// You should implement the actual logic as in your main server code.
	return c.SendStatus(http.StatusOK)
}

// ---------------------------
// Main Transfer API Test
// ---------------------------

func TestTransferAPI(t *testing.T) {
	// Test data
	amount := int64(10)

	// Create three test users
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)
	user3, _ := randomUser(t)

	// Create accounts for the users
	account1 := randomAccount(user1.Username)
	account2 := randomAccount(user2.Username)
	account3 := randomAccount(user3.Username)

	account1.Currency = util.USD
	account2.Currency = util.USD
	account3.Currency = util.EUR // Different currency for mismatch tests

	// ---------------------------
	// Define Test Cases
	// ---------------------------
	testCases := []struct {
		name          string
		body          fiber.Map // Use Fiber's map type for JSON body
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorizationFiber(t, req, tokenMaker, "Bearer", user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Expect GetAccount for both accounts
				store.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), account2.ID).Times(1).Return(account2, nil)

				// Expect TransferTx with the correct arguments
				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}
				store.EXPECT().TransferTx(gomock.Any(), arg).Times(1).Return(db.TransferTxResult{}, nil)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "UnauthorizedUser",
			body: fiber.Map{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				// User2 tries to transfer from user1's account
				addAuthorizationFiber(t, req, tokenMaker, "Bearer", user2.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), account2.ID).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		// Add more test cases like NoAuthorization, CurrencyMismatch, NegativeAmount, etc.
	}

	// ---------------------------
	// Execute Test Cases
	// ---------------------------
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// Create mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// Initialize Fiber app
			app := fiber.New()
			tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
			require.NoError(t, err)
			server := &Server{
				store:      store,
				tokenMaker: tokenMaker,
			}

			// Setup route
			app.Post("/transfers", func(c *fiber.Ctx) error {
				return server.createTransferFiber(c) // Fiber version of createTransfer
			})

			// Marshal body to JSON
			bodyData, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(bodyData))
			req.Header.Set("Content-Type", "application/json")

			// Setup authorization
			if tc.setupAuth != nil {
				tc.setupAuth(t, req, server.tokenMaker)
			}

			recorder := httptest.NewRecorder()
			app.Test(req, -1) // Perform the request

			// Check response
			tc.checkResponse(recorder)
		})
	}
}
