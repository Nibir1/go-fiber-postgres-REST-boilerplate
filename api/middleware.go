package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"
)

// Constants for authorization header and payload keys
const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// authMiddlewareFiber returns a Fiber middleware function that validates JWT/Paseto tokens.
// It ensures requests to protected routes include a valid Authorization header.
func authMiddlewareFiber(tokenMaker token.Maker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Retrieve the Authorization header from the request
		authorizationHeader := strings.TrimSpace(c.Get(authorizationHeaderKey))
		if len(authorizationHeader) == 0 {
			// No header provided
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization header is missing",
			})
		}

		// Split the header into type and token value
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			// Header format is invalid
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		authType := fields[0]
		accessToken := fields[1]

		// Ensure the auth type is "Bearer" (case-insensitive)
		if strings.ToLower(authType) != authorizationTypeBearer {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unsupported authorization type",
			})
		}

		// Verify the token
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			// Token is invalid or expired
			// Internal error can be logged instead of sent to client in production
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Store the verified payload in Fiber locals for handlers to access
		c.Locals(authorizationPayloadKey, payload)

		// Call the next middleware or route handler
		return c.Next()
	}
}
