package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ASTeterin/gophkeeper/internal/model"
)

const (
	userLogin    = "user_login"
	userPassword = "user_password"
	hash         = "$2a$10$F1.UPmsh2bwaHH427GJ0A.Qmr3dWdDdLFhut/LiPaqAazvatOLyiK"
)

func Test_userService_Register(t *testing.T) {
	repo := &mockUserRepository{
		user: []model.User{
			{
				UUID:     uuid.New(),
				Login:    "user",
				PassHash: "",
			},
		},
	}
	tests := []struct {
		name     string
		login    string
		password string
		error    error
	}{
		{
			name:     "positive test",
			login:    "login",
			password: "password",
			error:    nil,
		},
		{
			name:     "user already exists",
			login:    "user",
			password: "password",
			error:    ErrUserExists,
		},
	}
	for _, tt := range tests {
		ctx := context.TODO()
		t.Run(tt.name, func(t *testing.T) {
			s := &userService{
				repo: repo,
			}
			_, err := s.Register(ctx, tt.login, tt.password)
			if tt.error != nil {
				assert.ErrorIs(t, err, tt.error)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_userService_Authenticate(t *testing.T) {
	ctx := context.TODO()
	repo := &mockUserRepository{
		user: []model.User{
			{
				UUID:     uuid.New(),
				Login:    userLogin,
				PassHash: hash,
			},
		},
	}
	s := &userService{
		repo: repo,
	}
	tests := []struct {
		name     string
		login    string
		password string
		error    error
	}{
		{
			name:     "positive test",
			login:    userLogin,
			password: userPassword,
			error:    nil,
		},
		{
			name:     "user not authenticate",
			login:    userLogin,
			password: "password",
			error:    ErrUserNotAuthenticate,
		},
	}
	for _, tt := range tests {
		_, err := s.Authenticate(ctx, tt.login, tt.password)
		t.Run(tt.name, func(t *testing.T) {
			if tt.error != nil {
				assert.ErrorIs(t, err, tt.error)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type mockUserRepository struct {
	user []model.User
}

func (m mockUserRepository) NextUserID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m mockUserRepository) Store(_ context.Context, user model.User) error {
	m.user = append(m.user, user)
	return nil
}

func (m mockUserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	for _, user := range m.user {
		if user.Login == login {
			return &user, nil
		}
	}
	return nil, model.ErrUserNotFound
}
