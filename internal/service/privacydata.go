package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"

	"github.com/ASTeterin/gophkeeper/internal/model"
)

var (
	ErrKeyExists = errors.New("data key already exists for this user")
	ErrNotFound  = errors.New("record not found")
)

type DataItem struct {
	DataKey     string `json:"data_key"`
	Description string `json:"description"`
	Data        []byte `json:"data"`
}

type PrivateDataService interface {
	AddData(ctx context.Context, userID uuid.UUID, key string, description string, data []byte) (*model.PrivateData, error)
	GetDataByKey(ctx context.Context, userID uuid.UUID, key string) (*model.PrivateData, error)
	GetAllData(ctx context.Context, userID uuid.UUID) ([]*model.PrivateData, error)
	DeleteData(ctx context.Context, id uuid.UUID) error
	ReplaceAllData(ctx context.Context, userID uuid.UUID, items []DataItem) error
}

type PrivateDataServiceImpl struct {
	repo model.PrivateDataRepository
}

func NewPrivateDataService(repo model.PrivateDataRepository) *PrivateDataServiceImpl {
	return &PrivateDataServiceImpl{repo: repo}
}

func (s *PrivateDataServiceImpl) AddData(ctx context.Context, userID uuid.UUID, key string, description string, data []byte) (*model.PrivateData, error) {
	_, err := s.repo.GetByUserAndKey(ctx, userID, key)
	if err == nil {
		return nil, ErrKeyExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
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

func (s *PrivateDataServiceImpl) GetDataByKey(ctx context.Context, userID uuid.UUID, key string) (*model.PrivateData, error) {
	data, err := s.repo.GetByUserAndKey(ctx, userID, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

func (s *PrivateDataServiceImpl) GetAllData(ctx context.Context, userID uuid.UUID) ([]*model.PrivateData, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *PrivateDataServiceImpl) DeleteData(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *PrivateDataServiceImpl) ReplaceAllData(ctx context.Context, userID uuid.UUID, items []DataItem) error {
	// Используем транзакцию на уровне сервиса, если репозиторий не поддерживает транзакции напрямую
	// Для простоты здесь последовательные вызовы, но лучше обернуть в tx

	// 1. Удаляем старые
	if err := s.repo.DeleteByUserID(ctx, userID); err != nil {
		return err
	}

	// 2. Формируем новые
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

	// 3. Сохраняем новые
	return s.repo.BatchStore(ctx, newItems)
}
