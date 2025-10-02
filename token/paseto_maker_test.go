package token

import (
	"testing" // Used for test assertions (require)
	"time"

	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/util"
	"github.com/stretchr/testify/require"
)

// TestPasetoMaker tests the functionality of the PasetoMaker struct
func TestPasetoMaker(t *testing.T) {
	// Create a new PasetoMaker instance with a random key
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err, "Failed to create PasetoMaker") // Assert no errors

	// Generate a random username
	username := util.RandomOwner()
	// Set a token duration of one minute
	duration := time.Minute

	// Capture the current time for later verification
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	// Create a new token with the username and duration
	token, payload, err := maker.CreateToken(username, duration)
	require.NoError(t, err, "Failed to create token")       // Assert no errors
	require.NotEmpty(t, token, "Empty token generated")     // Assert token is not empty
	require.NotEmpty(t, payload, "Empty payload generated") // Assert payload is not empty

	// Verify the generated token
	payload, err = maker.VerifyToken(token)
	require.NoError(t, err, "Failed to verify token") // Assert no errors

	// Assert various properties of the extracted payload
	require.NotZero(t, payload.ID, "Payload ID is zero")
	require.Equal(t, username, payload.Username, "Username mismatch")
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second, "IssuedAt time mismatch")
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second, "ExpiredAt time mismatch")
}

// TestExpiredPasetoToken tests handling of expired tokens
func TestExpiredPasetoToken(t *testing.T) {
	// Create a new PasetoMaker instance with a random key
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err, "Failed to create PasetoMaker") // Assert no errors

	// Create a token with a negative duration (expired)
	token, payload, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoError(t, err, "Failed to create token")       // Assert no errors
	require.NotEmpty(t, token, "Empty token generated")     // Assert token is not empty
	require.NotEmpty(t, payload, "Empty payload generated") // Assert payload is not empty

	// Verify the token (should fail due to expiration)
	payload, err = maker.VerifyToken(token)
	require.Error(t, err, "Expected error for expired token")                       // Assert error occurs
	require.EqualError(t, err, ErrExpiredToken.Error(), "Unexpected error message") // Assert specific error
	require.Nil(t, payload, "Payload should be nil for invalid token")              // Assert no payload extracted
}
