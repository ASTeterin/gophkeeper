package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ASTeterin/gophkeeper/internal/cookie"
	"github.com/ASTeterin/gophkeeper/internal/model"
	"github.com/ASTeterin/gophkeeper/internal/service"
)

type requestBody struct {
	DataKey     string `json:"data_key"`
	Description string `json:"description"`
	Data        string `json:"data"`
}

func setupDataRouter() *gin.Engine {
	mockSvc := &mockPrivateDataService{
		data: []*model.PrivateData{},
	}
	h := NewPrivateDataHandler(mockSvc)

	r := gin.New()
	r.Use(gin.Recovery())

	routerUserID := uuid.New().String()

	testAuthMiddleware := func(c *gin.Context) {
		c.Set(cookie.GetUserKey(), routerUserID)
		c.Next()
	}

	r.POST("/api/data", testAuthMiddleware, h.Store)
	r.GET("/api/data", testAuthMiddleware, h.GetAll)
	r.GET("/api/data/:key", testAuthMiddleware, h.GetByKey)
	r.DELETE("/api/data/:id", testAuthMiddleware, h.Delete)
	r.PUT("/api/data", testAuthMiddleware, h.ReplaceAll)

	return r
}

func Test_privateDataHandler_Store(t *testing.T) {
	tests := []struct {
		name   string
		body   requestBody
		expect int
	}{
		{
			name: "success",
			body: requestBody{
				DataKey: "gmail",
				Data:    "dGVzdA==",
			},
			expect: http.StatusCreated,
		},
		{
			name: "invalid json",
			body: requestBody{
				DataKey: "",
				Data:    "",
			},
			expect: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupDataRouter()
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expect, w.Code)
		})
	}
}

func Test_privateDataHandler_Store_Duplicate(t *testing.T) {
	router := setupDataRouter() // Один роутер = один пользователь
	body := requestBody{DataKey: "gmail", Data: "dGVzdA=="}
	jsonBody, _ := json.Marshal(body)

	req1 := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewReader(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	req2 := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewReader(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)
}

func Test_privateDataHandler_GetByKey(t *testing.T) {
	router := setupDataRouter()

	jsonBody, _ := json.Marshal(AddDataRequest{
		DataKey: "test_key",
		Data:    "dGF0YQ==",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewReader(jsonBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	tests := []struct {
		name   string
		key    string
		expect int
	}{
		{
			name:   "success",
			key:    "test_key",
			expect: http.StatusOK,
		},
		{
			name:   "not found",
			key:    "missing",
			expect: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/data/"+tt.key, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expect, w.Code)
		})
	}
}

func Test_privateDataHandler_GetAll(t *testing.T) {
	router := setupDataRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var result []*model.PrivateData
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.IsType(t, []*model.PrivateData{}, result)
}

func Test_privateDataHandler_Delete(t *testing.T) {
	router := setupDataRouter()

	// Создаем запись для удаления
	jsonBody, _ := json.Marshal(AddDataRequest{
		DataKey: "delete_me",
		Data:    "ZGF0YQ==",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewReader(jsonBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var created model.PrivateData
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &created))

	tests := []struct {
		name   string
		id     string
		expect int
	}{
		{
			name:   "success",
			id:     created.ID.String(),
			expect: http.StatusNoContent,
		},
		{
			name:   "invalid id",
			id:     "not-uuid",
			expect: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/data/"+tt.id, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expect, w.Code)
		})
	}
}

func Test_privateDataHandler_ReplaceAll(t *testing.T) {
	router := setupDataRouter()

	body, _ := json.Marshal(BatchDataRequest{
		Items: []service.DataItem{
			{DataKey: "key1", Data: []byte("data1")},
			{DataKey: "key2", Data: []byte("data2")},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/data", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

type mockPrivateDataService struct {
	data []*model.PrivateData
}

func (m *mockPrivateDataService) AddData(_ context.Context, userID uuid.UUID, key, desc string, data []byte) (*model.PrivateData, error) {
	for _, d := range m.data {
		if d.UserID == userID && d.DataKey == key {
			return nil, service.ErrKeyExists
		}
	}
	item := &model.PrivateData{
		ID:          uuid.New(),
		UserID:      userID,
		DataKey:     key,
		Description: desc,
		Data:        data,
	}
	m.data = append(m.data, item)
	return item, nil
}

func (m *mockPrivateDataService) GetDataByKey(_ context.Context, userID uuid.UUID, key string) (*model.PrivateData, error) {
	for _, d := range m.data {
		if d.UserID == userID && d.DataKey == key {
			return d, nil
		}
	}
	return nil, service.ErrNotFound
}

func (m *mockPrivateDataService) GetAllData(_ context.Context, userID uuid.UUID) ([]*model.PrivateData, error) {
	var result []*model.PrivateData
	for _, d := range m.data {
		if d.UserID == userID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *mockPrivateDataService) DeleteData(_ context.Context, id uuid.UUID) error {
	for i, d := range m.data {
		if d.ID == id {
			m.data = append(m.data[:i], m.data[i+1:]...)
			return nil
		}
	}
	return service.ErrNotFound
}

func (m *mockPrivateDataService) ReplaceAllData(_ context.Context, userID uuid.UUID, items []service.DataItem) error {
	var filtered []*model.PrivateData
	for _, d := range m.data {
		if d.UserID != userID {
			filtered = append(filtered, d)
		}
	}
	m.data = filtered

	for _, item := range items {
		m.data = append(m.data, &model.PrivateData{
			ID:          uuid.New(),
			UserID:      userID,
			DataKey:     item.DataKey,
			Description: item.Description,
			Data:        item.Data,
		})
	}
	return nil
}
