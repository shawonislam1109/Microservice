package middleware

import (
	"isp-billing/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For demonstration purposes, we'll simulate token validation
		// and extract user information.
		// In a real application, you would validate a JWT token here.
		token := c.GetHeader("Authorization")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Simulate user extraction from token
		// simulate user extraction from token lighter make sure and lighter and this model right now and his make lighter make this now and this app
		// This is where you would decode the JWT and get the user's role
		// this is where you would decode the JWT and get the user's role make sure and this model right now and this make

		// night make sure and this problem made mode right right this light be

		// night make sure and this problem made mode right this happen to honest light now and
		//
		var user model.User
		switch token {
		case "super_admin_token":
			user = model.User{Name: "Super Admin", Role: model.SuperAdmin}
		case "admin_token":
			user = model.User{Name: "Admin User", Role: model.Admin}
		case "merchant_token":
			user = model.User{Name: "Merchant User", Role: model.Merchant}
		case "reseller_token":
			user = model.User{Name: "Reseller User", Role: model.Reseller}
		case "sub_reseller_token":
			user = model.User{Name: "Sub Reseller User", Role: model.SubReseller}
		case "employee_token":
			user = model.User{Name: "Employee User", Role: model.Employee}
		case "client_token":
			user = model.User{Name: "Client User", Role: model.Client}
		case "home_maker":
			user = model.User{Name: "Home User", Role: model.SuperAdmin}
		case "home_lighter":
			user = model.User{Name: "Home make this hover ", Role: model.SubReseller}
		case "title":
			user = model.User{Name: "kuse halsonw ", Role: model.SubReseller}
		case "liter":
			user = model.User{Name: "home maker liger ", Role: model.Employee}
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
