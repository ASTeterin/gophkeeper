package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ASTeterin/gophkeeper/internal/model"
)

func Test_privateDataService_AddData(t *testing.T) {
	ctx := context.TODO()
	userID := uuid.New()

	repo := &mockPrivateDataRepository{
		data: []*model.PrivateData{
			{
				ID:          uuid.New(),
				UserID:      userID,
				DataKey:     "existing_key",
				Description: "Existing",
				Data:        []byte("data"),
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		desc     string
		data     []byte
		expected error
	}{
		{
			name:     "positive test",
			key:      "new_key",
			desc:     "New item",
			data:     []byte("new_data"),
			expected: nil,
		},
		{
			name:     "key already exists",
			key:      "existing_key",
			desc:     "Duplicate",
			data:     []byte("data"),
			expected: ErrKeyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewPrivateDataService(repo)
			result, err := s.AddData(ctx, userID, tt.key, tt.desc, tt.data)

			if tt.expected != nil {
				assert.ErrorIs(t, err, tt.expected)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.key, result.DataKey)
			}
		})
	}
}

func Test_privateDataService_GetDataByKey(t *testing.T) {
	ctx := context.TODO()
	userID := uuid.New()
	existingID := uuid.New()

	repo := &mockPrivateDataRepository{
		data: []*model.PrivateData{
			{
				ID:          existingID,
				UserID:      userID,
				DataKey:     "test_key",
				Description: "Test",
				Data:        []byte("test_data"),
			},
		},
	}
	s := NewPrivateDataService(repo)

	tests := []struct {
		name     string
		key      string
		expected error
	}{
		{
			name:     "positive test",
			key:      "test_key",
			expected: nil,
		},
		{
			name:     "not found",
			key:      "missing_key",
			expected: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.GetDataByKey(ctx, userID, tt.key)

			if tt.expected != nil {
				assert.ErrorIs(t, err, tt.expected)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.key, result.DataKey)
			}
		})
	}
}

func Test_privateDataService_GetAllData(t *testing.T) {
	ctx := context.TODO()
	userID := uuid.New()

	repo := &mockPrivateDataRepository{
		data: []*model.PrivateData{
			{ID: uuid.New(), UserID: userID, DataKey: "key1"},
			{ID: uuid.New(), UserID: userID, DataKey: "key2"},
		},
	}
	s := NewPrivateDataService(repo)

	result, err := s.GetAllData(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func Test_privateDataService_DeleteData(t *testing.T) {
	ctx := context.TODO()
	itemID := uuid.New()

	repo := &mockPrivateDataRepository{
		data: []*model.PrivateData{
			{ID: itemID, UserID: uuid.New(), DataKey: "key"},
		},
	}
	s := NewPrivateDataService(repo)

	err := s.DeleteData(ctx, itemID)
	assert.NoError(t, err)
	assert.Len(t, repo.data, 0)
}

func Test_privateDataService_ReplaceAllData(t *testing.T) {
	ctx := context.TODO()
	userID := uuid.New()

	repo := &mockPrivateDataRepository{
		data: []*model.PrivateData{
			{ID: uuid.New(), UserID: userID, DataKey: "old_key"},
		},
	}
	s := NewPrivateDataService(repo)

	items := []DataItem{
		{DataKey: "new_1", Description: "Desc 1", Data: []byte("1")},
		{DataKey: "new_2", Description: "Desc 2", Data: []byte("2")},
	}

	err := s.ReplaceAllData(ctx, userID, items)
	assert.NoError(t, err)

	assert.Len(t, repo.data, 2)
	assert.Equal(t, "new_1", repo.data[0].DataKey)
	assert.Equal(t, "new_2", repo.data[1].DataKey)
}

type mockPrivateDataRepository struct {
	data []*model.PrivateData
}

func (m *mockPrivateDataRepository) Store(_ context.Context, data *model.PrivateData) error {
	m.data = append(m.data, data)
	return nil
}

func (m *mockPrivateDataRepository) Delete(_ context.Context, id uuid.UUID) error {
	for i, d := range m.data {
		if d.ID == id {
			m.data = append(m.data[:i], m.data[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockPrivateDataRepository) GetByID(_ context.Context, id uuid.UUID) (*model.PrivateData, error) {
	for _, d := range m.data {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockPrivateDataRepository) ListByUserID(_ context.Context, userID uuid.UUID) ([]*model.PrivateData, error) {
	var result []*model.PrivateData
	for _, d := range m.data {
		if d.UserID == userID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *mockPrivateDataRepository) GetByUserAndKey(_ context.Context, userID uuid.UUID, key string) (*model.PrivateData, error) {
	for _, d := range m.data {
		if d.UserID == userID && d.DataKey == key {
			return d, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockPrivateDataRepository) DeleteByUserID(_ context.Context, userID uuid.UUID) error {
	var result []*model.PrivateData
	for _, d := range m.data {
		if d.UserID != userID {
			result = append(result, d)
		}
	}
	m.data = result
	return nil
}

func (m *mockPrivateDataRepository) BatchStore(_ context.Context, data []*model.PrivateData) error {
	m.data = append(m.data, data...)
	return nil
}
