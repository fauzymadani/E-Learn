package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"elearning/internal/domain"
	"elearning/pkg/token"
)

const (
	authHeaderKey      = "Authorization"
	authTypeBearer     = "bearer"
	authPayloadContext = "auth_payload"
	authTokenContext   = "auth_token" // Store the actual token string
)

func errorResponse(msg string) gin.H {
	return gin.H{"error": msg}
}

// AuthMiddleware verifies JWT and stores claims into context
func AuthMiddleware(tokenMaker token.TokenMaker, blacklist token.TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(authHeaderKey)
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("missing authorization header"))
			return
		}

		parts := strings.Fields(header)
		if len(parts) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("invalid authorization header format"))
			return
		}

		if strings.ToLower(parts[0]) != authTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("authorization type must be Bearer"))
			return
		}

		tokenStr := parts[1]

		// Check if token is blacklisted
		if blacklist != nil && blacklist.IsBlacklisted(tokenStr) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("token has been revoked"))
			return
		}

		claims, err := tokenMaker.VerifyToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("invalid or expired token"))
			return
		}

		c.Set(authPayloadContext, claims)
		c.Set(authTokenContext, tokenStr) // Store token for logout
		c.Next()
	}
}

// RequireRole checks whether the user has one of the allowed roles
func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		payloadRaw, exists := c.Get(authPayloadContext)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("no authorization payload found"))
			return
		}

		claims, ok := payloadRaw.(*token.Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("invalid authorization payload"))
			return
		}

		userRole := domain.UserRole(claims.Role)

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, errorResponse("insufficient permissions"))
	}
}

// GetCurrentUser extracts current JWT user claims
func GetCurrentUser(c *gin.Context) (*token.Claims, error) {
	payload, exists := c.Get(authPayloadContext)
	if !exists {
		return nil, errors.New("auth payload not found")
	}

	claims, ok := payload.(*token.Claims)
	if !ok {
		return nil, errors.New("invalid auth payload")
	}

	return claims, nil
}

// GetCurrentToken extracts the current JWT token string
func GetCurrentToken(c *gin.Context) (string, error) {
	tokenStr, exists := c.Get(authTokenContext)
	if !exists {
		return "", errors.New("auth token not found")
	}

	tokenString, ok := tokenStr.(string)
	if !ok {
		return "", errors.New("invalid auth token")
	}

	return tokenString, nil
}
