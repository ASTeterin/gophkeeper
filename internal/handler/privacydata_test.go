package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ASTeterin/gophkeeper/internal/contracts"
	"github.com/ASTeterin/gophkeeper/internal/cookie"
	"github.com/ASTeterin/gophkeeper/internal/mock"
	"github.com/ASTeterin/gophkeeper/internal/model"
)

type requestBody struct {
	DataKey     string `json:"data_key"`
	Description string `json:"description"`
	Data        string `json:"data"`
}

func setupTest(t *testing.T) (*gin.Engine, *mock.MockPrivateDataService) {
	ctrl := gomock.NewController(t)
	mockSvc := mock.NewMockPrivateDataService(ctrl)
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
	r.DELETE("/api/data/:key", testAuthMiddleware, h.Delete)
	r.PUT("/api/data", testAuthMiddleware, h.ReplaceAll)

	return r, mockSvc
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
			name:   "invalid json",
			body:   requestBody{}, // Пустой боди или невалидный JSON
			expect: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc := setupTest(t)

			if tt.expect == http.StatusCreated {
				mockSvc.EXPECT().
					AddData(gomock.Any(), gomock.Any(), tt.body.DataKey, gomock.Any(), gomock.Any()).
					Return(&model.PrivateData{ID: uuid.New(), DataKey: tt.body.DataKey}, nil).
					Times(1)
			}

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
	router, mockSvc := setupTest(t)
	body := requestBody{DataKey: "gmail", Data: "dGVzdA=="}
	jsonBody, _ := json.Marshal(body)

	mockSvc.EXPECT().
		AddData(gomock.Any(), gomock.Any(), "gmail", gomock.Any(), gomock.Any()).
		Return(&model.PrivateData{ID: uuid.New(), DataKey: "gmail"}, nil).
		Times(1)

	req1 := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewReader(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	mockSvc.EXPECT().
		AddData(gomock.Any(), gomock.Any(), "gmail", gomock.Any(), gomock.Any()).
		Return(nil, contracts.ErrKeyExists).
		Times(1)

	req2 := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewReader(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)
}

func Test_privateDataHandler_GetByKey(t *testing.T) {
	router, mockSvc := setupTest(t)

	tests := []struct {
		name   string
		key    string
		setup  func()
		expect int
	}{
		{
			name: "success",
			key:  "test_key",
			setup: func() {
				mockSvc.EXPECT().
					GetDataByKey(gomock.Any(), gomock.Any(), "test_key").
					Return(&model.PrivateData{ID: uuid.New(), DataKey: "test_key"}, nil).
					Times(1)
			},
			expect: http.StatusOK,
		},
		{
			name: "not found",
			key:  "missing",
			setup: func() {
				mockSvc.EXPECT().
					GetDataByKey(gomock.Any(), gomock.Any(), "missing").
					Return(nil, contracts.ErrNotFound).
					Times(1)
			},
			expect: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			req := httptest.NewRequest(http.MethodGet, "/api/data/"+tt.key, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expect, w.Code)
		})
	}
}

func Test_privateDataHandler_GetAll(t *testing.T) {
	router, mockSvc := setupTest(t)

	mockSvc.EXPECT().
		GetAllData(gomock.Any(), gomock.Any()).
		Return([]*model.PrivateData{}, nil).
		Times(1)

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
	router, mockSvc := setupTest(t)

	tests := []struct {
		name   string
		key    string
		setup  func()
		expect int
	}{
		{
			name: "success",
			key:  "delete_me",
			setup: func() {
				mockSvc.EXPECT().
					DeleteData(gomock.Any(), gomock.Any(), "delete_me").
					Return(nil).
					Times(1)
			},
			expect: http.StatusNoContent,
		},
		{
			name: "not found",
			key:  "missing",
			setup: func() {
				mockSvc.EXPECT().
					DeleteData(gomock.Any(), gomock.Any(), "missing").
					Return(contracts.ErrNotFound).
					Times(1)
			},
			expect: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			req := httptest.NewRequest(http.MethodDelete, "/api/data/"+tt.key, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expect, w.Code)
		})
	}
}

func Test_privateDataHandler_ReplaceAll(t *testing.T) {
	router, mockSvc := setupTest(t)

	body, _ := json.Marshal(BatchDataRequest{
		Items: []contracts.DataItem{
			{DataKey: "key1", Data: []byte("data1")},
			{DataKey: "key2", Data: []byte("data2")},
		},
	})

	mockSvc.EXPECT().
		ReplaceAllData(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	req := httptest.NewRequest(http.MethodPut, "/api/data", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
