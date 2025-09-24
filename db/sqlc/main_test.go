package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq" // Import the PostgreSQL driver anonymously for its side-effects
)

// Database connection configuration constants
const (
	dbDriver = "postgres"                                                            // Name of the database driver
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" // Connection string for the test database
)

// testQueries will be used to execute SQL queries in tests
var testQueries *Queries

// TestMain is the entry point for testing in this package.
// It allows setup before tests run and teardown after all tests finish.
func TestMain(m *testing.M) {
	// Open a connection to the test database using the provided driver and source
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		// If the connection fails, log the error and stop the tests
		log.Fatal("cannot connect to db:", err)
	}
	// Initialize testQueries with a new Queries instance using the database connection
	testQueries = New(conn)

	// Run all tests. os.Exit is used to ensure deferred functions are not run.
	os.Exit(m.Run())
}
