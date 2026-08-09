package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/ASTeterin/gophkeeper/internal/model"
)

type privateDataRepository struct {
	db *sql.DB
}

func NewPrivateDataRepo(db *sql.DB) model.PrivateDataRepository {
	return &privateDataRepository{db: db}
}

func (r *privateDataRepository) Store(ctx context.Context, data *model.PrivateData) error {
	query := `INSERT INTO public.private_data (id, user_id, data_key, description, data, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, data.ID, data.UserID, data.DataKey, data.Description, data.Data)
	return err
}

func (r *privateDataRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE public.private_data SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("record not found or already deleted")
	}
	return nil
}

func (r *privateDataRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.PrivateData, error) {
	query := `SELECT id, user_id, data_key, description, data, created_at, deleted_at FROM public.private_data WHERE id = $1 AND deleted_at IS NULL`
	var p model.PrivateData
	err := r.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.UserID, &p.DataKey, &p.Description, &p.Data)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *privateDataRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.PrivateData, error) {
	query := `SELECT id, user_id, data_key, description, data, created_at, deleted_at FROM public.private_data WHERE user_id = $1 AND deleted_at IS NULL`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.PrivateData
	for rows.Next() {
		var p model.PrivateData
		if err := rows.Scan(&p.ID, &p.UserID, &p.DataKey, &p.Description, &p.Data); err != nil {
			return nil, err
		}
		result = append(result, &p)
	}
	return result, rows.Err()
}

func (r *privateDataRepository) GetByUserAndKey(ctx context.Context, userID uuid.UUID, dataKey string) (*model.PrivateData, error) {
	query := `SELECT id, user_id, data_key, description, data, created_at, deleted_at FROM public.private_data WHERE user_id = $1 AND data_key = $2 AND deleted_at IS NULL`
	var p model.PrivateData
	err := r.db.QueryRowContext(ctx, query, userID, dataKey).Scan(&p.ID, &p.UserID, &p.DataKey, &p.Description, &p.Data)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
