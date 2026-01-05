package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SuperAdminHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome Super Admin"})
}

func AdminHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome Admin"})
}

func MerchantHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome Merchant"})
}

func ResellerHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome Reseller	"})
}

func SubResellerHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome Sub Reseller"})
}

func EmployeeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome Employee"})
}

func ClientHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome Client"})
}
