package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver for database/sql
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
)

// testQueries will be used to execute SQL queries in tests
var testQueries *Queries

// testDB is the database connection used for testing
var testDB *sql.DB

// TestMain is the entry point for testing in this package.
// It sets up the test database connection and initializes testQueries before running tests.
func TestMain(m *testing.M) {
	// Load configuration (like DB connection info) from the project root directory
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err) // Stop tests if config can't be loaded
	}

	// Open a connection to the test database using the loaded config
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to db:", err) // Stop tests if DB can't be connected
	}

	// Initialize testQueries with the test database connection
	testQueries = New(testDB)

	// Run all tests in the package and exit with the result code
	os.Exit(m.Run())
}
