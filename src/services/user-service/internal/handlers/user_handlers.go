package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zahartd/social-network/src/services/user-service/internal/auth"
	"github.com/zahartd/social-network/src/services/user-service/internal/models"
	"github.com/zahartd/social-network/src/services/user-service/internal/service"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Login     string `json:"login" binding:"required"`
		Firstname string `json:"firstname" binding:"required"`
		Surname   string `json:"surname" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.service.CreateUser(req.Login, req.Firstname, req.Surname, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	login := c.Query("login")
	password := c.Query("password")
	if login == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login and password are required"})
		return
	}
	user, _, err := h.service.Login(login, password, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token, err := auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token"})
		return
	}
	token = auth.TrimBearerPrefix(token)
	if err := auth.DeleteSession(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	identifier := c.Param("identifier")
	var user *models.User
	var err error
	if strings.Contains(identifier, "-") {
		user, err = h.service.GetUserByID(identifier)
	} else {
		user, err = h.service.GetUserByLogin(identifier)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	requesterID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	if requesterID == user.ID {
		c.JSON(http.StatusOK, user)
	} else {
		summary := gin.H{
			"login":     user.Login,
			"email":     user.Email,
			"firstname": user.Firstname,
			"surname":   user.Surname,
			"bio":       user.Bio,
		}
		c.JSON(http.StatusOK, summary)
	}
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	identifier := c.Param("identifier")
	var user *models.User
	var err error
	if strings.Contains(identifier, "-") {
		user, err = h.service.GetUserByID(identifier)
	} else {
		user, err = h.service.GetUserByLogin(identifier)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	requesterID, exists := c.Get("userID")
	if !exists || requesterID != user.ID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized to update this user"})
		return
	}
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Firstname string `json:"firstname" binding:"required"`
		Surname   string `json:"surname" binding:"required"`
		Phone     string `json:"phone"`
		Bio       string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedUser, err := h.service.UpdateUser(user.ID, req.Email, req.Firstname, req.Surname, req.Phone, req.Bio, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedUser)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	identifier := c.Param("identifier")
	var user *models.User
	var err error
	if strings.Contains(identifier, "-") {
		user, err = h.service.GetUserByID(identifier)
	} else {
		user, err = h.service.GetUserByLogin(identifier)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	requesterID, exists := c.Get("userID")
	if !exists || requesterID != user.ID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized to delete this user"})
		return
	}
	err = h.service.DeleteUser(user.ID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
