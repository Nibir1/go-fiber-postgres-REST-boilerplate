package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
	"github.com/stretchr/testify/require"
)

// CreateRandomAccount creates a random account for testing purposes.
// It uses a randomly generated user, balance, and currency.
func CreateRandomAccount(t *testing.T) Account {
	user := CreateRandomUser(t) // Create a random user for the account owner
	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)

	// Ensure no error occurred and the returned account is not empty
	require.NoError(t, err)
	require.NotEmpty(t, account)

	// Validate that the created account matches the input parameters
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)
	require.Equal(t, arg.Owner, account.Owner)

	// Ensure the account has a valid ID and creation timestamp
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

// TestCreateAccount verifies that an account can be created successfully.
func TestCreateAccount(t *testing.T) {
	CreateRandomAccount(t)
}

// TestGetAccount verifies that an account can be retrieved by its ID.
func TestGetAccount(t *testing.T) {
	account1 := CreateRandomAccount(t)
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	// Check that the retrieved account matches the created one
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}

// TestUpdateAccount verifies that an account's balance can be updated.
func TestUpdateAccount(t *testing.T) {
	account1 := CreateRandomAccount(t)

	arg := UpdateAccountParams{
		ID:      account1.ID,
		Balance: util.RandomMoney(),
	}

	account2, err := testQueries.UpdateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	// Check that only the balance is updated, other fields remain the same
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, arg.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}

// TestDeleteAccount verifies that an account can be deleted and is no longer retrievable.
func TestDeleteAccount(t *testing.T) {
	account1 := CreateRandomAccount(t)
	err := testQueries.DeleteAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	account2, err := testQueries.GetAccount(context.Background(), account1.ID)

	// After deletion, getting the account should return sql.ErrNoRows
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, account2)
}

// TestListAccounts verifies that multiple accounts can be listed with pagination.
func TestListAccounts(t *testing.T) {
	var lastAccount Account
	// Create 10 accounts, keeping the last one for filtering
	for i := 0; i < 10; i++ {
		lastAccount = CreateRandomAccount(t)
	}

	arg := ListAccountsParams{
		Owner:  lastAccount.Owner, // Filter by the last created account's owner
		Limit:  5,                 // Limit the number of results
		Offset: 0,                 // Start from the beginning
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	// Ensure all returned accounts belong to the specified owner
	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, lastAccount.Owner, account.Owner)
	}
}
