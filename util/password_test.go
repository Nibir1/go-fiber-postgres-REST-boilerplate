package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt" // Used for password hashing with bcrypt
)

// TestPassword tests the HashPassword and CheckPassword functions
func TestPassword(t *testing.T) {
	// Generate a random string to use as password
	password := RandomString(6)

	// Hash the password with bcrypt
	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err, "Failed to hash password") // Assert no errors during hashing

	// Assert that the hashed password is not empty
	require.NotEmpty(t, hashedPassword1, "Empty hashed password")

	// Check if the original password matches the hashed password
	err = CheckPassword(password, hashedPassword1)
	require.NoError(t, err, "Password mismatch") // Assert successful password check

	// Generate a different random string as a wrong password
	wrongPassword := RandomString(6)

	// Check if the wrong password matches the original hashed password (should fail)
	err = CheckPassword(wrongPassword, hashedPassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error(), "Expected bcrypt error for wrong password") // Assert specific error for mismatch

	// Hash the password again (should result in a different hash due to bcrypt salting)
	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err, "Failed to hash password") // Assert no errors during second hashing

	// Assert that the second hashed password is not empty
	require.NotEmpty(t, hashedPassword2, "Empty hashed password")

	// Assert that the two hashed passwords are different due to bcrypt salting
	require.NotEqual(t, hashedPassword1, hashedPassword2, "Hashed passwords should be different")
}
