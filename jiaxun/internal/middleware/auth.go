package middleware

import (
	"errors"
	"fmt"
	"jiaxun/internal/config"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in the JWT
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Public paths that don't require authentication
var publicPaths = []string{
	"/api/auth/login",
	"/api/auth/register",
	"/api/health",
}

// AuthMiddleware authenticates requests using JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for public paths
		path := c.Request.URL.Path
		for _, p := range publicPaths {
			if strings.HasPrefix(path, p) {
				c.Next()
				return
			}
		}

		// Get the JWT from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// The token should be in the format "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate the token
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			jwtSecret := []byte(config.GetConfig().Application.Secret) // Get secret key from config
			return jwtSecret, nil
		})

		// Handle token validation errors
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(401, gin.H{"error": "Token has expired"})
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				c.JSON(401, gin.H{"error": "Malformed token"})
			} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				c.JSON(401, gin.H{"error": "Invalid token signature"})
			} else {
				c.JSON(401, gin.H{"error": "Invalid token: " + err.Error()})
			}
			c.Abort()
			return
		}

		// Ensure the token is valid
		if !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set the user ID and role in the context for use in handlers
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// GenerateToken generates a new JWT token for a user
func GenerateToken(userID uint, email, role string) (string, error) {
	// Set the expiration time for the token (e.g., 24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the JWT claims
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "jiaxun",
		},
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	jwtSecret := []byte(config.GetConfig().Application.Secret)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
