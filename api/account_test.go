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
// Helper Functions
// ---------------------------

// randomAccount creates a fake account for testing
func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomInt(0, 1000),
		Currency: util.USD,
	}
}

// randomUserStruct generates a random test user
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

// requireBodyMatchAccount asserts that the JSON response matches expected account
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := body.ReadBytes(0)
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

// addAuthorization adds a valid Bearer token to the request
func addAuthorization(req *http.Request, tokenMaker token.Maker, username string, duration time.Duration) {
	accessToken, _, err := tokenMaker.CreateToken(username, duration)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
}

// ---------------------------
// Fiber Test Server Helper
// ---------------------------

// newFiberTestServer creates a test Fiber server with routes for account API
func newFiberTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey: util.RandomString(32),
	}
	server, err := NewServer(config, store)
	require.NoError(t, err)
	return server
}

// ---------------------------
// TestCreateAccountAPI
// ---------------------------

func TestCreateAccountAPI(t *testing.T) {
	user := randomUserStruct()
	account := randomAccount(user.Username)

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
				addAuthorization(req, tokenMaker, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
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
				// Do not add auth header
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newFiberTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(data))
			req.Header.Set("Content-Type", "application/json")

			tc.setupAuth(req, server.tokenMaker)

			// Fiber's Test returns *http.Response
			resp, err := server.app.Test(req, -1)
			require.NoError(t, err)

			// Copy response body to recorder for consistent assertions
			bodyBytes := new(bytes.Buffer)
			_, err = bodyBytes.ReadFrom(resp.Body)
			require.NoError(t, err)
			recorder.Body = bodyBytes
			recorder.Code = resp.StatusCode

			tc.checkResponse(recorder)
		})
	}
}
