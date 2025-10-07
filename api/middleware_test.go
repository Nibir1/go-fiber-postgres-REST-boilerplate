package api

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
	"github.com/stretchr/testify/require"
)

// addAuthorization adds a valid Bearer token to the request
func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

// ---------------------------
// Tests for Auth Middleware
// ---------------------------

func TestAuthMiddlewareFiber(t *testing.T) {
	username := util.RandomOwner()

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, req *http.Request, maker token.Maker)
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationTypeBearer, username, time.Minute)
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name:      "NoAuthorization",
			setupAuth: func(t *testing.T, req *http.Request, maker token.Maker) {},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, "unsupported", username, time.Minute)
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, req *http.Request, maker token.Maker) {
				req.Header.Set(authorizationHeaderKey, "malformed-token")
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationTypeBearer, username, -time.Minute)
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use your full Server struct
			server := newFiberTestServer(t, nil) // nil store is fine; middleware test doesn't need DB

			// Register a test route with the middleware
			authPath := "/auth"
			server.app.Get(authPath, authMiddlewareFiber(server.tokenMaker), func(ctx *fiber.Ctx) error {
				return ctx.JSON(fiber.Map{"status": "ok"})
			})

			// Create request
			req, err := http.NewRequest(http.MethodGet, authPath, bytes.NewReader(nil))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			// Use Fiber Test to execute request
			resp, err := server.app.Test(req, -1)
			require.NoError(t, err)

			tc.checkResponse(t, resp)
		})
	}
}
