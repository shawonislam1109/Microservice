package api

import (
	"fmt"
	"isp-management-system/internal/db"
	"isp-management-system/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Handler holds the dependencies for the API handlers.
type Handler struct {
	Repo db.Repository
}

// NewHandler creates a new API handler.
func NewHandler(repo db.Repository) *Handler {
	return &Handler{Repo: repo}
}

// CreateUserRequest defines the expected JSON body for creating a user.
type CreateUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	PackageName string `json:"packageName"`
	ExpiryDays  int    `json:"expiryDays"`
}

// CreateUserResponse defines the JSON response after creating a user.
type CreateUserResponse struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Status     string    `json:"status"`
	ExpiryDate time.Time `json:"expiryDate"`
	Message    string    `json:"message"`
}

// CreateUser is the HTTP handler for POST /api/clients
func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// --- Validation ---
	if req.Username == "" || req.Password == "" || req.PackageName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username, password, and packageName are required",
		})
	}
	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters long",
		})
	}
	if req.ExpiryDays <= 0 {
		req.ExpiryDays = 30 // Default to 30 days
	}

	// --- Password Hashing ---
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process password",
		})
	}

	// --- Database Insertion ---
	newUser := &models.User{
		Username:   req.Username,
		Password:   string(hashedPassword),
		Status:     "active",
		ExpiryDate: time.Now().AddDate(0, 0, req.ExpiryDays),
	}

	createdUser, err := h.Repo.CreateUser(c.Context(), newUser, req.PackageName)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create user: %v", err),
		})
	}

	// --- Success Response ---
	return c.Status(fiber.StatusCreated).JSON(CreateUserResponse{
		ID:         createdUser.ID,
		Username:   createdUser.Username,
		Status:     createdUser.Status,
		ExpiryDate: createdUser.ExpiryDate,
		Message:    "User created successfully",
	})
}
