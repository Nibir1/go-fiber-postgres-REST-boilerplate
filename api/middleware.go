package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nibir1/go-fiber-postgres-REST-boilerplate/token"
)

// Key used to store payload in Fiber context locals
const authorizationPayloadKey = "authorization_payload"

// ---------------------------
// Auth Middleware for Fiber
// ---------------------------

// authMiddlewareFiber creates a Fiber middleware function that validates JWT/Paseto tokens.
// It ensures that requests to protected routes include a valid Authorization header.
func authMiddlewareFiber(tokenMaker token.Maker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Retrieve the Authorization header from the request
		authorizationHeader := c.Get("Authorization")
		if len(authorizationHeader) == 0 {
			// If no header provided, return 401 Unauthorized
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization header is missing",
			})
		}

		// Split the header into type and token value
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			// Header does not contain a valid type/token pair
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		authType := fields[0]
		accessToken := fields[1]

		// Ensure the auth type is "Bearer" (case-insensitive)
		if strings.ToLower(authType) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unsupported authorization type",
			})
		}

		// Verify the access token using the provided token maker
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			// Invalid or expired token
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token: " + err.Error(),
			})
		}

		// Store the verified payload in Fiber's Locals, so it can be accessed in handlers
		c.Locals(authorizationPayloadKey, payload)

		// Continue to the next middleware or handler
		return c.Next()
	}
}

// ---------------------------
// Usage Example in Fiber App
// ---------------------------
//
// app := fiber.New()
// tokenMaker, _ := token.NewPasetoMaker("some-secret-key")
//
// // Protected route
// app.Get("/accounts", authMiddlewareFiber(tokenMaker), getAccountsHandler)
