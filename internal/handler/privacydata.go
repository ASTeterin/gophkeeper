package handler

import (
	"encoding/base64"
	"errors"
	"github.com/ASTeterin/gophkeeper/internal/cookie"
	"github.com/ASTeterin/gophkeeper/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type privateDataHandler struct {
	service service.PrivateDataService
}

type AddDataRequest struct {
	DataKey     string `json:"data_key" binding:"required"`
	Description string `json:"description"`
	Data        string `json:"data" binding:"required"`
}

type BatchDataRequest struct {
	Items []service.DataItem `json:"items" binding:"required"`
}

type PrivateDataHandler interface {
	Store(c *gin.Context)
	GetByKey(c *gin.Context)
	GetAll(c *gin.Context)
	Delete(c *gin.Context)
	ReplaceAll(c *gin.Context)
}

func NewPrivateDataHandler(svc service.PrivateDataService) PrivateDataHandler {
	return &privateDataHandler{service: svc}
}

func (h *privateDataHandler) Store(c *gin.Context) {
	userIDStr := c.GetString(cookie.GetUserKey())
	if userIDStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req AddDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Декодируем данные, если они в base64, иначе используем как есть
	var data []byte
	if isValidBase64(req.Data) {
		data, err = base64.StdEncoding.DecodeString(req.Data)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid base64 data"})
			return
		}
	} else {
		data = []byte(req.Data)
	}

	result, err := h.service.AddData(c.Request.Context(), userID, req.DataKey, req.Description, data)
	if err != nil {
		if errors.Is(err, service.ErrKeyExists) {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "key already exists"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *privateDataHandler) GetByKey(c *gin.Context) {
	userIDStr := c.GetString(cookie.GetUserKey())
	if userIDStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "key parameter is required"})
		return
	}

	result, err := h.service.GetDataByKey(c.Request.Context(), userID, key)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *privateDataHandler) GetAll(c *gin.Context) {
	userIDStr := c.GetString(cookie.GetUserKey())
	if userIDStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	result, err := h.service.GetAllData(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *privateDataHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteData(c.Request.Context(), id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *privateDataHandler) ReplaceAll(c *gin.Context) {
	userIDStr := c.GetString(cookie.GetUserKey())
	if userIDStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req BatchDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.ReplaceAllData(c.Request.Context(), userID, req.Items); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusOK)
}

// isValidBase64 проверяет, является ли строка валидным base64
func isValidBase64(s string) bool {
	if len(s) == 0 {
		return false
	}
	if len(s)%4 != 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
