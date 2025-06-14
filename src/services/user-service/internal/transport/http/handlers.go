package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/user-service/internal/domain/service"
	"github.com/zahartd/social-network/src/services/user-service/internal/infrastructure/auth"
	"github.com/zahartd/social-network/src/services/user-service/internal/utils"
)

func AttachRoutes(r *gin.Engine, us *service.Service) {
	h := handler{us}
	r.POST("/user", h.createUser)
	r.GET("/user/login", h.login)
	r.GET("/user/logout", h.logout)

	g := r.Group("/user", AuthMiddleware())
	g.GET(":identifier", h.getUser)
	g.PUT(":identifier", h.updateUser)
	g.DELETE(":identifier", h.deleteUser)
}

type handler struct{ svc *service.Service }

func (h handler) createUser(c *gin.Context) {
	var req struct {
		Login     string `json:"login" binding:"required,login"`
		Firstname string `json:"firstname" binding:"required"`
		Surname   string `json:"surname" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, tok, err := h.svc.Create(req.Login, req.Firstname, req.Surname, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": u, "token": tok})
}

func (h handler) login(c *gin.Context) {
	login := c.Query("login")
	pass := c.Query("password")
	if !utils.ValidateLogin(login) || !utils.ValidatePassword(pass) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}
	tok, err := h.svc.Login(c.ClientIP(), login, pass)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tok})
}

func (h handler) logout(c *gin.Context) {
	tok := auth.TrimBearer(c.GetHeader("Authorization"))
	if err := h.svc.Logout(tok); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bye"})
}

func (h handler) getUser(c *gin.Context) {
	idStr := c.Param("identifier")
	id, isUUID, err := utils.ParseIdentifier(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var u *models.User
	if isUUID {
		u, err = h.svc.GetByID(id)
	} else {
		u, err = h.svc.GetByLogin(id)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if reqID, _ := c.Get("userID"); reqID == u.ID {
		c.JSON(http.StatusOK, u)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"login":     u.Login,
			"email":     u.Email,
			"firstname": u.Firstname,
			"surname":   u.Surname,
			"bio":       u.Bio,
		})
	}
}

func (h handler) updateUser(c *gin.Context) {
	ident := c.Param("identifier")
	id, isUUID, err := utils.ParseIdentifier(ident)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isUUID {
		u, err := h.svc.GetByLogin(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		id = u.ID
	}

	if c.GetString("userID") != id {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Email     string `json:"email"     binding:"required,email"`
		Firstname string `json:"firstname" binding:"required"`
		Surname   string `json:"surname"   binding:"required"`
		Phone     string `json:"phone"     binding:"omitempty,phone"`
		Bio       string `json:"bio"       binding:"omitempty,max=1400"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := &models.User{
		Email:     req.Email,
		Firstname: req.Firstname,
		Surname:   req.Surname,
		Phone:     req.Phone,
		Bio:       req.Bio,
	}
	updated, err := h.svc.Update(id, u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h handler) deleteUser(c *gin.Context) {
	ident := c.Param("identifier")
	id, isUUID, err := utils.ParseIdentifier(ident)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isUUID {
		u, err := h.svc.GetByLogin(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		id = u.ID
	}

	if reqID := c.GetString("userID"); reqID != id {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	token := c.GetString("token")
	if err := h.svc.Delete(id, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
