package token

import (
	"errors"
	"time"

	"github.com/google/uuid" // Used for generating UUIDs
)

// Predefined errors returned by the VerifyToken function
var (
	// Error for invalid tokens (e.g., structure, signature)
	ErrInvalidToken = errors.New("token is invalid")
	// Error for expired tokens
	ErrExpiredToken = errors.New("token has expired")
)

// Payload represents the data contained within a token
type Payload struct {
	ID        uuid.UUID `json:"id"`         // Unique identifier for the token (UUID)
	Username  string    `json:"username"`   // Username associated with the token
	IssuedAt  time.Time `json:"issued_at"`  // Time the token was issued
	ExpiredAt time.Time `json:"expired_at"` // Time the token expires
}

// NewPayload creates a new Payload object with a username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	// Generate a random UUID for the token ID
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err // Return error if generation fails
	}

	// Create a new Payload object with provided data and calculated expiration
	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

// Valid checks if the Payload object represents a valid, unexpired token
func (payload *Payload) Valid() error {
	// Check if current time is past the token's expiration
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken // Return error if expired
	}
	return nil // Token is valid
}
