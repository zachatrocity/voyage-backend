package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// APIKeyMiddleware returns an Echo middleware that enforces API key auth.
// If apiKey is empty, the middleware is a no-op (auth disabled).
func APIKeyMiddleware(apiKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if apiKey == "" {
				return next(c)
			}

			// Check X-API-Key header first, then Authorization: Bearer
			key := c.Request().Header.Get("X-API-Key")
			if key == "" {
				auth := c.Request().Header.Get("Authorization")
				key = strings.TrimPrefix(auth, "Bearer ")
			}

			if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "unauthorized — set X-API-Key header with your VOYAGE_API_KEY",
				})
			}

			return next(c)
		}
	}
}
