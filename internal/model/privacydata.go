package model

import (
	"context"

	"github.com/google/uuid"
)

type PrivateData struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	DataKey     string    `json:"data_key"`
	Description string    `json:"description"`
	Data        []byte    `json:"data"`
}

type PrivateDataRepository interface {
	Store(ctx context.Context, data *PrivateData) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*PrivateData, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*PrivateData, error)
	GetByUserAndKey(ctx context.Context, userID uuid.UUID, dataKey string) (*PrivateData, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	BatchStore(ctx context.Context, data []*PrivateData) error
}
