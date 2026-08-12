package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/ASTeterin/gophkeeper/internal/model"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) model.UserRepository {
	repo := &userRepo{
		db: db,
	}
	return repo
}

func (repo *userRepo) NextUserID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (repo *userRepo) Store(ctx context.Context, user model.User) error {
	const query = `
        INSERT INTO users (id, login, pass_hash)
        VALUES ($1, $2, $3)
        ON CONFLICT (id) DO NOTHING
    `

	_, err := repo.db.ExecContext(ctx, query, user.UUID.String(), user.Login, user.PassHash)
	return err
}

func (repo *userRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `SELECT id, login, pass_hash FROM users WHERE login = $1`
	user := model.User{}
	err := repo.db.QueryRowContext(ctx, query, login).Scan(
		&user.UUID, &user.Login, &user.PassHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrUserNotFound
	}

	return &user, err
}
