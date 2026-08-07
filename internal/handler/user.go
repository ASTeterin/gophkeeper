package handler

import (
	"context"
	"errors"
	"github.com/ASTeterin/gothkeeper/internal/cookie"
	"github.com/ASTeterin/gothkeeper/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type handler struct {
	userService service.UserService
}

type User struct {
	Login    string `json:"login" binding:"required"`
	PassHash string `json:"password" binding:"required"`
}

type Handler interface {
	Register(ctx context.Context, c *gin.Context)
	Authenticate(ctx context.Context, c *gin.Context)
}

func NewHandler(userService service.UserService) Handler {
	return &handler{
		userService: userService,
	}
}

func (h *handler) Register(ctx context.Context, c *gin.Context) {
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

func (h *handler) Authenticate(ctx context.Context, c *gin.Context) {
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

func getUserID(c *gin.Context) string {
	return c.GetString(cookie.GetUserKey())
}

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "userID", userID)
}
