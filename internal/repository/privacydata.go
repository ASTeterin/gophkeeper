package repository

import (
	"context"
	"database/sql"
	"errors"
	contracts "github.com/ASTeterin/gophkeeper/internal/api"
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
	query := `INSERT INTO public.private_data (id, user_id, data_key, description, data) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, data.ID, data.UserID, data.DataKey, data.Description, data.Data)
	return err
}

func (r *privateDataRepository) DeleteByUserAndKey(ctx context.Context, userID uuid.UUID, dataKey string) error {
	query := `DELETE FROM public.private_data WHERE user_id = $1 AND data_key = $2`
	result, err := r.db.ExecContext(ctx, query, userID, dataKey)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return contracts.ErrNotFound
	}
	return nil
}

func (r *privateDataRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.PrivateData, error) {
	query := `SELECT id, user_id, data_key, description, data FROM public.private_data WHERE user_id = $1`
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
	query := `SELECT id, user_id, data_key, description, data FROM public.private_data WHERE user_id = $1 AND data_key = $2`
	var p model.PrivateData
	err := r.db.QueryRowContext(ctx, query, userID, dataKey).Scan(&p.ID, &p.UserID, &p.DataKey, &p.Description, &p.Data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, contracts.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *privateDataRepository) ReplaceAllData(ctx context.Context, userID uuid.UUID, items []*model.PrivateData) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM public.private_data WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	if len(items) > 0 {
		stmt, err := tx.PrepareContext(ctx, `INSERT INTO public.private_data (id, user_id, data_key, description, data) VALUES ($1, $2, $3, $4, $5)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, item := range items {
			_, err := stmt.ExecContext(ctx, item.ID, item.UserID, item.DataKey, item.Description, item.Data)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
