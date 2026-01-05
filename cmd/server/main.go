package main

import (
	"isp-billing/internal/handler"
	"isp-billing/internal/middleware"
	"isp-billing/internal/model"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Public routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the ISP Billing Software",
		})
	})

	// Authenticated routes
	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		// Super Admin routes
		superAdmin := auth.Group("/super-admin")
		superAdmin.Use(middleware.RoleMiddleware(model.SuperAdmin))
		{
			superAdmin.GET("/dashboard", handler.SuperAdminHandler)
		}

		// Admin routes
		admin := auth.Group("/admin")
		admin.Use(middleware.RoleMiddleware(model.Admin, model.SuperAdmin))
		{
			admin.GET("/dashboard", handler.AdminHandler)
		}

		// Merchant routes
		merchant := auth.Group("/merchant")
		merchant.Use(middleware.RoleMiddleware(model.Merchant, model.Admin, model.SuperAdmin))
		{
			merchant.GET("/dashboard", handler.MerchantHandler)
		}

		// Reseller routes
		reseller := auth.Group("/reseller")
		reseller.Use(middleware.RoleMiddleware(model.Reseller, model.Admin, model.SuperAdmin))
		{
			reseller.GET("/dashboard", handler.ResellerHandler)
		}

		// Sub Reseller routes
		subReseller := auth.Group("/sub-reseller")
		subReseller.Use(middleware.RoleMiddleware(model.SubReseller, model.Reseller, model.Admin, model.SuperAdmin))
		{
			subReseller.GET("/dashboard", handler.SubResellerHandler)
		}

		// Employee routes
		employee := auth.Group("/employee")
		employee.Use(middleware.RoleMiddleware(model.Employee, model.Admin, model.SuperAdmin))
		{
			employee.GET("/dashboard", handler.EmployeeHandler)
		}

		// Client routes
		client := auth.Group("/client")
		client.Use(middleware.RoleMiddleware(model.Client, model.Employee, model.Admin, model.SuperAdmin))
		{
			client.GET("/dashboard", handler.ClientHandler)
		}
	}

	r.Run(":8080")
}