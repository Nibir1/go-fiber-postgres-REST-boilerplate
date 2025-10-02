package token

import (
	"fmt"
	"time"

	// External libraries for PASETO (Protocol for Authenticated Sealed Objects) encryption
	"github.com/aead/chacha20poly1305" // (dependency for key size validation)
	"github.com/o1egl/paseto"
)

// PasetoMaker is a struct representing a PASETO token maker
type PasetoMaker struct {
	// Instance of the paseto.V2 library for encryption/decryption
	paseto *paseto.V2
	// Symmetric key used for PASETO operations (as byte slice)
	symmetricKey []byte
}

// NewPasetoMaker creates a new PasetoMaker instance
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	// Validate the symmetric key size
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	// Create a new PasetoMaker object
	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}

// CreateToken generates a new token for a username and duration using PASETO encryption
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	// Create a new Payload object with username and duration
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	// Encrypt the payload using the Paseto instance and symmetric key
	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	return token, payload, err
}

// VerifyToken checks if the provided token is valid or not
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	// Create an empty Payload object to be filled during decryption
	payload := &Payload{}

	// Decrypt the token using the Paseto instance, symmetric key, and the provided payload object
	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		// Return a specific error for invalid tokens
		return nil, ErrInvalidToken
	}

	// Validate the decrypted payload (e.g., expiration check)
	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	// Return the decrypted and validated Payload object if successful
	return payload, nil
}
