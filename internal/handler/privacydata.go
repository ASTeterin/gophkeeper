package handler

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ASTeterin/gophkeeper/internal/cookie"
	"github.com/ASTeterin/gophkeeper/internal/service"
)

type privateDataHandler struct {
	service service.PrivateDataService
}

// AddDataRequest represents the request body for storing a single private data item.
type AddDataRequest struct {
	DataKey     string `json:"data_key" binding:"required"`
	Description string `json:"description"`
	Data        string `json:"data" binding:"required"`
}

// BatchDataRequest represents the request body for replacing all user data.
type BatchDataRequest struct {
	Items []service.DataItem `json:"items" binding:"required"`
}

// PrivateDataHandler defines the HTTP handlers for managing private data.
type PrivateDataHandler interface {
	Store(c *gin.Context)
	GetByKey(c *gin.Context)
	GetAll(c *gin.Context)
	Delete(c *gin.Context)
	ReplaceAll(c *gin.Context)
}

// NewPrivateDataHandler creates a new instance of the private data handler.
func NewPrivateDataHandler(svc service.PrivateDataService) PrivateDataHandler {
	return &privateDataHandler{service: svc}
}

// Store adds a new private data item for the authenticated user.
// Returns 201 Created on success, 409 Conflict if the key already exists, 400/401/500 on error.
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

// GetByKey retrieves a specific private data item by its key.
// Returns 200 OK with the data, 404 Not Found if missing, 400/401/500 on error.
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

// GetAll retrieves all private data items for the authenticated user.
// Returns 200 OK with a list of items, 401/500 on error.
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

// Delete removes a private data item by its ID.
// Returns 204 No Content on success, 400/500 on error.
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

// ReplaceAll deletes all existing data for the user and saves a new batch.
// Returns 200 OK on success, 400/401/500 on error.
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

// isValidBase64 checks if a string is valid base64.
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
