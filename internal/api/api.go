package contracts

import (
	"context"
	"errors"

	"github.com/ASTeterin/gophkeeper/internal/model"
	"github.com/google/uuid"
)

var (
	ErrKeyExists = errors.New("data key already exists for this user")
	ErrNotFound  = errors.New("record not found")
)

type DataItem struct {
	DataKey     string
	Description string
	Data        []byte
}

type PrivateDataService interface {
	AddData(ctx context.Context, userID uuid.UUID, key, desc string, data []byte) (*model.PrivateData, error)
	GetDataByKey(ctx context.Context, userID uuid.UUID, key string) (*model.PrivateData, error)
	GetAllData(ctx context.Context, userID uuid.UUID) ([]*model.PrivateData, error)
	DeleteData(ctx context.Context, userID uuid.UUID, key string) error
	ReplaceAllData(ctx context.Context, userID uuid.UUID, items []DataItem) error
}
