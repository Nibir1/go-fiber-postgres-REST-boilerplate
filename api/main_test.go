package api

import (
	"os"
	"testing"

	// External dependencies

	db "github.com/nibir1/go-fiber-postgres-REST-boilerplate/db/sqlc"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
	"github.com/stretchr/testify/require"
)

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

// TestMain is the entry point for running tests
func TestMain(m *testing.M) {
	// Run the tests defined in other parts of the codebase
	os.Exit(m.Run())
}
