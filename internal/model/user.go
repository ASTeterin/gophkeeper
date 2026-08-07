package model

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type BatchDeleteResult struct {
	SuccessCount int
	Error        error
}

type User struct {
	UUID     uuid.UUID `json:"uuid" db:"id"`
	Login    string    `json:"login" db:"login"`
	PassHash string    `json:"pass" db:"pass_hash"`
}

type UserRepository interface {
	NextUserID() (uuid.UUID, error)
	Store(ctx context.Context, user User) error
	GetByLogin(ctx context.Context, login string) (*User, error)
}

func NewUser(id uuid.UUID, login, passHash string) User {
	return User{
		UUID:     id,
		Login:    login,
		PassHash: passHash,
	}
}
