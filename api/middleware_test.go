package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
	"github.com/stretchr/testify/require"
)

// ---------------------------
// Helper Function
// ---------------------------

// addAuthorizationFiber sets the Authorization header for a Fiber HTTP request
func addAuthorizationFiber(t *testing.T, req *http.Request, tokenMaker token.Maker, authType, username string, duration time.Duration) {
	// Create a token for the test user
	accessToken, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)      // Ensure token creation did not fail
	require.NotEmpty(t, payload) // Ensure payload is returned

	// Set the Authorization header in the request
	req.Header.Set("Authorization", authType+" "+accessToken)
}

// ---------------------------
// Main Auth Middleware Test
// ---------------------------

func TestAuthMiddlewareFiber(t *testing.T) {
	// Random username for testing
	username := util.RandomOwner()

	// Define test cases for different scenarios
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				// Set a valid Bearer token
				addAuthorizationFiber(t, req, tokenMaker, "Bearer", username, time.Minute)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// Expect 200 OK
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				// Do not set Authorization header
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// Expect 401 Unauthorized
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				// Set an unsupported auth type
				addAuthorizationFiber(t, req, tokenMaker, "Unsupported", username, time.Minute)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// Expect 401 Unauthorized
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				// Set a token that already expired
				addAuthorizationFiber(t, req, tokenMaker, "Bearer", username, -time.Minute)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// Expect 401 Unauthorized
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
	}

	// Execute each test case
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// Initialize Fiber app
			app := fiber.New()

			// Create a token maker for testing
			tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
			require.NoError(t, err)

			// Protected route using the auth middleware
			app.Get("/auth", func(c *fiber.Ctx) error {
				// Call the actual auth middleware
				if err := authMiddlewareFiber(tokenMaker)(c); err != nil {
					return err
				}

				// Respond 200 OK if authorized
				return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
			})

			// Create a test HTTP request
			req := httptest.NewRequest(http.MethodGet, "/auth", nil)

			// Setup Authorization header according to test case
			if tc.setupAuth != nil {
				tc.setupAuth(t, req, tokenMaker)
			}

			// Perform the request through Fiber
			recorder := httptest.NewRecorder()
			app.Test(req, -1) // Timeout -1 waits indefinitely

			// Validate response according to test case
			tc.checkResponse(t, recorder)
		})
	}
}
