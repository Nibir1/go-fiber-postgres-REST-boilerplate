package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store interface defines all methods for database operations, including queries and transfer transactions.
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions.
// It embeds Queries (for generated query methods) and holds a reference to the database connection.
type SQLStore struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new SQLStore instance with the given database connection.
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db), // Initialize Queries with the db connection
	}
}

// execTX executes a function within a database transaction.
// It begins a transaction, executes the provided function, and commits or rolls back as needed.
func (store *SQLStore) execTX(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil) // Start a new transaction
	if err != nil {
		return err
	}

	q := New(tx) // Create a new Queries instance using the transaction
	err = fn(q)  // Execute the function with the transactional Queries

	if err != nil {
		// If the function returns an error, roll back the transaction
		if rbErr := tx.Rollback(); rbErr != nil {
			// If rollback also fails, return both errors
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	// Commit the transaction if no error occurred
	return tx.Commit()
}

// TransferTxParams contains the input parameters for a transfer transaction.
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of a transfer transaction, including all affected records.
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// txKey is used as a context key for transaction naming/logging (optional, for debugging).
var txKey = struct{}{}

// TransferTx performs a money transfer from one account to another within a single database transaction.
// It creates a transfer record, adds account entries, and updates account balances atomically.
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	// Execute all operations within a transaction using execTX
	err := store.execTX(ctx, func(q *Queries) error {
		var err error

		// Optional: Retrieve transaction name from context for logging/debugging
		txName := ctx.Value(txKey)

		// 1. Create a transfer record
		fmt.Println(txName, "create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// 2. Create an entry for the sender (negative amount)
		fmt.Println(txName, "create entry1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// 3. Create an entry for the receiver (positive amount)
		fmt.Println(txName, "create entry2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// 4. Update account balances in a consistent order to avoid deadlocks
		if arg.FromAccountID < arg.ToAccountID {
			fmt.Println(txName, "From Account 1 -> To Account 2")
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			fmt.Println(txName, "From Account 2 -> To Account 1")
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}
		return err // Return any error from addMoney (if any)
	})

	return result, err
}

// addMoney updates the balances of two accounts atomically within a transaction.
// It returns the updated account records and any error encountered.
func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}
