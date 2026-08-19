package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ASTeterin/gophkeeper/internal/contracts"
	"github.com/ASTeterin/gophkeeper/internal/model"
)

type privateDataService struct {
	repo model.PrivateDataRepository
}

func NewPrivateDataService(repo model.PrivateDataRepository) contracts.PrivateDataService {
	return &privateDataService{repo: repo}
}

func (s *privateDataService) AddData(ctx context.Context, userID uuid.UUID, key string, description string, data []byte) (*model.PrivateData, error) {
	_, err := s.repo.GetByUserAndKey(ctx, userID, key)
	if err == nil {
		return nil, contracts.ErrKeyExists
	}
	if !errors.Is(err, contracts.ErrNotFound) {
		return nil, err
	}

	newData := &model.PrivateData{
		ID:          uuid.New(),
		UserID:      userID,
		DataKey:     key,
		Description: description,
		Data:        data,
	}

	if err := s.repo.Store(ctx, newData); err != nil {
		return nil, err
	}

	return newData, nil
}

func (s *privateDataService) GetDataByKey(ctx context.Context, userID uuid.UUID, key string) (*model.PrivateData, error) {
	data, err := s.repo.GetByUserAndKey(ctx, userID, key)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *privateDataService) GetAllData(ctx context.Context, userID uuid.UUID) ([]*model.PrivateData, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *privateDataService) DeleteData(ctx context.Context, userID uuid.UUID, key string) error {
	return s.repo.DeleteByUserAndKey(ctx, userID, key)
}

func (s *privateDataService) ReplaceAllData(ctx context.Context, userID uuid.UUID, items []contracts.DataItem) error {
	var newItems []*model.PrivateData
	for _, item := range items {
		newItems = append(newItems, &model.PrivateData{
			ID:          uuid.New(),
			UserID:      userID,
			DataKey:     item.DataKey,
			Description: item.Description,
			Data:        item.Data,
		})
	}

	return s.repo.ReplaceAllData(ctx, userID, newItems)
}
