package service

import (
	"context"
	"errors"
	contracts "github.com/ASTeterin/gophkeeper/internal/api"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/ASTeterin/gophkeeper/internal/mock"
	"github.com/ASTeterin/gophkeeper/internal/model"
)

func Test_privateDataService_AddData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	userID := uuid.New()

	tests := []struct {
		name     string
		key      string
		desc     string
		data     []byte
		setup    func(m *mock.MockPrivateDataRepository)
		expected error
	}{
		{
			name: "positive test",
			key:  "new_key",
			desc: "New item",
			data: []byte("new_data"),
			setup: func(m *mock.MockPrivateDataRepository) {
				m.EXPECT().
					GetByUserAndKey(gomock.Any(), userID, "new_key").
					Return(nil, contracts.ErrNotFound).
					Times(1)
				m.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expected: nil,
		},
		{
			name: "key already exists",
			key:  "existing_key",
			desc: "Duplicate",
			data: []byte("data"),
			setup: func(m *mock.MockPrivateDataRepository) {
				existingItem := &model.PrivateData{
					ID:          uuid.New(),
					UserID:      userID,
					DataKey:     "existing_key",
					Description: "Existing",
					Data:        []byte("data"),
				}
				m.EXPECT().
					GetByUserAndKey(gomock.Any(), userID, "existing_key").
					Return(existingItem, nil).
					Times(1)
			},
			expected: contracts.ErrKeyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock.NewMockPrivateDataRepository(ctrl)
			tt.setup(mockRepo)

			s := NewPrivateDataService(mockRepo)
			result, err := s.AddData(ctx, userID, tt.key, tt.desc, tt.data)

			if tt.expected != nil {
				assert.ErrorIs(t, err, tt.expected)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if result != nil {
					assert.Equal(t, tt.key, result.DataKey)
				}
			}
		})
	}
}

func Test_privateDataService_GetDataByKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	userID := uuid.New()
	existingID := uuid.New()

	tests := []struct {
		name     string
		key      string
		setup    func(m *mock.MockPrivateDataRepository)
		expected error
	}{
		{
			name: "positive test",
			key:  "test_key",
			setup: func(m *mock.MockPrivateDataRepository) {
				item := &model.PrivateData{
					ID:          existingID,
					UserID:      userID,
					DataKey:     "test_key",
					Description: "Test",
					Data:        []byte("test_data"),
				}
				m.EXPECT().
					GetByUserAndKey(gomock.Any(), userID, "test_key").
					Return(item, nil).
					Times(1)
			},
			expected: nil,
		},
		{
			name: "not found",
			key:  "missing_key",
			setup: func(m *mock.MockPrivateDataRepository) {
				m.EXPECT().
					GetByUserAndKey(gomock.Any(), userID, "missing_key").
					Return(nil, contracts.ErrNotFound).
					Times(1)
			},
			expected: contracts.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock.NewMockPrivateDataRepository(ctrl)
			tt.setup(mockRepo)

			s := NewPrivateDataService(mockRepo)
			result, err := s.GetDataByKey(ctx, userID, tt.key)

			if tt.expected != nil {
				assert.ErrorIs(t, err, tt.expected)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if result != nil {
					assert.Equal(t, tt.key, result.DataKey)
				}
			}
		})
	}
}

func Test_privateDataService_GetAllData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	userID := uuid.New()

	mockRepo := mock.NewMockPrivateDataRepository(ctrl)

	expectedItems := []*model.PrivateData{
		{ID: uuid.New(), UserID: userID, DataKey: "key1"},
		{ID: uuid.New(), UserID: userID, DataKey: "key2"},
	}

	mockRepo.EXPECT().
		ListByUserID(gomock.Any(), userID).
		Return(expectedItems, nil).
		Times(1)

	s := NewPrivateDataService(mockRepo)
	result, err := s.GetAllData(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedItems, result)
}

func Test_privateDataService_DeleteData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	userID := uuid.New()
	key := "test_key"

	tests := []struct {
		name     string
		setup    func(m *mock.MockPrivateDataRepository)
		expected error
	}{
		{
			name: "success",
			setup: func(m *mock.MockPrivateDataRepository) {
				m.EXPECT().
					DeleteByUserAndKey(gomock.Any(), userID, key).
					Return(nil).
					Times(1)
			},
			expected: nil,
		},
		{
			name: "not found",
			setup: func(m *mock.MockPrivateDataRepository) {
				m.EXPECT().
					DeleteByUserAndKey(gomock.Any(), userID, key).
					Return(errors.New("record not found")).
					Times(1)
			},
			expected: errors.New("record not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock.NewMockPrivateDataRepository(ctrl)
			tt.setup(mockRepo)

			s := NewPrivateDataService(mockRepo)
			err := s.DeleteData(ctx, userID, key)
			if tt.expected != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expected.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_privateDataService_ReplaceAllData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	userID := uuid.New()

	mockRepo := mock.NewMockPrivateDataRepository(ctrl)

	items := []contracts.DataItem{
		{DataKey: "new_1", Description: "Desc 1", Data: []byte("1")},
		{DataKey: "new_2", Description: "Desc 2", Data: []byte("2")},
	}

	expectedItems := []*model.PrivateData{
		{DataKey: "new_1", Description: "Desc 1", Data: []byte("1")},
		{DataKey: "new_2", Description: "Desc 2", Data: []byte("2")},
	}

	for i := range expectedItems {
		expectedItems[i].ID = uuid.New()
		expectedItems[i].UserID = userID
	}

	mockRepo.EXPECT().
		ReplaceAllData(gomock.Any(), userID, gomock.Any()).
		Return(nil).
		Times(1)

	s := NewPrivateDataService(mockRepo)
	err := s.ReplaceAllData(ctx, userID, items)

	assert.NoError(t, err)
}
