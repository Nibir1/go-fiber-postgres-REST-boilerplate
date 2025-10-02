package token

import (
	"errors"
	"fmt"
	"time"

	// External library for JWT (JSON Web Token) handling
	"github.com/dgrijalva/jwt-go"
)

const minSecretKeySize = 32

// JWTMaker is a struct representing a JSON Web Token maker
type JWTMaker struct {
	secretKey string // The secret key used for signing tokens
}

// NewJWTMaker creates a new JWTMaker instance
func NewJWTMaker(secretKey string) (Maker, error) {
	// Validate secret key size
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	// Return a pointer to the new JWTMaker object
	return &JWTMaker{secretKey}, nil
}

// CreateToken generates a new JWT token for a given username and duration
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	// Create a new Payload object with the provided username and duration
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	// Create a new JWT token with the HS256 signing method and the payload
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	// Sign the token using the maker's secret key
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, payload, err
}

// VerifyToken checks if the provided token is valid or not
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// Define a key function to validate the token signature
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Check if the token signing method is HMAC
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		// Return the maker's secret key as bytes for verification
		return []byte(maker.secretKey), nil
	}

	// Parse the token with the provided key function and expected Payload type
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		// Handle different types of validation errors
		verr, ok := err.(*jwt.ValidationError)
		if ok {
			if errors.Is(verr.Inner, ErrExpiredToken) {
				return nil, ErrExpiredToken
			}
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	// Check if the claims are of the expected Payload type
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Return the extracted Payload object if the token is valid
	return payload, nil
}
