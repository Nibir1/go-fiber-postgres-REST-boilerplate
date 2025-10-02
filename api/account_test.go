package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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
// Utility Functions
// ---------------------------

// randomAccount creates a random account for testing
func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomInt(0, 1000),
		Currency: util.USD,
	}
}

// requireBodyMatchAccount checks that the response body matches expected account
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := body.ReadBytes(0) // read entire body
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("failed to read body: %v", err)
	}

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account.ID, gotAccount.ID)
	require.Equal(t, account.Owner, gotAccount.Owner)
	require.Equal(t, account.Balance, gotAccount.Balance)
	require.Equal(t, account.Currency, gotAccount.Currency)
}

// ---------------------------
// Tests for Account API
// ---------------------------

func TestCreateAccountAPI(t *testing.T) {
	user := randomUserStruct() // Create a random test user

	account := randomAccount(user.Username)
	account.Currency = util.USD

	testCases := []struct {
		name          string
		body          fiber.Map
		setupAuth     func(req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"currency": account.Currency,
			},
			setupAuth: func(req *http.Request, tokenMaker token.Maker) {
				addAuthorizationToFiberTest(req, tokenMaker, user.Username)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).Times(1).Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "UnauthorizedUser",
			body: fiber.Map{
				"currency": account.Currency,
			},
			setupAuth: func(req *http.Request, tokenMaker token.Maker) {
				// No authorization header added
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	// Loop through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newFiberTestServer(t, store) // Initialize Fiber server
			recorder := httptest.NewRecorder()

			// Marshal body to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(data))
			require.NoError(t, err)

			// Add authorization header if required
			tc.setupAuth(req, server.tokenMaker)

			// Perform the request
			server.app.Test(req, -1)
			tc.checkResponse(recorder)
		})
	}
}

// ---------------------------
// Helper Functions for Fiber Testing
// ---------------------------

// addAuthorizationToFiberTest adds Bearer token to the request for testing
func addAuthorizationToFiberTest(req *http.Request, tokenMaker token.Maker, username string) {
	accessToken, _, err := tokenMaker.CreateToken(username, time.Minute)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
}

// randomUserStruct returns a random test user
func randomUserStruct() db.User {
	password := util.RandomString(6)
	hashed, _ := util.HashPassword(password)
	return db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashed,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
}
