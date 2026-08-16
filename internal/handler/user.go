package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ASTeterin/gophkeeper/internal/cookie"
	"github.com/ASTeterin/gophkeeper/internal/service"
)

type handler struct {
	userService service.UserService
}

type User struct {
	Login    string `json:"login" binding:"required"`
	PassHash string `json:"password" binding:"required"`
}

type UserHandler interface {
	Register(c *gin.Context)
	Authenticate(c *gin.Context)
}

func NewUserHandler(userService service.UserService) UserHandler {
	return &handler{
		userService: userService,
	}
}

func (h *handler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	body := User{}
	err := c.ShouldBindBodyWithJSON(&body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	userID, err := h.userService.Register(ctx, body.Login, body.PassHash)

	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			c.AbortWithStatus(http.StatusConflict)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Set(cookie.GetUserKey(), *userID)
	c.Status(http.StatusOK)
}

func (h *handler) Authenticate(c *gin.Context) {
	ctx := c.Request.Context()
	body := User{}
	err := c.ShouldBindBodyWithJSON(&body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	userID, err := h.userService.Authenticate(ctx, body.Login, body.PassHash)

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set(cookie.GetUserKey(), *userID)
	c.Status(http.StatusOK)
}
