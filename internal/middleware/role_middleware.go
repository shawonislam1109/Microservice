package middleware

import (
	"isp-billing/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(roles ...model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		userModel, ok := user.(model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context "})
			c.Abort()
			return
		}

		hasPermission := false
		for _, role := range roles {
			if userModel.Role == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this resource"})
			c.Abort()
			return
		}

		c.Next()
	}
}
