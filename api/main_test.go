package api

import (
	"os"
	"testing"
	"time"

	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
	"github.com/stretchr/testify/require"
)

// ---------------------------
// Utility Functions
// ---------------------------

// newFiberTestServer creates a new server instance with Fiber for testing purposes
func newFiberTestServer(t *testing.T, store db.Store) *Server {
	// Generate random token secret for test
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32), // Random key for signing tokens
		AccessTokenDuration: time.Minute,           // Token validity duration
	}

	// Create a new Fiber server with configuration and store
	server, err := NewServer(config, store)
	require.NoError(t, err) // Ensure server creation does not fail

	return server
}

// ---------------------------
// TestMain Entry Point
// ---------------------------

func TestMain(m *testing.M) {
	// Fiber does not have a "TestMode" like Gin.
	// We can still use testing without starting a real server.
	// Any Fiber logs can be silenced in production or testing if needed.

	// Run all tests
	os.Exit(m.Run())
}
