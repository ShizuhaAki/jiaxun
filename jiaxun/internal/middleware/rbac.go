package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LoginRequired ensures a user is authenticated
func LoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user ID exists in context (set by AuthMiddleware)
		userID, exists := c.Get("userID")
		if !exists || userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// User is authenticated, continue
		c.Next()
	}
}

// TeacherRequired ensures a user has teacher role
func TeacherRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		userID, exists := c.Get("userID")
		if !exists || userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Then check for teacher role
		role, exists := c.Get("role")
		if !exists || role != "teacher" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Teacher privileges required"})
			c.Abort()
			return
		}

		// User is teacher, continue
		c.Next()
	}
}


// CanModifyUser checks if the authenticated user has permission to modify the target user
func CanModifyUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user ID from the URL parameter
		targetIDStr := c.Param("id")
		targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// Get the authenticated user's ID and role from the context
		authenticatedUserID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userRole, _ := c.Get("role")

		// Teachers can modify any user
		if userRole == "teacher" {
			c.Next()
			return
		}

		// Regular users can only modify themselves
		if authenticatedUserID.(uint) == uint(targetID) {
			c.Next()
			return
		}

		// User is not authorized to modify this user
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only modify your own profile"})
		c.Abort()
	}
}
